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

type Handler interface {
	Handle(context.Context, *Task, []*Task) bool
}

type HandleSingle func(context.Context, *Task) bool

type HandleBatch func(context.Context, []*Task) bool

func (handle HandleSingle) Handle(ctx context.Context, task *Task, _ []*Task) bool {

	return handle(ctx, task)
}

func (handle HandleBatch) Handle(ctx context.Context, _ *Task, tasks []*Task) bool {

	return handle(ctx, tasks)
}
