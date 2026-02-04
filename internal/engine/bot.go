package engine

import (
	"context"
	"errors"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/executor"
	"mev_bot/internal/marketdata"
	"mev_bot/internal/monitoring"
	"mev_bot/internal/risk"
	"mev_bot/internal/strategy"
)

type Bot struct {
	cfg        config.Config
	market     *marketdata.Service
	executor   *executor.Service
	risk       *risk.Manager
	monitor    *monitoring.Service
	strategies []strategy.Strategy
}

func NewBot(cfg config.Config, market *marketdata.Service, exec *executor.Service, riskManager *risk.Manager, monitor *monitoring.Service, strategies []strategy.Strategy) *Bot {
	return &Bot{
		cfg:        cfg,
		market:     market,
		executor:   exec,
		risk:       riskManager,
		monitor:    monitor,
		strategies: strategies,
	}
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.market.Start(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case opportunity := <-b.market.Opportunities():
			b.monitor.TrackOpportunity(opportunity)
			for _, strat := range b.strategies {
				plan, err := strat.Evaluate(ctx, opportunity)
				if err != nil {
					b.monitor.TrackError(strat.Name(), err)
					continue
				}
				if plan == nil {
					continue
				}

				if !b.risk.Approve(plan) {
					b.monitor.TrackRejected(plan)
					continue
				}

				if err := b.executor.Execute(ctx, plan); err != nil {
					b.monitor.TrackExecution(plan, err)
					continue
				}
				b.monitor.TrackExecution(plan, nil)
			}
		case <-time.After(3 * time.Second):
			if err := b.market.Refresh(ctx); err != nil {
				b.monitor.TrackError("marketdata", err)
			}
		}
	}
}

func IsTerminal(err error) bool {
	return errors.Is(err, context.Canceled)
}
