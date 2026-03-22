package scanner

import "time"

type Scheduler struct {
	interval time.Duration
	fn       func()
	stop     chan struct{}
	done     chan struct{}
}

func NewScheduler(interval time.Duration, fn func()) *Scheduler {
	return &Scheduler{
		interval: interval,
		fn:       fn,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (s *Scheduler) Start() {
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.fn()
			case <-s.stop:
				return
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	close(s.stop)
	<-s.done
}
