// Package yakit contains the first, read-only Yakit connectivity adapter.
// The endpoint must be copied from Yakit's documented local automation interface;
// this package intentionally does not guess a port or API path.
package yakit

import (
	"context"

	"github.com/Autumn-27/artex/integrations"
)

type Config struct {
	HealthURL string
	Client    *http.Client
}

type Adapter struct {
	probe integrations.HTTPProbe
}

func New(cfg Config) *Adapter {
	return &Adapter{probe: integrations.HTTPProbe{Endpoint: cfg.HealthURL, Client: cfg.Client}}
}

func (a *Adapter) ID() string   { return "yakit" }
func (a *Adapter) Name() string { return "Yakit" }

func (a *Adapter) CheckHealth(ctx context.Context) integrations.Health {
	return a.probe.Check(ctx, a.ID(), a.Name())
}
