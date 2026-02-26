package marketdata

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"mev_bot/internal/config"
)

type FileSource struct {
	cfg           config.Config
	path          string
	opportunities chan Opportunity
}

type fileOpportunity struct {
	Type     string         `json:"type"`
	Payload  map[string]any `json:"payload"`
	Observed *time.Time     `json:"observed,omitempty"`
	Chain    string         `json:"chain,omitempty"`
}

func NewFileSource(cfg config.Config, path string) *FileSource {
	return &FileSource{
		cfg:           cfg,
		path:          path,
		opportunities: make(chan Opportunity, 64),
	}
}

func (f *FileSource) Start(ctx context.Context) error {
	file, err := os.Open(f.path)
	if err != nil {
		return fmt.Errorf("open opportunity file: %w", err)
	}

	go func() {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			default:
			}

			var payload fileOpportunity
			if err := json.Unmarshal(scanner.Bytes(), &payload); err != nil {
				continue
			}

			observed := time.Now()
			if payload.Observed != nil {
				observed = *payload.Observed
			}

			chain := payload.Chain
			if chain == "" {
				chain = f.cfg.Chain
			}

			f.opportunities <- Opportunity{
				Chain:    chain,
				Type:     payload.Type,
				Payload:  payload.Payload,
				Observed: observed,
			}
		}
	}()

	return nil
}

func (f *FileSource) Opportunities() <-chan Opportunity {
	return f.opportunities
}
