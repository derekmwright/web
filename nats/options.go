package nats

import (
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

type Option func(*config)

type config struct {
	serverOpts      *server.Options
	connectOpts     []nats.Option
	externalURL     string
	readyTimeout    time.Duration
	shutdownTimeout time.Duration
}

func WithExternalURL(url string) Option {
	return func(c *config) { c.externalURL = url }
}

func WithServerOpts(opts *server.Options) Option {
	return func(c *config) { c.serverOpts = opts }
}

func WithConnectOpts(opts ...nats.Option) Option {
	return func(c *config) { c.connectOpts = append(c.connectOpts, opts...) }
}

func WithReadyTimeout(d time.Duration) Option {
	return func(c *config) { c.readyTimeout = d }
}

func WithCluster(name string, routes []string) Option {
	return func(c *config) {
		if c.serverOpts == nil {
			c.serverOpts = &server.Options{}
		}
		c.serverOpts.Cluster = server.ClusterOpts{
			Name: name,
		}
		c.serverOpts.Routes = server.RoutesFromStr(strings.Join(routes, ","))
	}
}

func WithClusterListen() Option {
	return func(c *config) {
		if c.serverOpts == nil {
			c.serverOpts = &server.Options{}
		}
		c.serverOpts.Cluster.ListenStr = "0.0.0.0:6222"
	}
}

func WithJetStream(enabled bool, storeDir ...string) Option {
	return func(c *config) {
		if c.serverOpts == nil {
			c.serverOpts = &server.Options{}
		}
		c.serverOpts.JetStream = enabled
		if len(storeDir) > 0 {
			c.serverOpts.StoreDir = storeDir[0]
		}
	}
}

func WithClientName(name string) Option {
	return func(c *config) {
		c.connectOpts = append(c.connectOpts, nats.Name(name))
	}
}
