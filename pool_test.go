package wpool

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		maxWorkers    int
		maxBufferChan int
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_new_dispatcher with default values",
			args: args{
				maxWorkers:    5,
				maxBufferChan: 20,
			},
		},
		{
			name: "test_new_dispatcher other values ",
			args: args{
				maxWorkers:    10,
				maxBufferChan: 25,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(
				WithNumWorkers(tt.args.maxWorkers),
				WithBufferSize(tt.args.maxBufferChan),
			)

			if got == nil {
				t.Errorf("New()() = %v, want not nil", got)
			}

			if got.nWorkers != tt.args.maxWorkers {
				t.Errorf("New()() = %v, want max workers %v", got.workers, tt.args.maxWorkers)
			}

			if cap(got.taskBufferedChannel) != tt.args.maxBufferChan {
				t.Errorf("New()() = %v, want a buffer capacity %v", cap(got.taskBufferedChannel), tt.args.maxBufferChan)
			}

			if cap(got.workers) != tt.args.maxWorkers {
				t.Errorf("New()() = %v, want a workers capacity %v", cap(got.workers), tt.args.maxWorkers)
			}
		})
	}
}

func TestDispatcher_AddTask(t *testing.T) {
	var dummyFunc = func() {
		t.Log("doing work")
	}

	type fields struct {
		maxWorkers          int
		workers             []worker
		taskBufferedChannel chan Task
	}

	type args struct {
		assertions func(*WorkerPool)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test_add_task empty buffer",
			fields: fields{
				maxWorkers:          5,
				workers:             make([]worker, 0, 5),
				taskBufferedChannel: make(chan Task, 1),
			},
			args: args{
				assertions: func(d *WorkerPool) {
					task := NewTask(nil, dummyFunc)

					d.AddTask(task)

					if len(d.taskBufferedChannel) != 1 {
						t.Errorf("AddTask() = %v, want 1 task in buffer", len(d.taskBufferedChannel))
					}
				},
			},
		},
		{
			name: "test_add_task full buffer",
			fields: fields{
				maxWorkers:          5,
				workers:             make([]worker, 0, 5),
				taskBufferedChannel: make(chan Task, 1),
			},

			args: args{func(d *WorkerPool) {
				task := NewTask(nil, dummyFunc)
				d.AddTask(task) // Put 1 task in the channel

				select {
				case d.taskBufferedChannel <- NewTask(nil, dummyFunc): // Put 2 in the channel
					t.Errorf("AddTask() = %v, want 1 task in buffer", len(d.taskBufferedChannel))
				default:
					t.Log("buffer is full")
				}
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &WorkerPool{
				nWorkers:            tt.fields.maxWorkers,
				workers:             tt.fields.workers,
				taskBufferedChannel: tt.fields.taskBufferedChannel,
			}

			tt.args.assertions(d)
		})
	}
}

