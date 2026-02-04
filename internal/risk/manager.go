package risk

import (
	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
)

type Manager struct {
	cfg config.Config
}

func NewManager(cfg config.Config) *Manager {
	return &Manager{cfg: cfg}
}

func (m *Manager) Approve(plan *strategy.Plan) bool {
	if plan == nil {
		return false
	}
	return plan.ExpectedPNL >= m.cfg.MinProfitUSD
}
