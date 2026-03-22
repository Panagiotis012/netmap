package scanner

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestSchedulerRunsOnInterval(t *testing.T) {
	var count atomic.Int32

	sched := NewScheduler(100*time.Millisecond, func() {
		count.Add(1)
	})
	sched.Start()
	time.Sleep(350 * time.Millisecond)
	sched.Stop()

	got := count.Load()
	if got < 2 || got > 4 {
		t.Errorf("expected 2-4 runs, got %d", got)
	}
}

func TestSchedulerStop(t *testing.T) {
	var count atomic.Int32

	sched := NewScheduler(50*time.Millisecond, func() {
		count.Add(1)
	})
	sched.Start()
	time.Sleep(100 * time.Millisecond)
	sched.Stop()
	after := count.Load()
	time.Sleep(100 * time.Millisecond)

	if count.Load() != after {
		t.Error("scheduler continued running after stop")
	}
}
