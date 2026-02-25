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
	"mev_bot/internal/nonce"
	"mev_bot/internal/strategy"
	"mev_bot/internal/txbuilder"
)

type Service struct {
	cfg          config.Config
	builder      *txbuilder.Builder
	client       *chain.Client
	nonceManager *nonce.Manager
	http         *http.Client
}

func NewService(cfg config.Config, builder *txbuilder.Builder, client *chain.Client, nonceManager *nonce.Manager) *Service {
	return &Service{
		cfg:          cfg,
		builder:      builder,
		client:       client,
		nonceManager: nonceManager,
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

	if tx.Nonce == "" && s.nonceManager != nil {
		nonceHex, nonceErr := s.nonceManager.Next(ctx, tx.From)
		if nonceErr != nil {
			return tx, nonceErr
		}
		tx.Nonce = nonceHex
	}

	if s.cfg.PrivateRelayURL != "" && len(plan.BundleTxs) > 0 {
		err := s.SubmitBundle(ctx, plan.BundleTxs, plan.TargetBlock)
		if err != nil {
			return tx, err
		}
		return tx, nil
	}

	raw, err := s.signTransaction(ctx, tx)
	if err != nil {
		return tx, err
	}
	if err := s.sendWithRetry(ctx, raw, plan); err != nil {
		return tx, err
	}
	if s.nonceManager != nil {
		s.nonceManager.MarkUsed(tx.From, tx.Nonce)
	}
	return tx, nil
}

func (s *Service) signTransaction(ctx context.Context, tx *chain.SendTransactionRequest) (string, error) {
	if tx == nil {
		return "", fmt.Errorf("nil tx")
	}
	var signed chain.SignedTransaction
	if err := s.client.Call(ctx, "eth_signTransaction", []interface{}{tx}, &signed); err != nil {
		return "", fmt.Errorf("sign transaction failed: %w", err)
	}
	if signed.Raw == "" {
		return "", fmt.Errorf("empty signed raw tx")
	}
	return signed.Raw, nil
}

func (s *Service) sendWithRetry(ctx context.Context, raw string, plan *strategy.Plan) error {
	if raw == "" {
		return fmt.Errorf("empty raw tx")
	}

	maxAttempts := 3
	backoff := 500 * time.Millisecond
	lastErr := error(nil)
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := s.client.Call(ctx, "eth_sendRawTransaction", []interface{}{raw}, nil); err == nil {
			return nil
		} else {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}
		backoff *= 2
		if plan != nil {
			plan.PriorityFeeGwei++
		}
	}
	return fmt.Errorf("send raw transaction failed after retries: %w", lastErr)
}

func (s *Service) SubmitBundle(ctx context.Context, signedTxs []string, targetBlock string) error {
	if s.cfg.PrivateRelayURL == "" {
		return fmt.Errorf("missing private relay url")
	}
	if len(signedTxs) == 0 {
		return fmt.Errorf("empty bundle")
	}
	if targetBlock == "" {
		var latest string
		if err := s.client.Call(ctx, "eth_blockNumber", []interface{}{}, &latest); err == nil {
			targetBlock = latest
		}
	}

	payload := chain.BundleRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "eth_sendBundle",
		Params: []chain.BundleItem{{
			Txs:         signedTxs,
			BlockNumber: targetBlock,
		}},
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
