package clock

import (
	"sync"
	"time"
)

type Clock interface {
	Now() time.Time
}

type clock struct{}

func (clock) Now() time.Time {
	return time.Now()
}

func New() Clock {
	return &clock{}
}

type mockClock struct {
	mu  sync.Mutex
	now time.Time
}

func (m *mockClock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

func Mock() Clock {
	return &mockClock{now: time.Unix(0, 0)}
}
