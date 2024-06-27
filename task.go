package wpool

import (
	"sync"

	"github.com/google/uuid"
)

// Task represents a task to be executed by a worker.
type Task struct {
	ID     string          // ID is the unique identifier of the task.
	wg     *sync.WaitGroup // wg is the wait group used to wait for the task to be done.
	doWork func()          // doWork is the function that the task executes.
}

// NewTask creates a new task with the given work function.
func NewTask(wg *sync.WaitGroup, work func()) Task {
	return Task{
		ID:     uuid.New().String(),
		wg:     wg,
		doWork: work,
	}
}

// Do execute the task's work function and signals the wait group that the task is done.
func (t Task) Do() {
	t.doWork()

	t.wg.Done()
}
