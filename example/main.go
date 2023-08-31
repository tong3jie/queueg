package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/tong3jie/queueg"
)

func main() {
	q := queueg.NewQueueWithFn[string](1000, func(s string) {
		Print(s)
	})
	for i := 0; i < 1000; i++ {
		q.Push(strconv.Itoa(i))
	}
	q.Run()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	s := <-signalChan
	fmt.Println("退出信号", s)
}

func Print(s string) {
	fmt.Println(s)
}
