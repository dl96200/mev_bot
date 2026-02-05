package marketdata

import (
	"context"
	"fmt"
	"strings"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
)

type RPCSource struct {
	cfg           config.Config
	client        *chain.Client
	opportunities chan Opportunity
	lastBlock     string
}

func NewRPCSource(cfg config.Config, client *chain.Client) *RPCSource {
	return &RPCSource{
		cfg:           cfg,
		client:        client,
		opportunities: make(chan Opportunity, 128),
	}
}

func (r *RPCSource) Start(ctx context.Context) error {
	go r.poll(ctx)
	return nil
}

func (r *RPCSource) Opportunities() <-chan Opportunity {
	return r.opportunities
}

func (r *RPCSource) poll(ctx context.Context) {
	interval := r.cfg.BlockPollInterval
	if interval == 0 {
		interval = 5 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.pollOnce(ctx)
		}
	}
}

func (r *RPCSource) pollOnce(ctx context.Context) {
	var blockNumber string
	if err := r.client.Call(ctx, "eth_blockNumber", []interface{}{}, &blockNumber); err != nil {
		return
	}
	if blockNumber == r.lastBlock {
		return
	}

	var block chain.Block
	if err := r.client.Call(ctx, "eth_getBlockByNumber", []interface{}{blockNumber, true}, &block); err != nil {
		return
	}

	r.lastBlock = block.Number
	r.opportunities <- Opportunity{
		Chain:    r.cfg.Chain,
		Type:     "block",
		Payload:  map[string]any{"number": block.Number, "hash": block.Hash},
		Observed: time.Now(),
	}

	for _, tx := range block.Transactions {
		r.opportunities <- Opportunity{
			Chain:    r.cfg.Chain,
			Type:     "mempool-tx",
			Payload:  map[string]any{"hash": tx.Hash, "from": tx.From, "to": tx.To, "value": tx.Value},
			Observed: time.Now(),
		}
	}

	r.pollContracts(ctx, r.cfg.PoolCalls, "dex-pool")
	r.pollContracts(ctx, r.cfg.OracleCalls, "oracle")
}

func (r *RPCSource) pollContracts(ctx context.Context, calls []string, typ string) {
	for _, entry := range calls {
		parts := strings.Split(entry, ":")
		if len(parts) != 2 {
			continue
		}
		to := parts[0]
		data := parts[1]
		var result string
		err := r.client.Call(ctx, "eth_call", []interface{}{map[string]string{"to": to, "data": data}, "latest"}, &result)
		if err != nil {
			continue
		}
		r.opportunities <- Opportunity{
			Chain:    r.cfg.Chain,
			Type:     typ,
			Payload:  map[string]any{"to": to, "data": data, "result": result},
			Observed: time.Now(),
		}
	}
}

func (r *RPCSource) Health(ctx context.Context) error {
	var blockNumber string
	if err := r.client.Call(ctx, "eth_blockNumber", []interface{}{}, &blockNumber); err != nil {
		return fmt.Errorf("rpc health check failed: %w", err)
	}
	return nil
}
