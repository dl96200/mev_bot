package strategy

import (
	"context"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
)

type Plan struct {
	Strategy    string
	Chain       string
	ExpectedPNL float64
	MaxGasGwei  int64
	Actions     []string
	CreatedAt   time.Time
}

type Strategy interface {
	Name() string
	Evaluate(ctx context.Context, opportunity marketdata.Opportunity) (*Plan, error)
}

func buildPlan(cfg config.Config, name string, pnl float64, actions ...string) *Plan {
	return &Plan{
		Strategy:    name,
		Chain:       cfg.Chain,
		ExpectedPNL: pnl,
		MaxGasGwei:  cfg.MaxGasGwei,
		Actions:     actions,
		CreatedAt:   time.Now(),
	}
}
