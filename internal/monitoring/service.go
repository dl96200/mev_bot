package monitoring

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"mev_bot/internal/config"
	"mev_bot/internal/marketdata"
	"mev_bot/internal/strategy"
)

type Service struct {
	cfg           config.Config
	mu            sync.Mutex
	opportunities int
	successes     int
	rejections    int
	failures      int
	lastError     map[string]error
	auditFile     *os.File
	http          *http.Client
}

type auditEntry struct {
	Type      string         `json:"type"`
	Timestamp time.Time      `json:"timestamp"`
	Plan      *strategy.Plan `json:"plan,omitempty"`
	Error     string         `json:"error,omitempty"`
}

func NewService(cfg config.Config) *Service {
	service := &Service{
		cfg:       cfg,
		lastError: make(map[string]error),
		http:      &http.Client{Timeout: 5 * time.Second},
	}
	if cfg.AuditLogPath != "" {
		file, err := os.OpenFile(cfg.AuditLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err == nil {
			service.auditFile = file
		}
	}
	return service
}

func (s *Service) StartMetricsServer(ctx context.Context) {
	if s.cfg.MetricsAddr == "" {
		return
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		snapshot := s.Snapshot()
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = fmt.Fprintf(w, "mev_opportunities_total %v\n", snapshot["opportunities"])
		_, _ = fmt.Fprintf(w, "mev_success_total %v\n", snapshot["successes"])
		_, _ = fmt.Fprintf(w, "mev_rejections_total %v\n", snapshot["rejections"])
		_, _ = fmt.Fprintf(w, "mev_failures_total %v\n", snapshot["failures"])
	})

	srv := &http.Server{Addr: s.cfg.MetricsAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()
	go func() { _ = srv.ListenAndServe() }()
}

func (s *Service) TrackOpportunity(_ marketdata.Opportunity) {
	s.mu.Lock()
	s.opportunities++
	s.mu.Unlock()
}

func (s *Service) TrackError(source string, err error) {
	if err == nil {
		return
	}
	s.mu.Lock()
	s.lastError[source] = err
	s.failures++
	s.mu.Unlock()
	s.emitAudit("error", nil, err)
	if s.cfg.AlertWebhookURL != "" {
		s.sendAlert(source, err)
	}
}

func (s *Service) TrackRejected(plan *strategy.Plan) {
	s.mu.Lock()
	s.rejections++
	s.mu.Unlock()
	s.emitAudit("rejected", plan, nil)
}

func (s *Service) TrackExecution(plan *strategy.Plan, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.failures++
		s.emitAudit("execution_failed", plan, err)
		if s.cfg.AlertWebhookURL != "" {
			s.sendAlert("execution", err)
		}
		return
	}
	s.successes++
	s.emitAudit("execution_success", plan, nil)
}

func (s *Service) Snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return map[string]any{
		"chain":         s.cfg.Chain,
		"opportunities": s.opportunities,
		"successes":     s.successes,
		"rejections":    s.rejections,
		"failures":      s.failures,
		"updated_at":    time.Now().UTC(),
	}
}

func (s *Service) ReplayAudit(path string) ([]auditEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []auditEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry auditEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			entries = append(entries, entry)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Service) emitAudit(typ string, plan *strategy.Plan, err error) {
	if s.auditFile == nil {
		return
	}
	entry := auditEntry{Type: typ, Timestamp: time.Now().UTC(), Plan: plan}
	if err != nil {
		entry.Error = err.Error()
	}
	data, marshalErr := json.Marshal(entry)
	if marshalErr != nil {
		return
	}
	_, _ = s.auditFile.Write(append(data, '\n'))
}

func (s *Service) sendAlert(source string, err error) {
	payload := map[string]string{"source": source, "message": err.Error()}
	data, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return
	}
	req, reqErr := http.NewRequest(http.MethodPost, s.cfg.AlertWebhookURL, bytes.NewReader(data))
	if reqErr != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	_, _ = s.http.Do(req)
}
