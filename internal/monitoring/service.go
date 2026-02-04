package monitoring

import (
	"sync"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
	"mev_bot/internal/strategy"
)

type Service struct {
	cfg           config.Config
	mu            sync.Mutex
	opportunities int
	lastError     map[string]error
}

func NewService(cfg config.Config) *Service {
	return &Service{
		cfg:       cfg,
		lastError: make(map[string]error),
	}
}

func (s *Service) TrackOpportunity(_ marketdata.Opportunity) {
	s.mu.Lock()
	s.opportunities++
	s.mu.Unlock()
}

func (s *Service) TrackError(source string, err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.lastError[source] = err
	s.mu.Unlock()
}

func (s *Service) TrackRejected(_ *strategy.Plan) {
	// placeholder for metrics
}

func (s *Service) TrackExecution(_ *strategy.Plan, _ error) {
	// placeholder for metrics
}

func (s *Service) Snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	return map[string]any{
		"chain":         s.cfg.Chain,
		"opportunities": s.opportunities,
		"updated_at":    time.Now().UTC(),
	}
}
