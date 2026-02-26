package nonce

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"mev_bot/internal/chain"
)

type Manager struct {
	client *chain.Client
	mu     sync.Mutex
	cache  map[string]int64
}

func NewManager(client *chain.Client) *Manager {
	return &Manager{client: client, cache: make(map[string]int64)}
}

func (m *Manager) Next(ctx context.Context, address string) (string, error) {
	if address == "" {
		return "", fmt.Errorf("empty address")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if value, ok := m.cache[address]; ok {
		m.cache[address] = value + 1
		return toHex(value), nil
	}

	var nonceHex string
	if err := m.client.Call(ctx, "eth_getTransactionCount", []interface{}{address, "pending"}, &nonceHex); err != nil {
		return "", fmt.Errorf("fetch nonce failed: %w", err)
	}

	nonce, err := parseHexInt64(nonceHex)
	if err != nil {
		return "", fmt.Errorf("parse nonce failed: %w", err)
	}
	m.cache[address] = nonce + 1
	return toHex(nonce), nil
}

func (m *Manager) MarkUsed(address string, nonceHex string) {
	if address == "" {
		return
	}
	nonce, err := parseHexInt64(nonceHex)
	if err != nil {
		return
	}
	m.mu.Lock()
	if cached, ok := m.cache[address]; !ok || nonce >= cached {
		m.cache[address] = nonce + 1
	}
	m.mu.Unlock()
}

func parseHexInt64(v string) (int64, error) {
	if len(v) >= 2 && v[:2] == "0x" {
		if v == "0x" {
			return 0, nil
		}
		return strconv.ParseInt(v[2:], 16, 64)
	}
	return strconv.ParseInt(v, 10, 64)
}

func toHex(v int64) string {
	if v <= 0 {
		return "0x0"
	}
	return "0x" + strconv.FormatInt(v, 16)
}
