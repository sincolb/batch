package batch

import "errors"

var (
	ErrInvalidUniqid       = errors.New("invalid key")
	ErrDispatchExit        = errors.New("dispatch exit")
	ErrContextCancel       = errors.New("context cancel")
	ErrWorkerExit          = errors.New("worker exit")
	ErrWorkerShutdown      = errors.New("worker shutdown")
	ErrWorkerContextCancel = errors.New("worker context cancel")
	ErrExceedRetrys        = errors.New("exceed retrys")
)
