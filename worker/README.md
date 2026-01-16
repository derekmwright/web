# go-worker

A simple, reusable, production-ready background worker module for NATS JetStream in Go.

This module allows you to easily run reliable, concurrent background workers that consume messages from JetStream streams using pull-based subscriptions. It integrates seamlessly with the `go-nats` embedded/clustered NATS module and follows the same clean, functional-options pattern as the rest of your toolkit.

## Features

- Pull-based JetStream consumer with configurable batch fetching
- Concurrent message processing (configurable workers)
- Automatic stream and consumer creation (idempotent)
- Configurable storage (Memory or File) for development vs production durability
- Graceful shutdown with context cancellation and subscription draining
- Explicit Ack/Nak handling with retry support
- Custom logger injection
- Works with embedded or external NATS servers

## Installation

```bash
go get github.com/derekmwright/web/worker
```

## Quick Start

```go
package main

import (
    "context"
    "log/slog"
    "os"

    "github.com/derekmwright/web/nats"
    "github.com/derekmwright/web/worker"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

    // Connect to NATS (embedded or external)
    nc, natsShutdown, err := nats.New()
    if err != nil {
        logger.Error("failed to connect to NATS", "error", err)
        os.Exit(1)
    }
    defer natsShutdown()

    // Create background worker
    w, err := worker.New(nc,
        worker.WithStream("ORDERS"),
        worker.WithConsumer("order-processor", "orders.process"),
        worker.WithHandler(func(ctx context.Context, msg *nats.Msg) error {
            logger.Info("processing order", "data", string(msg.Data))
            // Your business logic here
            return nil
        }),
        worker.WithConcurrency(5),
        worker.WithStorage(nats.FileStorage), // or MemoryStorage for dev
        worker.WithLogger(logger),
    )
    if err != nil {
        logger.Error("failed to create worker", "error", err)
        os.Exit(1)
    }

    // Run worker (blocks until shutdown signal)
    go w.Run()

    // Your main application (e.g. HTTP server) runs here...

    // On exit
    defer w.Shutdown()
}
```

## Options

| Option                  | Description                                      | Default                |
|-------------------------|--------------------------------------------------|------------------------|
| `WithStream(name)`      | JetStream stream name                            | required               |
| `WithConsumer(name, subject)` | Consumer name and filter subject           | required               |
| `WithDurable(name)`     | Durable consumer name (if different from consumer) | uses consumer name     |
| `WithHandler(h)`        | Message processing function                      | required               |
| `WithConcurrency(n)`    | Number of concurrent goroutines                  | 1                      |
| `WithAckWait(d)`        | Time to wait for acknowledgment                  | 30s                    |
| `WithMaxDeliver(n)`     | Maximum delivery attempts                        | 1                      |
| `WithLogger(l)`         | Custom slog logger                               | slog.Default()         |
| `WithStorage(t)`        | Stream storage type (MemoryStorage or FileStorage) | FileStorage (safe)     |

## Storage Types

- `nats.MemoryStorage` — Fast, in-memory only (lost on restart). Great for dev/testing.
- `nats.FileStorage` — Persistent on disk (survives restarts). Recommended for production.