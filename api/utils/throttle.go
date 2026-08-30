package utils

import (
	"sync"
	"time"
)

type Throttle struct {
	mu         sync.Mutex
	interval   time.Duration
	lastAction time.Time
}

func NewThrottle(interval time.Duration) *Throttle {
	return &Throttle{
		interval:   interval,
		lastAction: time.Unix(0, 0),
	}
}

func (t *Throttle) Trigger(action func()) {
	if action == nil {
		return
	}

	run := false
	t.mu.Lock()
	if time.Now().After(t.lastAction.Add(t.interval)) {
		t.lastAction = time.Now()
		run = true
	}
	t.mu.Unlock()

	if run {
		action()
	}
}
