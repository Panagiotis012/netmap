package monitor

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/store"
)

// Runner manages periodic checks for all active monitors.
type Runner struct {
	store   store.MonitorRepo
	mu      sync.Mutex
	tickers map[string]*time.Ticker
	cancels map[string]context.CancelFunc
	cancel  context.CancelFunc
}

// NewRunner creates a new Runner backed by the given MonitorRepo.
func NewRunner(s store.MonitorRepo) *Runner {
	return &Runner{
		store:   s,
		tickers: make(map[string]*time.Ticker),
		cancels: make(map[string]context.CancelFunc),
	}
}

// Start loads all active monitors and begins a ticker goroutine for each.
func (r *Runner) Start(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	monitors, err := r.store.ListActive(ctx)
	if err != nil {
		log.Printf("monitor runner: failed to list active monitors: %v", err)
		return
	}

	for i := range monitors {
		r.Add(&monitors[i])
	}

	// Keep the goroutine alive until context is cancelled.
	<-ctx.Done()
}

// Stop cancels the runner context and stops all tickers.
func (r *Runner) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for id, ticker := range r.tickers {
		ticker.Stop()
		if cancel, ok := r.cancels[id]; ok {
			cancel()
		}
	}
	r.tickers = make(map[string]*time.Ticker)
	r.cancels = make(map[string]context.CancelFunc)
}

