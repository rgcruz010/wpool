# Go Worker Pool

This library provides a simple and efficient worker pool implementation in Go. It allows you to easily manage and control the execution of
tasks concurrently.

## Features

- Configurable number of workers and task buffer size.
- Graceful shutdown of workers.
- Task execution with wait groups.

## Usage

Here is an example of how to use this library:

```go

package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/rgcruz010/wpool"
)

func main() {
	// Create a new worker pool
	pool := wpool.New(wpool.WithNumWorkers(2), wpool.WithBufferSize(5))

	// Run the worker pool
	stop := pool.Run()

	// Stop the worker pool
	defer stop()

	// Create a wait group
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		// Add count to the wait group
		wg.Add(1)

		// Create a new task
		task := wpool.NewTask(&wg, func() {
			fmt.Println("Hello, World! I'm doing some hard work...")
			time.Sleep(1 * time.Second)
		})

		// Add the task to the worker pool
		pool.AddTask(task)
	}

	// Wait for all tasks to complete
	wg.Wait()

}

```

## Contributing

If you find a bug or have a feature request, please open an issue. We welcome contributions in the form of pull requests as well.