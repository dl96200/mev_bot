package simulator

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
)

type Result struct {
	Success        bool
	EstimatedGas   int64
	Reason         string
	SimulatedAt    time.Time
	PriceImpactBps int64
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

	call := chain.CallRequest{From: tx.From, To: tx.To, Gas: tx.Gas, Value: tx.Value, Data: tx.Data, MaxFee: tx.MaxFeePerGas, MaxPrio: tx.MaxPriorityFeePerGas}
	var callResult string
	err := s.client.Call(ctx, "eth_call", []interface{}{call, "latest"}, &callResult)
	if err != nil {
		return Result{Success: false, Reason: err.Error(), SimulatedAt: time.Now().UTC()}, nil
	}
	if strings.HasPrefix(strings.ToLower(callResult), "0x08c379a0") {
		return Result{Success: false, Reason: "reverted", SimulatedAt: time.Now().UTC()}, nil
	}

	gasHex := "0x0"
	_ = s.client.Call(ctx, "eth_estimateGas", []interface{}{call}, &gasHex)
	gas, _ := parseHexInt64(gasHex)

	impact := estimatePriceImpact(plan)
	return Result{Success: true, EstimatedGas: gas, Reason: "ok", SimulatedAt: time.Now().UTC(), PriceImpactBps: impact}, nil
}

func estimatePriceImpact(plan *strategy.Plan) int64 {
	if plan == nil {
		return 0
	}
	if plan.ExpectedPNL <= 0 {
		return 100
	}
	if plan.ExpectedPNL < 10 {
		return 40
	}
	return 10
}
