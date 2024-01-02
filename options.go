package batchrequests

import "time"

type Option interface {
	apply(*Worker)
}

type optionFunc func(*Worker)

func (f optionFunc) apply(worker *Worker) {
	f(worker)
}

func WithRetrys(retrys int) Option {
	return optionFunc(func(worker *Worker) {
		worker.Retrys = retrys
	})
}

func WithSubmitTimeOut(duration time.Duration) Option {
	return optionFunc(func(worker *Worker) {
		worker.submitTimeOut = duration
	})
}

func WithAutoCommitTimeOut(duration time.Duration) Option {
	return optionFunc(func(worker *Worker) {
		worker.autoCommitDuration = duration
	})
}

func WithAGraceDwonDuration(duration time.Duration) Option {
	return optionFunc(func(worker *Worker) {
		worker.graceDwonDuration = duration
	})
}
