package batchrequests

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

func WithSubmitTimeOut[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.submitTimeOut = duration
	})
}

func WithAutoCommitTimeOut[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.autoCommitDuration = duration
	})
}

func WithAGraceDwonDuration[T any](duration time.Duration) Option[T] {
	return optionFunc[T](func(worker *Worker[T]) {
		worker.graceDwonDuration = duration
	})
}
