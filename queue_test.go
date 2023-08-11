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
	q := NewQueueWithFn[*ABC](100000000, func(abc *ABC) {

		fmt.Println(abc)

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
	fn := func(abc *ABC) {
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
