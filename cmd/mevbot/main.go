package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mev_bot/internal/chain"
	"mev_bot/internal/config"
	"mev_bot/internal/engine"
	"mev_bot/internal/executor"
	"mev_bot/internal/marketdata"
	"mev_bot/internal/monitoring"
	"mev_bot/internal/nonce"
	"mev_bot/internal/risk"
	"mev_bot/internal/simulator"
	"mev_bot/internal/strategy"
	"mev_bot/internal/txbuilder"
)

func main() {
	logger := log.New(os.Stdout, "mev-bot ", log.LstdFlags|log.Lmicroseconds)

	cfg := config.LoadFromEnv()
	logger.Printf("starting mev bot chain=%s", cfg.Chain)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	rpcClient := chain.NewClient(cfg.RPCURL)
	nonceManager := nonce.NewManager(rpcClient)
	md := marketdata.NewService(cfg, rpcClient)
	builder := txbuilder.NewBuilder(cfg)
	exec := executor.NewService(cfg, builder, rpcClient, nonceManager)
	riskManager := risk.NewManager(cfg)
	monitor := monitoring.NewService(cfg)
	sim := simulator.NewService(cfg, rpcClient)

	strategies := []strategy.Strategy{
		strategy.NewDexArb(cfg),
		strategy.NewFlashLoanArb(cfg),
		strategy.NewLiquidation(cfg),
	}

	bot := engine.NewBot(cfg, md, exec, riskManager, monitor, sim, strategies)

	go func() {
		<-signalChan
		logger.Println("shutdown signal received")
		cancel()
	}()

	if err := bot.Run(ctx); err != nil {
		logger.Printf("bot stopped with error: %v", err)
		time.Sleep(2 * time.Second)
		os.Exit(1)
	}
}
