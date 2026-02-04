package strategy

import (
	"context"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
)

type DexArb struct {
	cfg config.Config
}

func NewDexArb(cfg config.Config) *DexArb {
	return &DexArb{cfg: cfg}
}

func (d *DexArb) Name() string {
	return "dex-arbitrage"
}

func (d *DexArb) Evaluate(_ context.Context, opportunity marketdata.Opportunity) (*Plan, error) {
	if opportunity.Type != "dex-arb" {
		return nil, nil
	}

	return buildPlan(d.cfg, d.Name(), d.cfg.MinProfitUSD+1, "quote path", "swap v2->v3"), nil
}
