package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/nats-io/nats.go"
)

type Handler func(context.Context, *nats.Msg) error

type Option func(*config)

type config struct {
	streamName    string
	consumerName  string
	durableName   string
	handler       Handler
	concurrency   int
	ackWait       time.Duration
	maxDeliver    int
	filterSubject string
	deliverPolicy nats.DeliverPolicy
	replayPolicy  nats.ReplayPolicy
	log           *slog.Logger
	storage       nats.StorageType
}

func WithStream(name string) Option {
	return func(c *config) { c.streamName = name }
}

func WithConsumer(consumer, subject string) Option {
	return func(c *config) {
		c.consumerName = consumer
		c.filterSubject = subject
	}
}

func WithDurable(name string) Option {
	return func(c *config) { c.durableName = name }
}

func WithHandler(h Handler) Option {
	return func(c *config) { c.handler = h }
}

func WithConcurrency(n int) Option {
	return func(c *config) { c.concurrency = n }
}

func WithAckWait(d time.Duration) Option {
	return func(c *config) { c.ackWait = d }
}

func WithMaxDeliver(n int) Option {
	return func(c *config) { c.maxDeliver = n }
}

func WithLogger(l *slog.Logger) Option {
	return func(c *config) { c.log = l }
}

func WithStorage(storage nats.StorageType) Option {
	return func(c *config) { c.storage = storage }
}