// Add starts a monitoring ticker for the given monitor.
func (r *Runner) Add(m *models.Monitor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Stop existing ticker if present.
	if ticker, ok := r.tickers[m.ID]; ok {
		ticker.Stop()
		if cancel, ok := r.cancels[m.ID]; ok {
			cancel()
		}
	}

	interval := time.Duration(m.Interval) * time.Second
	if interval <= 0 {
		interval = 60 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(interval)
	r.tickers[m.ID] = ticker
	r.cancels[m.ID] = cancel

	// Capture a local copy of the monitor for the goroutine.
	mCopy := *m
	go func() {
		// Run immediately on start.
		r.checkMonitor(ctx, &mCopy)

		for {
			select {
			case <-ticker.C:
				r.checkMonitor(ctx, &mCopy)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Remove stops the ticker for the monitor with the given id.
func (r *Runner) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if ticker, ok := r.tickers[id]; ok {
		ticker.Stop()
		delete(r.tickers, id)
	}
	if cancel, ok := r.cancels[id]; ok {
		cancel()
		delete(r.cancels, id)
	}
}

// checkMonitor runs the appropriate check for the monitor, persists the result,
// updates the monitor status, and sends a webhook notification if status changed.
func (r *Runner) checkMonitor(ctx context.Context, m *models.Monitor) {
	var check *models.MonitorCheck
	switch m.Type {
	case models.MonitorHTTP:
		check = checkHTTP(ctx, m)
	case models.MonitorTCP:
		check = checkTCP(ctx, m)
	case models.MonitorPing:
		check = checkPing(ctx, m)
	default:
		check = checkHTTP(ctx, m)
	}

	check.ID = uuid.New().String()
	check.MonitorID = m.ID

	// Save check result.
	if err := r.store.CreateCheck(ctx, check); err != nil {
		log.Printf("monitor runner: failed to save check for %s: %v", m.ID, err)
	}

	// Trim old checks.
	_ = r.store.DeleteOldChecks(ctx, m.ID, 1000)

	// Determine previous status to detect transitions.
	prevStatus := m.Status

	// Update monitor status.
	now := check.CheckedAt
	if err := r.store.UpdateStatus(ctx, m.ID, check.Status, now); err != nil {
		log.Printf("monitor runner: failed to update status for %s: %v", m.ID, err)
	}

	// Update local copy so the next iteration sees the right status.
	m.Status = check.Status
	m.LastCheckedAt = &now

	// Send webhook if status changed and webhook is configured.
	if m.NotifyWebhook != "" && prevStatus != check.Status &&
		(check.Status == models.MonitorStatusDown || prevStatus == models.MonitorStatusDown) {
		go sendWebhook(m, check, prevStatus)
	}
}

// checkHTTP performs an HTTP check against the monitor's URL.
func checkHTTP(ctx context.Context, m *models.Monitor) *models.MonitorCheck {
	timeout := time.Duration(m.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
		},
	}
	method := m.Method
	if method == "" {
		method = "GET"
	}

	req, err := http.NewRequestWithContext(ctx, method, m.URL, nil)
	if err != nil {
		return &models.MonitorCheck{
			Status:    models.MonitorStatusDown,
			Error:     err.Error(),
			CheckedAt: time.Now(),
		}
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		return &models.MonitorCheck{
			Status:         models.MonitorStatusDown,
			Error:          err.Error(),
			ResponseTimeMs: elapsed,
			CheckedAt:      time.Now(),
		}
	}
	defer resp.Body.Close()

	expectedStatus := m.ExpectedStatus
	if expectedStatus == 0 {
		expectedStatus = 200
	}

	if resp.StatusCode != expectedStatus {
		return &models.MonitorCheck{
			Status:         models.MonitorStatusDown,
			StatusCode:     resp.StatusCode,
			Error:          fmt.Sprintf("expected %d got %d", expectedStatus, resp.StatusCode),
			ResponseTimeMs: elapsed,
			CheckedAt:      time.Now(),
		}
	}

	// Check keyword if set.
	if m.Keyword != "" {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB max
		if !strings.Contains(string(body), m.Keyword) {
			return &models.MonitorCheck{
				Status:         models.MonitorStatusDown,
				StatusCode:     resp.StatusCode,
				Error:          "keyword not found",
				ResponseTimeMs: elapsed,
				CheckedAt:      time.Now(),
			}
		}
	}

	return &models.MonitorCheck{
		Status:         models.MonitorStatusUp,
		StatusCode:     resp.StatusCode,
		ResponseTimeMs: elapsed,
		CheckedAt:      time.Now(),
	}
}

// checkTCP performs a TCP dial check.
func checkTCP(ctx context.Context, m *models.Monitor) *models.MonitorCheck {
	timeout := time.Duration(m.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	addr := fmt.Sprintf("%s:%d", m.Host, m.Port)
	start := time.Now()
	conn, err := (&net.Dialer{Timeout: timeout}).DialContext(ctx, "tcp", addr)
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		return &models.MonitorCheck{
			Status:         models.MonitorStatusDown,
			Error:          err.Error(),
			ResponseTimeMs: elapsed,
			CheckedAt:      time.Now(),
		}
	}
	conn.Close()
	return &models.MonitorCheck{
		Status:         models.MonitorStatusUp,
		ResponseTimeMs: elapsed,
		CheckedAt:      time.Now(),
	}
}

// checkPing performs a ping check using os/exec.
func checkPing(ctx context.Context, m *models.Monitor) *models.MonitorCheck {
	timeout := time.Duration(m.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	var args []string
	switch runtime.GOOS {
	case "windows":
		args = []string{"-n", "1", "-w", fmt.Sprintf("%d", int(timeout.Milliseconds())), m.Host}
	default:
		args = []string{"-c", "1", "-W", fmt.Sprintf("%d", int(timeout.Seconds())), m.Host}
	}

	start := time.Now()
	cmdCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "ping", args...)
	err := cmd.Run()
	elapsed := int(time.Since(start).Milliseconds())

	if err != nil {
		return &models.MonitorCheck{
			Status:         models.MonitorStatusDown,
			Error:          err.Error(),
			ResponseTimeMs: elapsed,
			CheckedAt:      time.Now(),
		}
	}
	return &models.MonitorCheck{
		Status:         models.MonitorStatusUp,
		ResponseTimeMs: elapsed,
		CheckedAt:      time.Now(),
	}
}

// sendWebhook sends a Discord/Slack-compatible webhook notification.
func sendWebhook(m *models.Monitor, check *models.MonitorCheck, prevStatus models.MonitorStatus) {
	var content string
	if check.Status == models.MonitorStatusDown {
		content = fmt.Sprintf("🔴 **%s** is DOWN\nError: %s\nResponse time: %dms",
			m.Name, check.Error, check.ResponseTimeMs)
	} else if prevStatus == models.MonitorStatusDown && check.Status == models.MonitorStatusUp {
		content = fmt.Sprintf("🟢 **%s** is back UP", m.Name)
	} else {
		return
	}

	payload := map[string]string{"content": content}
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("monitor webhook: failed to marshal payload: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.NotifyWebhook, bytes.NewReader(body))
	if err != nil {
		log.Printf("monitor webhook: failed to create request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("monitor webhook: request failed: %v", err)
		return
	}
	defer resp.Body.Close()
}
