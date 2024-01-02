package batchrequests

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"
)

type Worker struct {
	dispatch   *dispatch
	Key        any
	batchSize  int
	bufferSize int
	Retrys     int
	// Handle             any
	Handle             Handler
	taskC              chan *Task
	submitTimeOut      time.Duration
	graceDwonDuration  time.Duration
	autoCommitDuration time.Duration
	exit               chan struct{}
	_closed            uint32
}

func (p *Worker) UniqID() any {
	return p.Key
}

func (p *Worker) worker() {
	logger.Infof("listen the key = %s", p.Key)
	timer := time.NewTimer(p.autoCommitDuration)
	defer timer.Stop()

	work := make([]*Task, 0, p.batchSize)
	defer func() {
		for s, i := len(work), 0; s > 0 && i < len(work); i++ {
			if work[i].err == nil {
				work[i].signal(errors.New("canceled"))
			}
		}
	}()

	for {
		select {
		case v := <-p.taskC:
			if v.ctx.Err() != nil {
				logger.Debugln("context cancel 2, request=", v)
				v.signal(errors.New("context cancel 2"))
				continue
			}
			work = append(work, v)
			if len(work) >= p.batchSize {
				timer.Reset(p.autoCommitDuration)
				data := make([]*Task, p.batchSize)
				copy(data, work)
				work = work[p.batchSize:]
				go p.work(data)
			}
		case <-timer.C:
			timer.Reset(p.autoCommitDuration)
			if len(work) > 0 {
				if len(work) >= p.batchSize {
					data := make([]*Task, p.batchSize)
					copy(data, work)
					logger.Warnf("[%s] [timeout 1] %v\n", p.Key, data)
					work = work[p.batchSize:]
					go p.work(data)
				} else {
					data := make([]*Task, len(work))
					copy(data, work)
					logger.Warnf("[%s] [timeout 2] %v\n", p.Key, data)
					work = make([]*Task, 0, p.batchSize)
					go p.work(data)
				}
			}
		case <-p.dispatch.exitC:
			logger.Debugf("ALL received exit %s", p.Key)
			p.shutdown()
			return
		case <-p.exit:
			logger.Debugf("Self received exit %s", p.Key)
			p.shutdown()
			return
		}
	}
}

func (p *Worker) work(data []*Task) {
	ctx, cancel := context.WithTimeoutCause(context.Background(), p.submitTimeOut, fmt.Errorf("timeout:%s", p.submitTimeOut))
	defer cancel()

	switch p.Handle.(type) {
	case HandleBatch:
		p.batch(ctx, data)
	case HandleSingle:
		p.single(ctx, data)
	default:
		logger.Errorf("an error occoured: invalid Handler %T\n", p.Handle)
	}
}

func (p *Worker) batch(ctx context.Context, data []*Task) {
	retrys := 0
OutLoop:
	for {
		select {
		case <-ctx.Done():
			logger.Errorf("[%s] [error] %v [retrys] %d [Cause] %v\n", p.Key, ctx.Err(), retrys, context.Cause(ctx))
			p.broadcast(context.Cause(ctx), data...)

			break OutLoop
		default:
			if retrys < p.Retrys {
				if p.Handle.Handle(ctx, nil, data) {
					p.broadcast(nil, data...)

					break OutLoop
				}
				retrys++
				time.Sleep(10 * time.Millisecond)
			} else {
				logger.Errorf("[%s] [exceed] %v [retrys] %d\n", p.Key, data, retrys)
				p.broadcast(errors.New("exceed retrys"), data...)

				break OutLoop
			}
		}
	}
}

func (p *Worker) single(ctx context.Context, data []*Task) error {
	for _, item := range data {
		if item == nil {
			break
		}
		select {
		case <-ctx.Done():
			logger.Debugln("--------", ctx.Err().Error())
			item.signal(ctx.Err())
		case <-item.ctx.Done():
			logger.Debugln("========", item.ctx.Err().Error())
			item.signal(ctx.Err())
		default:
			retrys := 0
		OutLoop:
			for {
				if retrys < p.Retrys {
					if p.Handle.Handle(ctx, item, nil) {
						item.signal(nil)
						break OutLoop
					}
					retrys++
					time.Sleep(10 * time.Millisecond)
				} else {
					logger.Errorf("[%s] [exceed] %v [retrys] %d\n", p.Key, data, retrys)
					item.signal(errors.New("exceed retrys"))
					break OutLoop
				}
			}
		}
	}
	return nil
}

func (p *Worker) broadcast(err error, data ...*Task) {
	for _, task := range data {
		task.signal(err)
	}
}

func (p *Worker) shutdown() {
	close(p.taskC)
	if p.graceDwonDuration > 0 && len(p.taskC) > 0 {
		time.Sleep(p.graceDwonDuration)
	}
}

func (p *Worker) stop() {
	if !atomic.CompareAndSwapUint32(&p._closed, 0, 1) {

		return
	}

	close(p.exit)
}

func (p *Worker) closed() bool {

	return atomic.LoadUint32(&p._closed) == 1
}
