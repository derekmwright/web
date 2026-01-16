package nats

import "errors"

var ErrNotReady = errors.New("nats server not ready after ready check timeout")
