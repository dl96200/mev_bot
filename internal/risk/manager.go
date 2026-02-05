package risk

import (
	"strings"

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
	if plan.ExpectedPNL < m.cfg.MinProfitUSD {
		return false
	}
	if plan.SlippageBps > m.cfg.MaxSlippageBps {
		return false
	}
	if plan.GasLimit > 0 && m.cfg.MaxGasUSD > 0 && plan.EstimatedGasUSD > m.cfg.MaxGasUSD {
		return false
	}
	if strings.EqualFold(plan.SimulationReason, "reverted") {
		return false
	}
	if plan.PriceImpactBps > m.cfg.MaxSlippageBps {
		return false
	}
	return true
}
