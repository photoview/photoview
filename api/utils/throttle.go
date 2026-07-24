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
	t.mu.Lock()
	defer t.mu.Unlock()
	if time.Now().After(t.lastAction.Add(t.interval)) {
		t.lastAction = time.Now()
		action()
	}
}
