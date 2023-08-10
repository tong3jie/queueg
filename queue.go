package queueg

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

const MaxIndex = 20

func defaultStackTraceHandler(e interface{}) {
	_, _ = os.Stderr.WriteString(fmt.Sprintf("panic: %v\n%s\n", e, debug.Stack())) //nolint:errcheck // This will never fail
}

type Queue[T any] struct {
	size      int
	shards    []shard[T]
	inoffset  atomic.Int32
	outoffset atomic.Int32
	fn        func(T)
	wg        sync.WaitGroup
	pool      chan struct{}
}

type shard[T any] struct {
	shard chan T
	isRun atomic.Bool
}

func NewQueue[T any](maxSize int) *Queue[T] {

	size := float64(maxSize) / 20

	q := Queue[T]{size: maxSize, inoffset: atomic.Int32{}, outoffset: atomic.Int32{}}

	q.shards = make([]shard[T], MaxIndex)
	q.pool = make(chan struct{}, MaxIndex*2)
	for i := 0; i < MaxIndex; i++ {
		q.pool <- struct{}{} // 初始化 goroutine 池
		q.shards[i] = shard[T]{shard: make(chan T, int(math.Ceil(size))), isRun: atomic.Bool{}}
	}
	q.wg = sync.WaitGroup{}
	
	return &q
}

func NewQueueWithFn[T any](maxSize int, fn func(T)) *Queue[T] {
	q := NewQueue[T](maxSize)
	q.fn = fn
	return q
}

func (s *shard[T]) push(v T) {
	s.shard <- v
}

func (s *shard[T]) pop() T {
	return <-s.shard
}

func (s *shard[T]) run(fn func(T)) {
	defer func() {
		s.isRun.Store(false)
		if r := recover(); r != nil {
			defaultStackTraceHandler(r)
		}
	}()

	s.isRun.Store(true)
	for v := range s.shard {
		fn(v)
	}
	s.isRun.Store(false)
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
	index := q.getInIndex()
	q.shards[index].push(v)
}

// 从队列中出队一个元素，出队即删除
// 当队列为空时则会进行阻塞
func (q *Queue[T]) Pop() T {
	index := q.getOutIndex()
	return q.shards[index].pop()
}

// 获取当前队列的大小
func (q *Queue[T]) Size() int {
	n := 0
	for indx := range q.shards {
		n += len(q.shards[indx].shard)
	}
	return n
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

func (q *Queue[T]) restartShart() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for i := 0; i < MaxIndex; i++ {
			if !q.shards[i].isRun.Load() {
				defer func() {
					q.pool <- struct{}{} // 将空位放回池中
				}()
				<-q.pool // 从池中获取一个空位
				fmt.Println(i, len(q.shards[i].shard))
				q.shards[i].run(q.fn)
			}
		}
	}
}
