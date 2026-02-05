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
	if opportunity.Type != "liquidation" {
		return nil, nil
	}

	plan := buildPlan(l.cfg, l.Name(), l.cfg.MinProfitUSD+3, "fetch account", "repay debt", "seize collateral")
	plan.TargetAddress = "0x0000000000000000000000000000000000000000"
	plan.Calldata = "0x"
	return plan, nil
}
