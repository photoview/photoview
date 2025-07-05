package queue

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"testing"
	"time"
)

func init() {
	// Avoid panic with providing flags in `test_utils/integration_setup.go`.
	flag.CommandLine.Init("queue", flag.ContinueOnError)
}

type mockJob struct {
	id            int
	sleepDuration time.Duration
	done          bool
	mu            sync.Mutex
}

func (j *mockJob) checkDone() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.done
}

func (j *mockJob) setDone() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.done = true
}

func (j *mockJob) Key() int {
	return j.id
}

func (j *mockJob) String() string {
	return fmt.Sprintf("job(%d)", j.id)
}

type mockProcessor struct {
	lastTriggerMu sync.Mutex
	lastTriggered time.Time
}

func (p *mockProcessor) processJob(ctx context.Context, job *mockJob) {
	time.Sleep(job.sleepDuration)

	job.setDone()
}

func (p *mockProcessor) fillPeriodicJobs(ctx context.Context) {
	p.lastTriggerMu.Lock()
	defer p.lastTriggerMu.Unlock()

	p.lastTriggered = time.Now()
}

func TestCommonQueueParallel(t *testing.T) {
	processer := &mockProcessor{}

	queue, err := newCommonQueue(t.Context(), time.Second, 3, processer)
	if err != nil {
		t.Fatalf("create common queue error: %v", err)
	}

	jobDuration := time.Second / 10
	jobs := []*mockJob{
		&mockJob{id: 1, sleepDuration: jobDuration},
		&mockJob{id: 2, sleepDuration: jobDuration},
		&mockJob{id: 3, sleepDuration: jobDuration},
	}
	queue.appendBacklog(jobs)

	start := time.Now()
	queue.ConsumeAllBacklog(t.Context())
	queue.Close() // ensure all jobs are done
	duration := time.Now().Sub(start)

	if got, want := duration, jobDuration; (got - want).Abs() >= (jobDuration / 10) {
		t.Errorf("queue.ConsumeAllBacklog() took %v to finish, which should be around %v since jobs run in parallel", got, want)
	}

	for _, job := range jobs {
		if !job.checkDone() {
			t.Errorf("job(%d) is not done", job.id)
		}
	}
}

func TestCommonQueueBackgroundPeriod(t *testing.T) {
	processer := &mockProcessor{}
	interval := time.Second / 10

	queue, err := newCommonQueue(t.Context(), 0, 3, processer)
	if err != nil {
		t.Fatalf("create common queue error: %v", err)
	}
	queue.RunBackground()
	defer queue.Close()

	t.Run("StopTickerFromBeginning", func(t *testing.T) {
		time.Sleep(interval * 2)

		processer.lastTriggerMu.Lock()
		noTicker := processer.lastTriggered
		processer.lastTriggerMu.Unlock()

		if !noTicker.IsZero() {
			t.Errorf("processer.lastTriggered is not zero when trigger period is 0")
		}
	})

	t.Run("CheckTicker", func(t *testing.T) {
		queue.UpdateScanInterval(interval)
		time.Sleep(interval * 2)

		processer.lastTriggerMu.Lock()
		firstTicker := processer.lastTriggered
		processer.lastTriggerMu.Unlock()

		if firstTicker.IsZero() {
			t.Errorf("processer.lastTriggered is zero after waiting 2 periods")
		}

		time.Sleep(interval * 2)
		processer.lastTriggerMu.Lock()
		secondTicker := processer.lastTriggered
		processer.lastTriggerMu.Unlock()

		if !secondTicker.After(firstTicker) {
			t.Errorf("processer.lastTriggered is not triggered after waiting 2 periods, first: %v, second: %v", firstTicker, secondTicker)
		}
	})

	t.Run("StopTickerAfterWhile", func(t *testing.T) {
		if err := queue.UpdateScanInterval(0); err != nil {
			t.Fatalf("queue.UpdateScanInterval(0) returns error: %v", err)
		}

		processer.lastTriggerMu.Lock()
		begin := processer.lastTriggered
		processer.lastTriggerMu.Unlock()

		time.Sleep(interval * 2)

		processer.lastTriggerMu.Lock()
		end := processer.lastTriggered
		processer.lastTriggerMu.Unlock()

		if !begin.Equal(end) {
			t.Errorf("processer.lastTriggered is changed when trigger period is 0, begin: %v, end: %v", begin, end)
		}
	})
}

