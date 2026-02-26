package marketdata

import (
	"context"
	"time"

	"mev_bot/internal/config"
)

type TickerSource struct {
	cfg           config.Config
	opportunities chan Opportunity
}

func NewTickerSource(cfg config.Config) *TickerSource {
	return &TickerSource{
		cfg:           cfg,
		opportunities: make(chan Opportunity, 64),
	}
}

func (t *TickerSource) Start(ctx context.Context) error {
	go t.seedTicker(ctx)
	return nil
}

func (t *TickerSource) Opportunities() <-chan Opportunity {
	return t.opportunities
}

func (t *TickerSource) seedTicker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case tick := <-ticker.C:
			t.opportunities <- Opportunity{
				Chain:    t.cfg.Chain,
				Type:     "dex-arb",
				Payload:  map[string]any{"source": "simulated"},
				Observed: tick,
			}
		}
	}
}
