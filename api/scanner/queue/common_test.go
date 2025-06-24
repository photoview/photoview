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

	job.done = true
}

func (p *mockProcessor) periodicTrigger(ctx context.Context) {
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
		if !job.done {
			t.Errorf("job(%d) is not done", job.id)
		}
	}
}

func TestCommonQueueBackgroundPeriod(t *testing.T) {
	processer := &mockProcessor{}
	interval := time.Second / 10

	queue, err := newCommonQueue(t.Context(), interval, 3, processer)
	if err != nil {
		t.Fatalf("create common queue error: %v", err)
	}
	queue.RunBackground()
	defer queue.Close()

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
		t.Errorf("processer.lastTriggered is not triggered after waiting 2 periods")
	}
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
		if !job.done {
			t.Errorf("job(%d) is not done", job.id)
		}
	}
}
