package worker

import "errors"

var (
	ErrInvalidHandler     = errors.New("handler is nil")
	ErrStreamNotFound     = errors.New("stream not found")
	ErrConsumerNotFound   = errors.New("consumer not found")
	ErrStreamNameRequired = errors.New("stream name required")
	ErrSubjectRequired    = errors.New("subject required")
)
