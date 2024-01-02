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

type Task struct {
	ctx   context.Context
	Id    string
	Value any
	mu    sync.Mutex
	cond  *sync.Cond
	err   error
}

func NewTask(req *Request) *Task {
	task := &Task{
		ctx:   req.Ctx,
		Id:    req.Id,
		Value: req.Value,
	}
	task.cond = sync.NewCond(&task.mu)

	return task
}

func (task *Task) Result() (any, error) {
	task.mu.Lock()
	task.cond.Wait()
	task.mu.Unlock()

	if task.err != nil {
		logger.Debugf("in the end, Error: %v\n", task.err)
		return nil, task.err
	}
	if e := context.Cause(task.ctx); e != nil {
		task.err = e
		logger.Debugf("Cause, Error: %v\n", e)
		return nil, e
	}

	return nil, nil
}

func (task *Task) signal(err error) {
	task.mu.Lock()
	defer task.mu.Unlock()

	task.err = err
	task.cond.Signal()
}
