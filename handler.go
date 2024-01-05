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
	Handle(context.Context, *Task[T], []*Task[T]) bool
}

type HandleSingle[T any] func(context.Context, *Task[T]) bool

type HandleBatch[T any] func(context.Context, []*Task[T]) bool

func (handle HandleSingle[T]) Handle(ctx context.Context, task *Task[T], _ []*Task[T]) bool {

	return handle(ctx, task)
}

func (handle HandleBatch[T]) Handle(ctx context.Context, _ *Task[T], tasks []*Task[T]) bool {

	return handle(ctx, tasks)
}
