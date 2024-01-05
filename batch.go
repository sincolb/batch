package batchrequests

import (
	"time"

	_ "go.uber.org/automaxprocs"
)

var DefaultDispatch = NewDispatch[any]()

func Register(uniqID any, batchSize int, AutoCommitDuration time.Duration,
	handle Handler[any], opts ...Option[any]) error {
	return DefaultDispatch.Register(uniqID, batchSize, AutoCommitDuration, handle, opts...)
}

func Unregister(uniqID any) {
	DefaultDispatch.Unregister(uniqID)
}

func UnregisterAll() {
	DefaultDispatch.UnregisterAll()
}

func Submit(key string, value any) {
	DefaultDispatch.Submit(key, value)
}

func Release() {
	DefaultDispatch.Release()
}

func Exit() <-chan struct{} {
	return DefaultDispatch.Exit()
}
