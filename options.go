package batch

import "time"

type Option[T any] interface {
	apply(*Worker[T])
}

type optionFunc[T any] func(*Worker[T])

func (f optionFunc[T]) apply(worker *Worker[T]) {
	f(worker)
}

func WithRetrys[T any](retrys int) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.Retrys = retrys
	})
}

func WithBatchTimeOut[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.batchTimeOut = duration
	})
}

func WithGraceDwonDuration[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.graceDwonDuration = duration
	})
}

func WithAutoCommitTimeOut[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.autoCommitDuration = duration
	})
}
