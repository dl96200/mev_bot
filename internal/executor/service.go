package executor

import (
	"context"
	"fmt"

	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
	"mev_bot/internal/txbuilder"
)

type Service struct {
	cfg     config.Config
	builder *txbuilder.Builder
}

func NewService(cfg config.Config, builder *txbuilder.Builder) *Service {
	return &Service{cfg: cfg, builder: builder}
}

func (s *Service) Execute(ctx context.Context, plan *strategy.Plan) error {
	_, err := s.builder.Build(plan)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (s *Service) SubmitBundle(bundle string) error {
	if bundle == "" {
		return fmt.Errorf("empty bundle")
	}
	return nil
}
