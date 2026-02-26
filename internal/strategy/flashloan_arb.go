package strategy

import (
	"context"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
)

type FlashLoanArb struct {
	cfg config.Config
}

func NewFlashLoanArb(cfg config.Config) *FlashLoanArb {
	return &FlashLoanArb{cfg: cfg}
}

func (f *FlashLoanArb) Name() string {
	return "flashloan-arbitrage"
}

func (f *FlashLoanArb) Evaluate(_ context.Context, opportunity marketdata.Opportunity) (*Plan, error) {
	if opportunity.Type != "oracle" {
		return nil, nil
	}
	plan := buildPlan(f.cfg, f.Name(), f.cfg.MinProfitUSD+2.0, "borrow flashloan", "swap multi-hop", "repay + profit")
	return plan, nil
}
