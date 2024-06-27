package wpool_test

import (
	"log/slog"
	"sync"
	"time"

	"github.com/rgcruz010/wpool"
)

func ExampleNew_default() {
	p := wpool.New()
	stopFunc := p.Run()

	defer stopFunc()
}

func ExampleNew_withTask() {
	p := wpool.New()

	stopFunc := p.Run()
	defer stopFunc()

	var wg sync.WaitGroup

	wg.Add(10)

	start := time.Now()

	for i := 0; i < 10; i++ {
		p.AddTask(wpool.NewTask(&wg, func() {
			// do hard work
			time.Sleep(10 * time.Millisecond)
		}))
	}

	wg.Wait()

	slog.Info("exampleWithDefault", slog.String("time", time.Since(start).String()))
}

func ExampleWithNumWorkers() {
	p := wpool.New(
		wpool.WithNumWorkers(100),
	)
	stopFunc := p.Run()

	defer stopFunc()
}

func ExampleWithBufferSize() {
	p := wpool.New(
		wpool.WithBufferSize(10),
	)
	stopFunc := p.Run()

	defer stopFunc()
}
