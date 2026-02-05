package marketdata

import (
	"context"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
)

type Opportunity struct {
	Chain    string
	Type     string
	Payload  map[string]any
	Observed time.Time
}

type Service struct {
	source Source
}

func NewService(cfg config.Config, client *chain.Client) *Service {
	var source Source
	switch cfg.MarketDataSource {
	case "file":
		source = NewFileSource(cfg, cfg.OpportunityFile)
	case "rpc":
		source = NewRPCSource(cfg, client)
	default:
		source = NewTickerSource(cfg)
	}
	return &Service{source: source}
}

func (s *Service) Start(ctx context.Context) error {
	return s.source.Start(ctx)
}

func (s *Service) Refresh(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *Service) Opportunities() <-chan Opportunity {
	return s.source.Opportunities()
}
