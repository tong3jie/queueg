package queueg

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSize(t *testing.T) {
	q := NewQueue[int](10000)

	for i := 0; i < 10000; i++ {
		q.Push(i)
	}

	assert.Equal(t, 10000, q.Size())
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
	assert.Equal(t, 0, q.Size())
}

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
