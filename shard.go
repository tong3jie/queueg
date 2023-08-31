package queueg

import "sync/atomic"

type shard[T any] struct {
	size         atomic.Int64
	Chan         chan T
	isRun        atomic.Bool
	panicHandler func(e interface{})
}

func (s *shard[T]) run(fn func(T)) {
	defer func() {
		s.isRun.Store(false)
		if r := recover(); r != nil {
			s.panicHandler(r)
		}
	}()

	s.isRun.Store(true)
	for {
		select {
		case v, ok := <-s.Chan:
			if !ok {
				return
			}
			fn(v)
			s.size.Add(-1)
		default:
			select {
			case v, ok := <-s.Chan:
				if !ok {
					return
				}
				fn(v)
				s.size.Add(-1)
			}
		}
	}
}
