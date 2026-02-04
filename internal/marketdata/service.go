package marketdata

import (
	"context"
	"time"

	"mev_bot/internal/config"
)

type Opportunity struct {
	Chain    string
	Type     string
	Payload  map[string]any
	Observed time.Time
}

type Service struct {
	cfg           config.Config
	opportunities chan Opportunity
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:           cfg,
		opportunities: make(chan Opportunity, 64),
	}
}

func (s *Service) Start(ctx context.Context) error {
	go s.seedTicker(ctx)
	return nil
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
	return s.opportunities
}

func (s *Service) seedTicker(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			s.opportunities <- Opportunity{
				Chain:    s.cfg.Chain,
				Type:     "dex-arb",
				Payload:  map[string]any{"source": "stub"},
				Observed: t,
			}
		}
	}
}
