package wpool

import (
	"context"
	"time"
)

const (
	MaxWorkers    = 1000 // is the maximum number of workers in the pool.
	MaxBufferSize = 1000 // is the maximum buffer size of the task channel.
)

// Option is a function that configures a WorkerPool.
// It is used to set the number of workers and the buffer size of the task channel.
type Option func(*WorkerPool)

// WithNumWorkers is an option to set the number of workers in the pool.
// If the number of workers is greater than the maximum number of workers,
// it will be set to the [MaxWorkers].
func WithNumWorkers(nWorkers int) Option {
	return func(d *WorkerPool) {
		if nWorkers > MaxWorkers {
			nWorkers = MaxWorkers
		}

		d.nWorkers = nWorkers
	}
}

// WithBufferSize is an option to set the buffer size of the task channel.
// If the buffer size is greater than the maximum buffer size, it will be set
// to the [MaxBufferSize].
func WithBufferSize(bufferSize int) Option {
	return func(d *WorkerPool) {
		if bufferSize > MaxBufferSize {
			bufferSize = MaxBufferSize
		}

		d.bufferSize = bufferSize
	}
}

// StopFunc is a function that stops the worker pool.
// It is used to stop the worker pool gracefully.
type StopFunc func()

// WorkerPool is a pool of workers that can execute tasks concurrently.
// It uses a buffered channel to receive tasks and distribute them to the workers.
type WorkerPool struct {
	nWorkers   int
	bufferSize int

	workers             []worker
	taskBufferedChannel chan Task
}

// New creates a new worker pool with the given options.
// The default number of workers is 2 and the default buffer size is 10.
// The number of workers and the buffer size can be changed using the options.
// The maximum number of workers is 1000 and the maximum buffer size is 1000.
func New(opts ...Option) *WorkerPool {
	d := &WorkerPool{
		nWorkers:   2,
		bufferSize: 10,
	}

	for _, o := range opts {
		o(d)
	}

	d.taskBufferedChannel = make(chan Task, d.bufferSize)
	d.workers = make([]worker, d.nWorkers)

	return d
}

// AddTask adds a task to the worker pool.
// The task is added to the task channel, which is a buffered channel.
// If the task channel is full, AddTask will block until there is space in the channel.
func (d *WorkerPool) AddTask(task Task) {
	d.taskBufferedChannel <- task
}

// Run starts the worker pool.
// It creates the workers, starts them, and returns a function that can be used to stop the worker pool.
// The function uses a context with a deadline to stop the workers gracefully.
// It waits for the deadline to expire and then stops the workers forcefully if they haven't stopped yet.
func (d *WorkerPool) Run() StopFunc {
	for i := 0; i < d.nWorkers; i++ {
		worker := newWorker(i+1, d.taskBufferedChannel)
		worker.Start()

		d.workers[i] = worker
	}

	return d.stop
}

// stop stops the worker pool.
// It stops all the workers and waits for them to stop.
// It uses a context with a deadline to stop the workers gracefully.
// It waits for the deadline to expire and then stops the workers forcefully if they haven't stopped yet.
func (d *WorkerPool) stop() {
	timeout := time.Duration(d.nWorkers) * (10 * time.Millisecond)

	if timeout > 3*time.Second {
		timeout = 3 * time.Second
	}

	deadline, cancelFunc := context.WithDeadline(context.Background(), time.Now().Add(timeout))
	defer cancelFunc()

	for i := 0; i < d.nWorkers; i++ {
		w := d.workers[i]
		w.Stop()
	}

	// Wait some time for all workers to stop
	<-deadline.Done()
}
