package batchrequests

import "context"

// type workerHandle interface {
// 	*Task | []*Task
// }

// type Handler[T workerHandle] interface {
// 	Handle(context.Context, T) bool
// }

// type Handle[T workerHandle] func(context.Context, T) bool

// func (h Handle[T]) Handle(ctx context.Context, data T) bool {
// 	return h(ctx, data)
// }

type Handler[T any] interface {
	Handle(context.Context, *T, []*T) bool
}

type HandleSingle[T any] func(context.Context, *T) bool

type HandleBatch[T any] func(context.Context, []*T) bool

func (handle HandleSingle[T]) Handle(ctx context.Context, task *T, _ []*T) bool {

	return handle(ctx, task)
}

func (handle HandleBatch[T]) Handle(ctx context.Context, _ *T, tasks []*T) bool {

	return handle(ctx, tasks)
}
