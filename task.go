package batchrequests

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
	Id    string
	Value T
	mu    sync.Mutex
	cond  *sync.Cond
	err   error
}

func NewTask[T any](req *Request[T]) *Task[T] {
	task := &Task[T]{
		ctx:   req.Ctx,
		Id:    req.Id,
		Value: req.Value,
	}
	task.cond = sync.NewCond(&task.mu)

	return task
}

type TaskIface interface {
	Result() (any, error)
	Context() context.Context
	Key() string
}

func (task *Task[T]) Key() string {
	return task.Id
}
func (task *Task[T]) Context() context.Context {
	return task.ctx
}

func (task *Task[T]) Result() error {
	task.mu.Lock()
	task.cond.Wait()
	task.mu.Unlock()

	if task.err != nil {
		logger.Debugf("in the end, Error: %v\n", task.err)
		return task.err
	}
	if e := context.Cause(task.ctx); e != nil {
		task.err = e
		logger.Debugf("Cause, Error: %v\n", e)
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
