package txbuilder

import (
	"fmt"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/strategy"
)

type Builder struct {
	cfg config.Config
}

func NewBuilder(cfg config.Config) *Builder {
	return &Builder{cfg: cfg}
}

func (b *Builder) Build(plan *strategy.Plan) (string, error) {
	if plan == nil {
		return "", fmt.Errorf("nil plan")
	}
	return fmt.Sprintf("tx:%s:%d", plan.Strategy, time.Now().Unix()), nil
}
