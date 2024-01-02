package batchrequests

import (
	"context"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	_ "go.uber.org/automaxprocs"
)

var logger = logrus.New()
var DefaultDispatch = NewDispatch()

func init() {
	logger.Formatter = new(logrus.TextFormatter) //default
	// logger.Formatter.(*logrus.TextFormatter).DisableColors = true    // remove colors
	// logger.Formatter.(*logrus.TextFormatter).DisableTimestamp = true // remove timestamp from test output
	// logger.Level = logrus.TraceLevel
	logger.Level = logrus.WarnLevel
	logger.Out = os.Stdout
}

type Request struct {
	Id    string
	Value any
	Ctx   context.Context
}

func NewRequest(id string, value any) *Request {
	req := &Request{
		Ctx:   context.Background(),
		Id:    id,
		Value: value,
	}
	return req
}

func NewRequestWithContext(ctx context.Context, id string, value any) *Request {
	req := &Request{
		Ctx:   ctx,
		Id:    id,
		Value: value,
	}
	return req
}

func Register(uniqID any, batchSize int, AutoCommitDuration time.Duration,
	handle Handler, opts ...Option) error {
	return DefaultDispatch.Register(uniqID, batchSize, AutoCommitDuration, handle, opts...)
}

func Unregister(uniqID any) {
	DefaultDispatch.Unregister(uniqID)
}

func UnregisterAll() {
	DefaultDispatch.UnregisterAll()
}

func Submit(req *Request) {
	DefaultDispatch.Submit(req)
}

func Release() {
	DefaultDispatch.Release()
}

func Exit() <-chan struct{} {
	return DefaultDispatch.Exit()
}
