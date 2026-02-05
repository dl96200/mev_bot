package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
	"mev_bot/internal/txbuilder"
)

type Service struct {
	cfg     config.Config
	builder *txbuilder.Builder
	client  *chain.Client
	http    *http.Client
}

func NewService(cfg config.Config, builder *txbuilder.Builder, client *chain.Client) *Service {
	return &Service{
		cfg:     cfg,
		builder: builder,
		client:  client,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *Service) Build(plan *strategy.Plan) (*chain.SendTransactionRequest, error) {
	return s.builder.Build(plan)
}

func (s *Service) Execute(ctx context.Context, plan *strategy.Plan) (*chain.SendTransactionRequest, error) {
	tx, err := s.builder.Build(plan)
	if err != nil {
		return nil, err
	}

	if s.cfg.PrivateRelayURL != "" && len(plan.BundleTxs) > 0 {
		err := s.SubmitBundle(ctx, plan.BundleTxs, plan.TargetBlock)
		if err != nil {
			return tx, err
		}
		return tx, nil
	}

	if err := s.client.Call(ctx, "eth_sendTransaction", []interface{}{tx}, nil); err != nil {
		return tx, err
	}
	return tx, nil
}

func (s *Service) SubmitBundle(ctx context.Context, signedTxs []string, targetBlock string) error {
	if s.cfg.PrivateRelayURL == "" {
		return fmt.Errorf("missing private relay url")
	}
	if len(signedTxs) == 0 {
		return fmt.Errorf("empty bundle")
	}

	payload := chain.BundleRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_sendBundle",
		Params: []chain.BundleItem{
			{
				Txs:         signedTxs,
				BlockNumber: targetBlock,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.PrivateRelayURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create bundle request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.cfg.BundleSignature != "" {
		req.Header.Set("X-Flashbots-Signature", s.cfg.BundleSignature)
	}

	resp, err := s.http.Do(req)
	if err != nil {
		return fmt.Errorf("bundle request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("bundle rejected: %s", resp.Status)
	}

	return nil
}
