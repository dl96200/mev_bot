package simulator

import (
	"context"
	"fmt"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
)

type Result struct {
	Success      bool
	EstimatedGas int64
	Reason       string
	SimulatedAt  time.Time
}

type Service struct {
	cfg    config.Config
	client *chain.Client
}

func NewService(cfg config.Config, client *chain.Client) *Service {
	return &Service{cfg: cfg, client: client}
}

func (s *Service) Simulate(ctx context.Context, plan *strategy.Plan, tx *chain.SendTransactionRequest) (Result, error) {
	if tx == nil {
		return Result{Success: false, Reason: "missing tx"}, fmt.Errorf("missing tx request")
	}

	call := chain.CallRequest{
		From:    tx.From,
		To:      tx.To,
		Gas:     tx.Gas,
		Value:   tx.Value,
		Data:    tx.Data,
		MaxFee:  tx.MaxFeePerGas,
		MaxPrio: tx.MaxPriorityFeePerGas,
	}

	var result string
	err := s.client.Call(ctx, "eth_call", []interface{}{call, "latest"}, &result)
	if err != nil {
		return Result{Success: false, Reason: err.Error(), SimulatedAt: time.Now().UTC()}, nil
	}

	gasEstimate := int64(0)
	_ = s.client.Call(ctx, "eth_estimateGas", []interface{}{call}, &gasEstimate)

	return Result{
		Success:      true,
		EstimatedGas: gasEstimate,
		Reason:       "ok",
		SimulatedAt:  time.Now().UTC(),
	}, nil
}