func TestDispatcher_Run_Stop_Workers(t *testing.T) {
	type fields struct {
		maxWorkers          int
		workers             []worker
		taskBufferedChannel chan Task
	}

	type args struct {
		assertions func(int, *WorkerPool)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test_run_stop_workers with default values",
			fields: fields{
				maxWorkers:          5,
				workers:             make([]worker, 5),
				taskBufferedChannel: make(chan Task, 20),
			},
			args: args{func(numGoRoutines int, dispatcher *WorkerPool) {
				beforeNumGoroutine := runtime.NumGoroutine()

				stopFunc := dispatcher.Run()

				actualNumGoroutine := runtime.NumGoroutine()

				i := actualNumGoroutine - beforeNumGoroutine

				if i != numGoRoutines {
					t.Errorf("Run() = want %d goroutines but got %d", numGoRoutines, i)
				}

				stopFunc()

				afterNumGoroutine := runtime.NumGoroutine()

				if afterNumGoroutine != beforeNumGoroutine {
					t.Errorf("Stop() = %d, want %d goroutines", afterNumGoroutine, beforeNumGoroutine)
				}
			}},
		},
		{
			name: "test_run_stop_workers with other values",
			fields: fields{
				maxWorkers:          3,
				workers:             make([]worker, 3),
				taskBufferedChannel: make(chan Task, 20),
			},
			args: args{func(numGoRoutines int, dispatcher *WorkerPool) {
				beforeNumGoroutine := runtime.NumGoroutine()

				stopFunc := dispatcher.Run()

				actualNumGoroutine := runtime.NumGoroutine()

				i := actualNumGoroutine - beforeNumGoroutine

				if i != numGoRoutines {
					t.Errorf("Run() = want %d goroutines but got %d", numGoRoutines, i)
				}

				stopFunc()

				afterNumGoroutine := runtime.NumGoroutine()

				if afterNumGoroutine != beforeNumGoroutine {
					t.Errorf("Stop() = %d, want %d goroutines", afterNumGoroutine, beforeNumGoroutine)
				}
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &WorkerPool{
				nWorkers:            tt.fields.maxWorkers,
				workers:             tt.fields.workers,
				taskBufferedChannel: tt.fields.taskBufferedChannel,
			}

			tt.args.assertions(tt.fields.maxWorkers, d)
		})
	}
}

func TestDispatcher_RunTask(t *testing.T) {
	type fields struct {
		maxWorkers             int
		capTaskBufferedChannel int
		val                    int32
	}

	type args struct {
		assertions func(int32, *WorkerPool)
	}

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test_execute_dummy_task with 1 worker 1 Task",
			fields: fields{
				maxWorkers:             1,
				capTaskBufferedChannel: 10,
			},
			args: args{
				assertions: func(val int32, d *WorkerPool) {
					var sum atomic.Int32
					var dummyFunc = func(val int32) {
						sum.Add(val)
					}

					d.Run()

					var wg sync.WaitGroup
					wg.Add(1)

					task := NewTask(&wg, func() {
						dummyFunc(val)
					})

					d.AddTask(task)

					wg.Wait()

					if sum.Load() != val {
						t.Errorf("ExecuteTask() = %v, want %d", val, sum.Load())
					}
				},
			},
		},
		{
			name: "test_execute_dummy_task with 1 worker 2 Task",
			fields: fields{
				maxWorkers:             1,
				capTaskBufferedChannel: 10,
			},
			args: args{
				assertions: func(val int32, d *WorkerPool) {
					var sum atomic.Int32
					var dummyFunc = func(val int32) {
						sum.Add(val)
					}

					d.Run()

					var wg sync.WaitGroup
					wg.Add(2)

					task := NewTask(&wg, func() {
						dummyFunc(val)
					})

					d.AddTask(task)
					d.AddTask(task)

					wg.Wait()

					if sum.Load() != val*2 {
						t.Errorf("ExecuteTask() = %v, want %d", val, sum.Load())
					}
				},
			},
		},
		{
			name: "test_execute_dummy_task with 2 worker 2 Task",
			fields: fields{
				maxWorkers:             2,
				capTaskBufferedChannel: 10,
			},
			args: args{
				assertions: func(val int32, d *WorkerPool) {
					var sum atomic.Int32
					var dummyFunc = func(val int32) {
						sum.Add(val)
					}

					d.Run()

					var wg sync.WaitGroup
					wg.Add(2)

					task := NewTask(&wg, func() {
						dummyFunc(val)
					})

					d.AddTask(task)
					d.AddTask(task)

					wg.Wait()

					if sum.Load() != val*2 {
						t.Errorf("ExecuteTask() = %v, want %d", val, sum.Load())
					}
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(
				WithNumWorkers(tt.fields.maxWorkers),
				WithBufferSize(tt.fields.capTaskBufferedChannel),
			)

			tt.args.assertions(tt.fields.val, d)
		})
	}
}