func TestCommonQueueBackgroundJobs(t *testing.T) {
	processer := &mockProcessor{}
	interval := time.Second / 10

	queue, err := newCommonQueue(t.Context(), interval, 3, processer)
	if err != nil {
		t.Fatalf("create common queue error: %v", err)
	}
	queue.RunBackground()

	jobDuration := time.Second / 10
	jobs := []*mockJob{
		&mockJob{id: 1, sleepDuration: jobDuration},
		&mockJob{id: 2, sleepDuration: jobDuration},
		&mockJob{id: 3, sleepDuration: jobDuration},
	}
	queue.appendBacklog(jobs)

	time.Sleep(jobDuration * 2)

	queue.Close()

	for _, job := range jobs {
		if !job.checkDone() {
			t.Errorf("job(%d) is not done", job.id)
		}
	}
}

func TestCommonQueueInvalidCase(t *testing.T) {
	t.Run("CreateWithInvalidInterval", func(t *testing.T) {
		processer := &mockProcessor{}
		_, err := newCommonQueue(t.Context(), -1*time.Second, 2, processer)
		if err == nil {
			t.Fatalf("should not create common queue, but no error happens")
		}
	})

	t.Run("CreateWithInvalidWorkerNumber", func(t *testing.T) {
		processer := &mockProcessor{}
		_, err := newCommonQueue(t.Context(), 0, -1, processer)
		if err == nil {
			t.Fatalf("should not create common queue, but no error happens")
		}
	})

	t.Run("UpdateWithInvalidInterval", func(t *testing.T) {
		processer := &mockProcessor{}
		queue, err := newCommonQueue(t.Context(), 0, 2, processer)
		if err != nil {
			t.Fatalf("create common queue error: %v", err)
		}
		defer queue.Close()

		if err := queue.UpdateScanInterval(-1 * time.Second); err == nil {
			t.Fatalf("queue.UpdateScanInterval(-1*time.Second) should returns error, which is not")
		}
	})

	t.Run("UpdateWithInvalidWorkerNumber", func(t *testing.T) {
		processer := &mockProcessor{}
		queue, err := newCommonQueue(t.Context(), 0, 2, processer)
		if err != nil {
			t.Fatalf("create common queue error: %v", err)
		}
		defer queue.Close()

		if err := queue.RescaleWorkers(-1); err == nil {
			t.Fatalf("queue.RescaleWorkers(-1) should returns error, which is not")
		}
	})
}

func TestCommonQueueRescaleWorkers(t *testing.T) {
	processer := &mockProcessor{}
	interval := time.Second / 10

	queue, err := newCommonQueue(t.Context(), interval, 0, processer)
	if err != nil {
		t.Fatalf("create common queue error: %v", err)
	}
	queue.RunBackground()
	defer queue.Close()

	t.Log("len(worker) = 0")

	jobDuration := time.Second / 10
	jobs := []*mockJob{
		&mockJob{id: 1, sleepDuration: jobDuration},
		&mockJob{id: 2, sleepDuration: jobDuration},
		&mockJob{id: 3, sleepDuration: jobDuration},
	}
	queue.appendBacklog(jobs)

	time.Sleep(jobDuration * 2)

	for _, job := range jobs {
		if job.checkDone() {
			t.Errorf("job(%d) should not be done when len(worker) = 0", job.id)
		}
	}

	if err := queue.RescaleWorkers(10); err != nil {
		t.Fatalf("queue.RescaleWorkers(10) returns error: %v", err)
	}
	t.Log("len(worker) = 10")

	time.Sleep(jobDuration * 2)

	for _, job := range jobs {
		if !job.checkDone() {
			t.Errorf("job(%d) should be done when len(worker) = 10", job.id)
		}
	}

	if err := queue.RescaleWorkers(1); err != nil {
		t.Fatalf("queue.RescaleWorkers(1) returns error: %v", err)
	}
	t.Log("len(worker) = 1")

	jobs = []*mockJob{
		&mockJob{id: 1, sleepDuration: jobDuration},
		&mockJob{id: 2, sleepDuration: jobDuration},
		&mockJob{id: 3, sleepDuration: jobDuration},
	}
	queue.appendBacklog(jobs)

	time.Sleep(jobDuration * 2)

	allDone := true
	for _, job := range jobs {
		if !job.checkDone() {
			allDone = false
		}
	}
	if allDone {
		t.Fatalf("jobs are all done when len(worker) = 1, which should not")
	}
}
