package strategy

import (
	"context"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
)

type Plan struct {
	Strategy         string
	Chain            string
	ExpectedPNL      float64
	MaxGasGwei       int64
	Actions          []string
	CreatedAt        time.Time
	TargetAddress    string
	Calldata         string
	ValueWei         int64
	GasLimit         int64
	PriorityFeeGwei  int64
	SlippageBps      int64
	BundleTxs        []string
	TargetBlock      string
	EstimatedGasUSD  float64
	SimulationReason string
	PriceImpactBps   int64
}

type Strategy interface {
	Name() string
	Evaluate(ctx context.Context, opportunity marketdata.Opportunity) (*Plan, error)
}

func buildPlan(cfg config.Config, name string, pnl float64, actions ...string) *Plan {
	return &Plan{
		Strategy:        name,
		Chain:           cfg.Chain,
		ExpectedPNL:     pnl,
		MaxGasGwei:      cfg.MaxGasGwei,
		Actions:         actions,
		CreatedAt:       time.Now(),
		GasLimit:        350000,
		PriorityFeeGwei: 1,
		SlippageBps:     cfg.MaxSlippageBps,
		TargetAddress:   "0x0000000000000000000000000000000000000000",
		Calldata:        "0x",
	}
}
