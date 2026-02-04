package config

import "os"

type Config struct {
	Chain              string
	RPCURL             string
	PrivateRelayURL    string
	MinProfitUSD       float64
	MaxGasGwei         int64
	RiskLimitUSD       float64
	FlashLoanProviders []string
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
