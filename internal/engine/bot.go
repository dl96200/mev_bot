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
	"mev_bot/internal/simulator"
	"mev_bot/internal/strategy"
)

type Bot struct {
	cfg        config.Config
	market     *marketdata.Service
	executor   *executor.Service
	risk       *risk.Manager
	monitor    *monitoring.Service
	simulator  *simulator.Service
	strategies []strategy.Strategy
}

func NewBot(cfg config.Config, market *marketdata.Service, exec *executor.Service, riskManager *risk.Manager, monitor *monitoring.Service, simulator *simulator.Service, strategies []strategy.Strategy) *Bot {
	return &Bot{cfg: cfg, market: market, executor: exec, risk: riskManager, monitor: monitor, simulator: simulator, strategies: strategies}
}

func (b *Bot) Run(ctx context.Context) error {
	if err := b.market.Start(ctx); err != nil {
		return err
	}
	b.monitor.StartMetricsServer(ctx)

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

				tx, err := b.executor.Build(plan)
				if err != nil {
					b.monitor.TrackError("txbuilder", err)
					continue
				}

				simResult, err := b.simulator.Simulate(ctx, plan, tx)
				if err != nil {
					b.monitor.TrackError("simulator", err)
					continue
				}
				plan.SimulationReason = simResult.Reason
				plan.PriceImpactBps = simResult.PriceImpactBps
				if simResult.EstimatedGas > 0 {
					plan.GasLimit = simResult.EstimatedGas
					plan.EstimatedGasUSD = estimateGasUSD(plan.GasLimit, b.cfg.MaxGasGwei)
				}
				if !simResult.Success {
					b.monitor.TrackRejected(plan)
					continue
				}

				if !b.risk.Approve(plan) {
					b.monitor.TrackRejected(plan)
					continue
				}

				if _, err := b.executor.Execute(ctx, plan); err != nil {
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

func estimateGasUSD(gasLimit int64, maxGasGwei int64) float64 {
	if gasLimit <= 0 || maxGasGwei <= 0 {
		return 0
	}
	return float64(gasLimit) * float64(maxGasGwei) / 1_000_000_000
}

func IsTerminal(err error) bool { return errors.Is(err, context.Canceled) }
