package txbuilder

import (
	"fmt"
	"strconv"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
)

type Builder struct {
	cfg config.Config
}

func NewBuilder(cfg config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) Build(plan *strategy.Plan) (*chain.SendTransactionRequest, error) {
	if plan == nil {
		return nil, fmt.Errorf("nil plan")
	}
	if b.cfg.FromAddress == "" {
		return nil, fmt.Errorf("missing MEV_FROM_ADDRESS")
	}

	maxFee := gweiToWei(b.cfg.MaxGasGwei)
	priorityFee := gweiToWei(plan.PriorityFeeGwei)

	return &chain.SendTransactionRequest{
		From:                 b.cfg.FromAddress,
		To:                   plan.TargetAddress,
		Gas:                  toHex(plan.GasLimit),
		Value:                toHex(plan.ValueWei),
		Data:                 plan.Calldata,
		MaxFeePerGas:         maxFee,
		MaxPriorityFeePerGas: priorityFee,
		Type:                 "0x2",
		ChainID:              chainIDForChain(b.cfg.Chain),
	}, nil
}

func chainIDForChain(chain string) string {
	switch chain {
	case "bsc":
		return "0x38"
	case "ethereum":
		return "0x1"
	default:
		return "0x1"
	}
}

func gweiToWei(gwei int64) string {
	if gwei <= 0 {
		return "0x0"
	}
	return "0x" + strconv.FormatInt(gwei*1_000_000_000, 16)
}

func toHex(value int64) string {
	if value <= 0 {
		return "0x0"
	}
	return "0x" + strconv.FormatInt(value, 16)
}
