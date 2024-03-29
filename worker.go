package batch

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type Worker[T any] struct {
	mu         sync.Mutex
	dispatch   *dispatch[T]
	Key        any
	batchSize  int
	bufferSize int
	Retrys     int
	// Handle             any
	Handle             Handler[T]
	taskC              chan *Task[T]
	batchTimeOut       time.Duration
	graceDwonDuration  time.Duration
	autoCommitDuration time.Duration
	exit               chan struct{}
	_closed            uint32
}

func (p *Worker[T]) UniqID() any {
	return p.Key
}

func (p *Worker[T]) worker() {
	timer := time.NewTimer(p.autoCommitDuration)
	defer timer.Stop()

	work := make([]*Task[T], 0, p.batchSize)
	defer func() {
		for s, i := len(work), 0; s > 0 && i < len(work); i++ {
			if work[i].err == nil {
				work[i].signal(ErrWorkerShutdown)
			}
		}
	}()

	for {
		select {
		case v := <-p.taskC:
			if v.ctx.Err() != nil {
				v.signal(ErrWorkerContextCancel)
				continue
			}
			work = append(work, v)
			if len(work) >= p.batchSize {
				timer.Reset(p.autoCommitDuration)
				data := make([]*Task[T], p.batchSize)
				copy(data, work)
				work = work[p.batchSize:]
				go p.work(data)
			}
		case <-timer.C:
			timer.Reset(p.autoCommitDuration)
			if len(work) > 0 {
				if len(work) >= p.batchSize {
					data := make([]*Task[T], p.batchSize)
					copy(data, work)
					work = work[p.batchSize:]
					go p.work(data)
				} else {
					data := make([]*Task[T], len(work))
					copy(data, work)
					work = make([]*Task[T], 0, p.batchSize)
					go p.work(data)
				}
			}
		case <-p.dispatch.exitC:
			p.shutdown()
			return
		case <-p.exit:
			p.shutdown()
			return
		}
	}
}

func (p *Worker[T]) work(data []*Task[T]) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), p.batchTimeOut, fmt.Errorf("timeout:%s", p.batchTimeOut))
	defer cancel()

	switch p.Handle.(type) {
	case HandleBatch[T]:
		p.batch(ctx, data)
	case HandleSingle[T]:
		p.single(ctx, data)
	default:
		return
	}
}

func (p *Worker[T]) batch(ctx context.Context, data []*Task[T]) {
	retrys := 0
	payload := make([]*T, len(data))
	for i, item := range data {
		payload[i] = &item.Value
	}
OutLoop:
	for {
		select {
		case <-ctx.Done():
			p.broadcast(context.Cause(ctx), data...)

			break OutLoop
		default:
			if retrys < p.Retrys {
				if p.Handle.Handle(ctx, nil, payload) {
					p.broadcast(nil, data...)

					break OutLoop
				}
				retrys++
				time.Sleep(10 * time.Millisecond)
			} else {
				p.broadcast(ErrExceedRetrys, data...)

				break OutLoop
			}
		}
	}
}

func (p *Worker[T]) single(ctx context.Context, data []*Task[T]) error {
	for _, item := range data {
		if item == nil {
			break
		}
		select {
		case <-ctx.Done():
			item.signal(ctx.Err())
		case <-item.ctx.Done():
			item.signal(ctx.Err())
		default:
			retrys := 0
		OutLoop:
			for {
				if retrys < p.Retrys {
					if p.Handle.Handle(ctx, &item.Value, nil) {
						item.signal(nil)
						break OutLoop
					}
					retrys++
					time.Sleep(10 * time.Millisecond)
				} else {
					item.signal(ErrExceedRetrys)
					break OutLoop
				}
			}
		}
	}
	return nil
}

func (p *Worker[T]) broadcast(err error, data ...*Task[T]) {
	for _, task := range data {
		task.signal(err)
	}
}

func (p *Worker[T]) shutdown() {
	close(p.taskC)
	if p.graceDwonDuration > 0 && len(p.taskC) > 0 {
		time.Sleep(p.graceDwonDuration)
	}
}

func (p *Worker[T]) stop() {
	if !atomic.CompareAndSwapUint32(&p._closed, 0, 1) {
		return
	}
	close(p.exit)
}

func (p *Worker[T]) closed() bool {
	return atomic.LoadUint32(&p._closed) == 1
}
