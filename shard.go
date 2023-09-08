package queueg

import (
	"context"
	"sync/atomic"
)

type shard[T any] struct {
	size         atomic.Int64
	Chan         chan T
	isRun        atomic.Bool
	panicHandler func(e interface{})
	ctx          context.Context
	partion      int
}

func (s *shard[T]) run(fn func(value T, partion int)) {
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
			fn(v, s.partion)
			s.size.Add(-1)

		case <-s.ctx.Done():
			return
		}
	}
}
