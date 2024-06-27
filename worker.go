package wpool

// worker represents a worker that executes tasks.
type worker struct {
	ID         int
	taskBuffer chan Task
	quit       chan bool
}

// newWorker creates a new worker with the given ID and task buffer.
func newWorker(id int, tasksBuff chan Task) worker {
	return worker{
		ID:         id,
		taskBuffer: tasksBuff,
		quit:       make(chan bool),
	}
}

// Start starts the worker's goroutine.
// It continuously listens for tasks in the task buffer and executes them.
// If the worker's quit channel receives a signal, it stops the goroutine.
func (w worker) Start() {
	go func() {
		for {
			select {
			case task := <-w.taskBuffer:
				task.Do()

			case <-w.quit:
				return
			}
		}
	}()
}

// Stop stops the worker's goroutine.
// It sends a signal to the worker's quit channel to stop the goroutine.
func (w worker) Stop() {
	w.quit <- true
}
