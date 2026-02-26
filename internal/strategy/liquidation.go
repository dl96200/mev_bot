package strategy

import (
	"context"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
)

type Liquidation struct {
	cfg config.Config
}

func NewLiquidation(cfg config.Config) *Liquidation {
	return &Liquidation{cfg: cfg}
}

func (l *Liquidation) Name() string {
	return "liquidation"
}

func (l *Liquidation) Evaluate(_ context.Context, opportunity marketdata.Opportunity) (*Plan, error) {
	if opportunity.Type != "block" {
		return nil, nil
	}
	plan := buildPlan(l.cfg, l.Name(), l.cfg.MinProfitUSD+0.8, "scan hf", "liquidate account", "collect bonus")
	return plan, nil
}
