package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Chain              string
	RPCURL             string
	PrivateRelayURL    string
	MinProfitUSD       float64
	MaxGasGwei         int64
	RiskLimitUSD       float64
	FlashLoanProviders []string
	OpportunityFile    string
	MarketDataSource   string
	BlockPollInterval  time.Duration
	PoolCalls          []string
	OracleCalls        []string
	FromAddress        string
	BundleSignature    string
	AuditLogPath       string
	AlertWebhookURL    string
	MaxSlippageBps     int64
	MaxGasUSD          float64
}

func LoadFromEnv() Config {
	return Config{
		Chain:              getEnv("MEV_CHAIN", "ethereum"),
		RPCURL:             getEnv("MEV_RPC_URL", "http://localhost:8545"),
		PrivateRelayURL:    getEnv("MEV_PRIVATE_RELAY", ""),
		MinProfitUSD:       getEnvFloat("MEV_MIN_PROFIT_USD", 5.0),
		MaxGasGwei:         getEnvInt64("MEV_MAX_GAS_GWEI", 120),
		RiskLimitUSD:       getEnvFloat("MEV_RISK_LIMIT_USD", 5000),
		FlashLoanProviders: []string{"aave-v3", "uniswap-v3"},
		OpportunityFile:    getEnv("MEV_OPPORTUNITY_FILE", ""),
		MarketDataSource:   getEnv("MEV_MARKETDATA_SOURCE", "ticker"),
		BlockPollInterval:  getEnvDuration("MEV_BLOCK_POLL_INTERVAL", 5*time.Second),
		PoolCalls:          getEnvCSV("MEV_POOL_CALLS"),
		OracleCalls:        getEnvCSV("MEV_ORACLE_CALLS"),
		FromAddress:        getEnv("MEV_FROM_ADDRESS", ""),
		BundleSignature:    getEnv("MEV_BUNDLE_SIGNATURE", ""),
		AuditLogPath:       getEnv("MEV_AUDIT_LOG", ""),
		AlertWebhookURL:    getEnv("MEV_ALERT_WEBHOOK", ""),
		MaxSlippageBps:     getEnvInt64("MEV_MAX_SLIPPAGE_BPS", 50),
		MaxGasUSD:          getEnvFloat("MEV_MAX_GAS_USD", 50),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := parseFloat(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := parseInt64(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvCSV(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	var cleaned []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}
