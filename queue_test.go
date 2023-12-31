package queueg

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ABC struct {
	A int
	B int
	C int
	D int
	E int
	F int
	G int
	H int
	I int
}

func TestSize(t *testing.T) {
	q := NewQueue[int](10000)

	for i := 0; i < 10000; i++ {
		q.Push(i)
	}

	assert.Equal(t, int64(10000), q.Size())
}

func TestPoP(t *testing.T) {
	q := NewQueueWithFn[*ABC](100000000, func(abc *ABC, part int) {

		fmt.Println(abc, part)

	})
	for i := 0; i < 1000; i++ {
		q.Push(&ABC{i, i, i, i, i, i, i, i, i})
	}

	q.Run()
	time.Sleep(time.Second * 1)
	assert.Equal(t, int64(0), q.Size())
}

func TestPoP2(t *testing.T) {
	total := atomic.Int32{}
	fn := func(abc *ABC, part int) {
		total.Add(1)
	}
	q := NewQueueWithFn[*ABC](1000, fn)
	for i := 0; i < 1000; i++ {
		q.Push(&ABC{i, i, i, i, i, i, i, i, i})
	}
	assert.Equal(t, int64(1000), q.Size())

	q.Run()
	time.Sleep(time.Second * 1)
	assert.Equal(t, int64(0), q.Size())
	time.Sleep(time.Second * 1)
	assert.Equal(t, int32(1000), total.Load())
}

func TestPanic(t *testing.T) {
	q := NewQueueWithFn[*ABC](100000000, func(abc *ABC, part int) {

		fmt.Println(abc)
		if abc.A == 100 {
			panic("100")
		}

	})
	for i := 0; i < 1000; i++ {
		q.Push(&ABC{i, i, i, i, i, i, i, i, i})
	}

	q.Run()
	time.Sleep(time.Second * 1)
	assert.Equal(t, 999, 999)
}

func TestShardsMax(t *testing.T) {
	q := New[int](WithShardsMax[int](100))
	assert.Equal(t, int64(100), q.shardsMax.Load())
}

func TestWithDefault(t *testing.T) {
	o := NewOption[int]()
	q := New[int](o)
	assert.Equal(t, int64(20), q.shardsMax.Load())
	assert.Equal(t, 10, cap(q.shards[0].Chan))
}

func TestWithSize(t *testing.T) {
	q := New[int](WithSize[int](100))
	assert.Equal(t, int(5), cap(q.shards[0].Chan))
}

func TestWithHash(t *testing.T) {
	q := NewQueueWithFn[int](1000, func(i int, part int) {
		fmt.Println(i)
	})
	go func() {
		for i := 0; i < 1000; i++ {
			q.PushByHash(i, "100")
		}
	}()

	q.Run()

	time.Sleep(time.Second * 10)

}

func BenchmarkPushInt(b *testing.B) {
	q := NewQueue[int](100000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push(i)
	}
	b.StopTimer()
}

func BenchmarkPushString(b *testing.B) {
	q := NewQueue[string](100000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push(strconv.Itoa(i))
	}
	b.StopTimer()
}

func BenchmarkPushSlice(b *testing.B) {
	q := NewQueue[[]string](100000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push([]string{strconv.Itoa(i), strconv.Itoa(i), strconv.Itoa(i)})
	}
	b.StopTimer()
}

func BenchmarkPushMap(b *testing.B) {
	q := NewQueue[map[string]int](100000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push(map[string]int{strconv.Itoa(i): i})
	}
	b.StopTimer()
}

func BenchmarkStruct(b *testing.B) {
	q := NewQueue[*ABC](100000000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push(&ABC{i, i, i, i, i, i, i, i, i})
	}
	b.StopTimer()
}

func BenchmarkByte(b *testing.B) {
	q := NewQueue[[]byte](100000000)
	s := strconv.Itoa(10000) + strconv.Itoa(10000) + strconv.Itoa(10000)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		q.Push([]byte(s))
	}
	b.StopTimer()
}
