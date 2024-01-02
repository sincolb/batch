package batchrequests

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type dispatch struct {
	pool  sync.Map
	exitC chan struct{}
}

func NewDispatch() *dispatch {
	d := &dispatch{
		exitC: make(chan struct{}),
	}

	return d
}

func (d *dispatch) Get(uniqID any) (*Worker, bool) {
	value, ok := d.pool.Load(uniqID)
	if !ok {
		return nil, false
	}
	if work, ok := value.(*Worker); ok {
		return work, true
	}
	return nil, false
}

func (d *dispatch) Register(uniqID any, batchSize int, AutoCommitDuration time.Duration,
	handle Handler, opts ...Option) error {
	// switch handle.(type) {
	// case Handler[*Task], Handler[[]*Task]:
	// default:
	// 	return errors.New("not support")
	// }

	if _, ok := d.pool.Load(uniqID); ok {
		return DuplicateUniqId{uniqID}
	}

	work := &Worker{
		Key:                uniqID,
		dispatch:           d,
		batchSize:          batchSize,
		bufferSize:         1000,
		Retrys:             3,
		Handle:             handle,
		submitTimeOut:      3000 * time.Millisecond,
		autoCommitDuration: AutoCommitDuration,
		graceDwonDuration:  3 * time.Second,
		exit:               make(chan struct{}),
	}
	for _, opt := range opts {
		opt.apply(work)
	}
	work.taskC = make(chan *Task, work.bufferSize)
	d.pool.Store(uniqID, work)

	go work.worker()

	return nil
}

func (d *dispatch) Unregister(uniqID any) {
	value, b := d.pool.Load(uniqID)
	if !b {
		return
	}

	work := value.(*Worker)
	work.stop()
	d.pool.Delete(uniqID)
}

func (d *dispatch) UnregisterAll() {
	d.pool.Range(func(key, value any) bool {
		if work, ok := value.(*Worker); ok {
			work.stop()
			d.pool.Delete(key)
		}

		return true
	})
}

func (d *dispatch) Submit(req *Request) (*Task, error) {
	task := NewTask(req)

	select {
	case <-d.exitC:
		logger.Debugln("ALL received exit 2")

		return nil, errors.New("all received exit 2")
	case <-task.ctx.Done():
		logger.Debugln("context cancel 0, request=", task)

		return nil, errors.New("context cancel 0")
	default:
		work, ok := d.Get(task.Id)
		if !ok {
			logger.Debugln("A not register key=", task.Id)

			return nil, fmt.Errorf("a not register key=%s", task.Id)
		}
		if work.closed() {
			logger.Debugf("A queue[%v] is closed\n", work.Key)

			return nil, fmt.Errorf("a queue[%v] is closed", work.Key)
		}

		select {
		case work.taskC <- task:
			return task, nil
		case <-work.exit:
			logger.Debugf("work received exit %s", work.Key)

			return nil, errors.New("work received exit")
		case <-task.ctx.Done():
			logger.Debugln("context cancel 3, request=", task)

			return nil, errors.New("context cancel 3,")
		case <-d.exitC:
			logger.Debugln("ALL received exit 2")

			return nil, errors.New("aLL received exit 2")
		}
	}
}

func (d *dispatch) Release() {
	d.UnregisterAll()
	close(d.exitC)
	logger.Debugln("close end-------------")
}

func (d *dispatch) Exit() <-chan struct{} {

	return d.exitC
}
