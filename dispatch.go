package batch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type dispatch[T any] struct {
	pool  sync.Map
	exitC chan struct{}
}

func NewDispatch[T any]() *dispatch[T] {
	d := &dispatch[T]{
		exitC: make(chan struct{}),
	}

	return d
}

func (d *dispatch[T]) Get(uniqID any) (*Worker[T], bool) {
	value, ok := d.pool.Load(uniqID)
	if !ok {
		return nil, false
	}
	if work, ok := value.(*Worker[T]); ok {
		return work, true
	}
	return nil, false
}

func (d *dispatch[T]) Register(uniqID string, batchSize int, AutoCommitDuration time.Duration,
	handle Handler[T], opts ...Option[T]) error {
	// switch handle.(type) {
	// case Handler[*Task], Handler[[]*Task]:
	// default:
	// 	return errors.New("not support")
	// }

	if _, ok := d.pool.Load(uniqID); ok {
		return DuplicateUniqId{uniqID}
	}

	work := &Worker[T]{
		Key:                uniqID,
		dispatch:           d,
		batchSize:          batchSize,
		bufferSize:         1000,
		Retrys:             3,
		Handle:             handle,
		batchTimeOut:       3000 * time.Millisecond,
		autoCommitDuration: AutoCommitDuration,
		graceDwonDuration:  3 * time.Second,
		exit:               make(chan struct{}),
	}
	for _, opt := range opts {
		opt.apply(work)
	}
	work.taskC = make(chan *Task[T], work.bufferSize)
	d.pool.Store(uniqID, work)

	go work.worker()

	return nil
}

func (d *dispatch[T]) Unregister(uniqID any) {
	value, b := d.pool.Load(uniqID)
	if !b {
		return
	}

	work := value.(*Worker[T])
	work.stop()
	d.pool.Delete(uniqID)
}

func (d *dispatch[T]) UnregisterAll() {
	d.pool.Range(func(key, value any) bool {
		if work, ok := value.(*Worker[T]); ok {
			work.stop()
			d.pool.Delete(key)
		}

		return true
	})
}

func (d *dispatch[T]) Submit(key string, value T) (*Task[T], error) {

	return d.submit(context.Background(), key, value)
}

func (d *dispatch[T]) SubmitWithContext(ctx context.Context, key any, value T) (*Task[T], error) {

	return d.submit(ctx, key, value)
}

func (d *dispatch[T]) submit(ctx context.Context, key any, value T) (*Task[T], error) {
	task := &Task[T]{
		ctx:   ctx,
		Id:    key,
		Value: value,
	}
	task.cond = sync.NewCond(&task.mu)

	select {
	case <-d.exitC:
		return nil, errors.New("all received exit 2")
	case <-task.ctx.Done():
		return nil, errors.New("context cancel 0")
	default:
		work, ok := d.Get(task.Id)
		if !ok {
			return nil, fmt.Errorf("a not register key=%s", task.Id)
		}
		if work.closed() {
			return nil, fmt.Errorf("a queue[%v] is closed", work.Key)
		}

		select {
		case work.taskC <- task:
			return task, nil
		case <-work.exit:
			return nil, errors.New("work received exit")
		case <-task.ctx.Done():
			return nil, errors.New("context cancel 3,")
		case <-d.exitC:
			return nil, errors.New("aLL received exit 2")
		}
	}
}

func (d *dispatch[T]) Release() {
	d.UnregisterAll()
	close(d.exitC)
}

func (d *dispatch[T]) Exit() <-chan struct{} {

	return d.exitC
}
