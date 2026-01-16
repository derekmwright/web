package nats

import (
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

func New(opts ...Option) (*nats.Conn, func(), error) {
	cfg := &config{
		readyTimeout: 5 * time.Second,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	var nc *nats.Conn
	var shutdown func()

	if cfg.externalURL != "" {
		var err error
		nc, err = nats.Connect(cfg.externalURL, cfg.connectOpts...)
		if err != nil {
			return nil, nil, err
		}
		shutdown = func() { nc.Close() }
		return nc, shutdown, nil
	}

	serverOpts := cfg.serverOpts
	if serverOpts == nil {
		serverOpts = &server.Options{
			Host:  "127.0.0.1",
			Port:  -1,
			NoLog: true,
		}
	}

	s, err := server.NewServer(serverOpts)
	if err != nil {
		return nil, nil, err
	}

	go s.Start()

	if !s.ReadyForConnections(cfg.readyTimeout) {
		s.Shutdown()
		return nil, nil, ErrNotReady
	}

	nc, err = nats.Connect(
		s.ClientURL(),
		append(cfg.connectOpts, nats.InProcessServer(s), nats.Name("embedded-client"))...,
	)
	if err != nil {
		s.Shutdown()
		return nil, nil, err
	}

	shutdown = func() {
		nc.Close()
		s.Shutdown()
	}

	return nc, shutdown, nil
}
