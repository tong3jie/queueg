package queueg

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"sync/atomic"
	"time"
)

const MaxIndex = 20

func defaultStackTraceHandler(e interface{}) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf("panic: %v\n%s\n", e, debug.Stack())) //nolint:errcheck // This will never fail
}

type Queue[T any] struct {
	size      atomic.Int64
	shards    []shard[T]
	inoffset  atomic.Int32
	outoffset atomic.Int32
	fn        func(T)
	pool      chan struct{}
}

type shard[T any] struct {
	size  atomic.Int64
	Chan  chan T
	isRun atomic.Bool
}

func NewQueue[T any](maxSize int) *Queue[T] {

	size := float64(maxSize) / 20

	q := Queue[T]{
		size:      atomic.Int64{},
		inoffset:  atomic.Int32{},
		outoffset: atomic.Int32{},
		shards:    make([]shard[T], MaxIndex),
		pool:      make(chan struct{}, MaxIndex*2),
	}

	for i := 0; i < MaxIndex*2; i++ {
		q.pool <- struct{}{} // 初始化 goroutine 池
	}

	for i := 0; i < MaxIndex; i++ {
		q.shards[i] = shard[T]{
			size:  atomic.Int64{},
			Chan:  make(chan T, int(math.Ceil(size))),
			isRun: atomic.Bool{},
		}
	}

	return &q
}

func NewQueueWithFn[T any](maxSize int, fn func(T)) *Queue[T] {
	q := NewQueue[T](maxSize)
	q.fn = fn
	return q
}

func (s *shard[T]) run(fn func(T)) {
	defer func() {
		s.isRun.Store(false)
		if r := recover(); r != nil {
			defaultStackTraceHandler(r)
		}
	}()

	s.isRun.Store(true)
	for {
		select {
		case v, ok := <-s.Chan:
			if !ok {
				return
			}
			s.size.Add(-1)
			fn(v)
		default:
		}
	}
}

// 随机获取一个index
func (q *Queue[T]) getInIndex() int {
	n := q.inoffset.Add(1)
	if n > MaxIndex-1 {
		q.inoffset.Store(0)
	}
	return int(n - 1)
}

// 随机获取一个index
func (q *Queue[T]) getOutIndex() int {
	n := q.outoffset.Add(1)
	if n > MaxIndex-1 {
		q.outoffset.Store(0)
	}
	return int(n - 1)
}

// 往队列中添加一个元素
// 该方法为阻塞式

func (q *Queue[T]) Push(v T) {
	// 首先，尝试将消息写入当前索引的 shard
	index := q.getInIndex()
	select {
	case q.shards[index].Chan <- v:
		q.shards[index].size.Add(1)
	default:
		// 如果当前 shard 已满，尝试在其他 shard 中写入
		for i := 0; i < MaxIndex; i++ {
			if len(q.shards[i].Chan) < cap(q.shards[i].Chan) {
				select {
				case q.shards[i].Chan <- v:
					q.shards[i].size.Add(1)
					return
				default:
				}
			}
		}
		// 如果所有 shard 都已满，阻塞等待直到有空位可用
		q.shards[index].Chan <- v
		q.shards[index].size.Add(1)
	}
}

// 从队列中出队一个元素，出队即删除
// 当队列为空时则会进行阻塞
func (q *Queue[T]) Pop() T {
	// 首先，尝试从当前索引的 shard 中出队
	index := q.getOutIndex()
	select {
	case v := <-q.shards[index].Chan:
		q.shards[index].size.Add(-1)
		return v
	default:
		// 如果当前 shard 为空，尝试从其他 shard 中出队
		for i := 0; i < MaxIndex; i++ {
			if len(q.shards[i].Chan) > 0 {
				select {
				case v := <-q.shards[i].Chan:
					q.shards[i].size.Add(-1)
					return v
				default:
				}
			}
		}
		// 如果所有 shard 都为空，阻塞等待直到有元素可出队
		v := <-q.shards[index].Chan
		q.shards[index].size.Add(-1)
		return v
	}
}

// 获取当前队列的大小
func (q *Queue[T]) Size() int64 {
	totalSize := int64(0)
	for i := 0; i < MaxIndex; i++ {
		totalSize += q.shards[i].size.Load()
	}
	q.size.Store(totalSize)
	return q.size.Load()
}

// 若创建的队列有回调方法，则可以使用run方式自动执行
func (q *Queue[T]) Run() {

	if q.fn != nil {
		for i := 0; i < MaxIndex; i++ {
			<-q.pool // 从池中获取一个空位
			go func(i int) {
				q.shards[i].run(q.fn)
			}(i)
		}
		go q.restartShart()

	} else {
		panic("fn is nil, please user NewQueueWithFn")
	}
}

// restartShard restarts the shards in the queue.
func (q *Queue[T]) restartShart() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		for i := 0; i < MaxIndex; i++ {
			if !q.shards[i].isRun.Load() {
				select {
				case <-q.pool: // 从池中获取一个空位
					go func(i int) {
						defer func() {
							q.pool <- struct{}{} // 将空位放回池中
						}()
						q.shards[i].run(q.fn)

					}(i)

				default:
				}

			}
		}
	}
}

// 关闭所有的channel，并保证所有channel已消费完毕
func (q *Queue[T]) Close() {
	for i := 0; i < MaxIndex; i++ {
		close(q.shards[i].Chan)
	}
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	runTotal := 0
	for range ticker.C {
		for i := 0; i < MaxIndex; i++ {
			if q.shards[i].isRun.Load() {
				runTotal++
			}
		}
		if runTotal == 0 {
			return
		}
	}
}
