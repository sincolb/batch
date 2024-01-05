package batch

import (
	"context"
	"fmt"
	"sync"
)

type DuplicateUniqId struct {
	msg any
}

func (e DuplicateUniqId) Error() string {
	return fmt.Sprintf("duplicate uniqID: %s", e.msg.(string))
}

type Task[T any] struct {
	ctx   context.Context
	Id    any
	Value T
	mu    sync.Mutex
	cond  *sync.Cond
	err   error
}

func (task *Task[T]) Key() string {
	return fmt.Sprintf("%v", task.Id)
}

func (task *Task[T]) Wait() error {
	task.mu.Lock()
	task.cond.Wait()
	task.mu.Unlock()

	if task.err != nil {
		return task.err
	}
	if e := context.Cause(task.ctx); e != nil {
		task.err = e
		return e
	}

	return nil
}

func (task *Task[T]) signal(err error) {
	task.mu.Lock()
	defer task.mu.Unlock()

	task.err = err
	task.cond.Signal()
}
