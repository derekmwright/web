package worker

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
)

type Worker struct {
	js          nats.JetStreamContext
	sub         *nats.Subscription
	handler     Handler
	log         *slog.Logger
	concurrency int
}

func New(nc *nats.Conn, opts ...Option) (*Worker, error) {
	cfg := &config{
		concurrency:   1,
		ackWait:       30 * time.Second,
		maxDeliver:    1,
		deliverPolicy: nats.DeliverAllPolicy,
		replayPolicy:  nats.ReplayInstantPolicy,
		log:           slog.Default(),
		storage:       nats.FileStorage,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.handler == nil {
		return nil, ErrInvalidHandler
	}

	if cfg.streamName == "" {
		return nil, ErrStreamNameRequired
	}

	if cfg.filterSubject == "" {
		return nil, ErrSubjectRequired
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, err
	}

	w := &Worker{
		js:          js,
		log:         cfg.log,
		handler:     cfg.handler,
		concurrency: cfg.concurrency,
	}

	_, err = js.StreamInfo(cfg.streamName)
	if err != nil {
		if errors.Is(err, nats.ErrStreamNotFound) {
			w.log.Info("creating stream", "name", cfg.streamName)

			_, err = js.AddStream(&nats.StreamConfig{
				Name:     cfg.streamName,
				Subjects: []string{cfg.filterSubject + ".>"},
				Storage:  cfg.storage,
			})
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	if cfg.consumerName == "" {
		cfg.consumerName = cfg.durableName
	}

	consumerCfg := &nats.ConsumerConfig{
		Durable:       cfg.durableName,
		AckPolicy:     nats.AckExplicitPolicy,
		AckWait:       cfg.ackWait,
		MaxDeliver:    cfg.maxDeliver,
		FilterSubject: cfg.filterSubject,
		DeliverPolicy: cfg.deliverPolicy,
		ReplayPolicy:  cfg.replayPolicy,
	}

	_, err = js.AddConsumer(cfg.streamName, consumerCfg)
	if err != nil {
		return nil, err
	}

	w.sub, err = js.PullSubscribe(cfg.filterSubject, cfg.consumerName, nats.BindStream(cfg.streamName))
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Worker) Run() {
	w.log.Info("background worker started")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		w.log.Info("shutting down background worker")
		cancel()
	}()

	var wg sync.WaitGroup
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			w.processMessages(ctx, id)
		}(i)
	}

	wg.Wait()
	w.log.Info("background worker stopped")
}

func (w *Worker) processMessages(ctx context.Context, workerID int) {
	batchSize := 10
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msgs, err := w.sub.Fetch(
				batchSize,
				nats.Context(ctx),
				nats.MaxWait(5*time.Second),
			)
			if err != nil {
				if errors.Is(err, nats.ErrTimeout) {
					continue
				}
				w.log.Error("error fetching messages", "error", err)
				time.Sleep(time.Second)
				continue
			}

			for _, msg := range msgs {
				if err := w.handler(ctx, msg); err != nil {
					w.log.Error("error processing message", "error", err)
					msg.Nak()
				} else {
					msg.Ack()
				}
			}
		}
	}
}

func (w *Worker) Shutdown() {
	if w.sub != nil {
		w.sub.Drain()
	}
}
