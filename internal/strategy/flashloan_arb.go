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
	if opportunity.Type != "flashloan-arb" {
		return nil, nil
	}

	plan := buildPlan(f.cfg, f.Name(), f.cfg.MinProfitUSD+2, "request flashloan", "execute swaps", "repay loan")
	plan.TargetAddress = "0x0000000000000000000000000000000000000000"
	plan.Calldata = "0x"
	return plan, nil
}
