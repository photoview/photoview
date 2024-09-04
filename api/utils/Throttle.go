package utils

import "time"

type Throttle struct {
	interval   time.Duration
	lastAction time.Time
}

func NewThrottle(interval time.Duration) Throttle {
	return Throttle{
		interval:   interval,
		lastAction: time.Unix(0, 0),
	}
}

func (t *Throttle) Trigger(action func()) {
	if action == nil {
		return
	}
	if time.Now().After(t.lastAction.Add(t.interval)) {
		t.lastAction = time.Now()
		action()
	}
}
