package auth0

import "errors"

var ErrNilLogger = errors.New("logger cannot be nil")
var ErrNilSessions = errors.New("sessions cannot be nil")
