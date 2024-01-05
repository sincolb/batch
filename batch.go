package batchrequests

import (
	"context"
	"time"

	_ "go.uber.org/automaxprocs"
)

var DefaultDispatch = NewDispatch[any]()

type Request[T any] struct {
	Id    string
	Value T
	Ctx   context.Context
}

func NewRequest[T any](id string, value T) *Request[T] {
	req := &Request[T]{
		Ctx:   context.Background(),
		Id:    id,
		Value: value,
	}
	return req
}

func NewRequestWithContext[T any](ctx context.Context, id string, value T) *Request[T] {
	req := &Request[T]{
		Ctx:   ctx,
		Id:    id,
		Value: value,
	}
	return req
}

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

func Submit(req *Request[any]) {
	DefaultDispatch.Submit(req)
}

func Release() {
	DefaultDispatch.Release()
}

func Exit() <-chan struct{} {
	return DefaultDispatch.Exit()
}
