package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/netmap/netmap/internal/api"
	"github.com/netmap/netmap/internal/api/handlers"
	"github.com/netmap/netmap/internal/api/ws"
	"github.com/netmap/netmap/internal/core/config"
	"github.com/netmap/netmap/internal/core/eventbus"
	"github.com/netmap/netmap/internal/core/models"
	"github.com/netmap/netmap/internal/scanner"
	"github.com/netmap/netmap/internal/store"
	"github.com/netmap/netmap/internal/store/sqlite"
)

var version = "0.1.0"

//go:embed all:dist
var webFS embed.FS

func staticHandler() http.Handler {
	dist, _ := fs.Sub(webFS, "dist")
	fileServer := http.FileServer(http.FS(dist))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; if not found, serve index.html (SPA fallback)
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := dist.Open(path); err != nil {
				r.URL.Path = "/"
			}
		}
		fileServer.ServeHTTP(w, r)
	})
}

func main() {
	cfg := config.Default()
	if err := cfg.ParseFlags(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Database
	db, err := sqlite.Open(cfg.DBPath())
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	s := &store.Store{
		Devices:  sqlite.NewDeviceRepo(db),
		Networks: sqlite.NewNetworkRepo(db),
		Scans:    sqlite.NewScanRepo(db),
		Alerts:   sqlite.NewAlertRepo(db),
		Sessions: sqlite.NewSessionRepo(db),
	}

	// Event bus
	bus := eventbus.New()
	defer bus.Close()

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	defer hub.Stop()

	// Bridge: event bus -> WebSocket
	for _, eventType := range []models.EventType{
		models.EventDeviceDiscovered, models.EventDeviceUpdated, models.EventDeviceLost,
		models.EventScanStarted, models.EventScanProgress, models.EventScanCompleted,
	} {
		et := eventType
		bus.Subscribe(et, func(e models.Event) {
			hub.Broadcast(e)
		})
	}

	configRepo := sqlite.NewConfigRepo(db)

	// Override defaults with DB values
	if v := configRepo.Get(context.Background(), "scan_workers"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			cfg.ScanWorkers = n
		}
	}
	if v := configRepo.Get(context.Background(), "scan_interval"); v != "" {
		if d, err := parseScanInterval(v); err == nil {
			cfg.ScanInterval = d
		}
	}

	// Port ranges from config (fall back to CommonPorts if not set / invalid)
	portRanges := scanner.CommonPorts
	if v := configRepo.Get(context.Background(), "port_ranges"); v != "" {
		if parsed := parsePortRanges(v); len(parsed) > 0 {
			portRanges = parsed
		}
	}

	// Scanner — NOTE: NewNetworkProber takes (timeout, workers)
	prober := scanner.NewNetworkProber(2*time.Second, cfg.ScanWorkers)
	sc := scanner.NewScanner(prober, cfg.ScanWorkers, portRanges)

	runScan := func(ctx context.Context, scanID string, scanType models.ScanType, target string) {
		job, err := s.Scans.GetByID(context.Background(), scanID)
		if err != nil {
			return
		}
		now := time.Now()
		bus.Publish(models.Event{Type: models.EventScanStarted, Payload: job, Timestamp: now})

		scanStart := now
		progressCb := func(scanned, total, found int) {
			pct := 0
			etaSecs := 0
			if total > 0 {
				pct = scanned * 100 / total
				if scanned > 0 {
					elapsed := time.Since(scanStart).Seconds()
					rate := elapsed / float64(scanned)
					remaining := float64(total-scanned) * rate
					etaSecs = int(remaining)
				}
			}
			bus.Publish(models.Event{
				Type: models.EventScanProgress,
				Payload: models.ScanProgressPayload{
					ScanID:       scanID,
					HostsScanned: scanned,
					HostsTotal:   total,
					HostsFound:   found,
					Percent:      pct,
					EtaSeconds:   etaSecs,
				},
				Timestamp: time.Now(),
			})
		}

		results, err := sc.Scan(ctx, target, scanType, progressCb)
		completed := time.Now()

		if err != nil {
			status := models.ScanFailed
			if ctx.Err() != nil {
				status = models.ScanCancelled
			}
			job.Status = status
			job.CompletedAt = &completed
			s.Scans.Update(context.Background(), job)
			return
		}

		resultsJSON, _ := json.Marshal(results)
		job.Status = models.ScanCompleted
		job.CompletedAt = &completed
		job.Results = json.RawMessage(resultsJSON)
		s.Scans.Update(context.Background(), job)

		bus.Publish(models.Event{Type: models.EventScanCompleted, Payload: results, Timestamp: completed})

		for _, host := range results.Hosts {
			var existing *models.Device
			var findErr error
			if host.MAC != "" {
				existing, findErr = s.Devices.GetByMAC(context.Background(), host.MAC)
			}
			if existing == nil && host.Hostname != "" {
				existing, findErr = s.Devices.GetByHostname(context.Background(), host.Hostname)
			}
			if existing == nil && host.IP != "" {
				existing, findErr = s.Devices.GetByIP(context.Background(), host.IP)
			}
			if findErr != nil {
				continue
			}
			if existing == nil {
				device := &models.Device{
					ID:              uuid.New().String(),
					Hostname:        host.Hostname,
					IPAddresses:     []string{host.IP},
					MACAddresses:    []string{host.MAC},
					OS:              host.OSGuess,
					Status:          models.StatusOnline,
					DiscoveryMethod: models.DiscoveryScan,
					FirstSeenAt:     now,
					LastSeenAt:      now,
					Tags:            []string{},
					Ports:           host.Ports,
					LatencyMs:       host.LatencyMs,
				}
				s.Devices.Create(context.Background(), device)
				bus.Publish(models.Event{Type: models.EventDeviceDiscovered, Payload: device, Timestamp: now})
			} else {
				existing.LastSeenAt = now
				existing.Status = models.StatusOnline
				if host.IP != "" && !contains(existing.IPAddresses, host.IP) {
					existing.IPAddresses = append(existing.IPAddresses, host.IP)
				}
				// MAC fix: only append if not already present
				if host.MAC != "" && !contains(existing.MACAddresses, host.MAC) {
					existing.MACAddresses = append(existing.MACAddresses, host.MAC)
				}
				if len(host.Ports) > 0 {
					existing.Ports = host.Ports
				}
				existing.LatencyMs = host.LatencyMs
				s.Devices.Update(context.Background(), existing)
				bus.Publish(models.Event{Type: models.EventDeviceUpdated, Payload: existing, Timestamp: now})
			}
		}

		// Mark devices in this subnet that didn't respond as offline.
		markOfflineInSubnet(context.Background(), s, bus, target, results.Hosts, now)
	}

	// Scheduler
	sched := scanner.NewScheduler(cfg.ScanInterval, func() {
		nets, _ := s.Networks.List(context.Background())
		for _, n := range nets {
			scanID := uuid.New().String()
			now := time.Now()
			job := &models.ScanJob{
				ID: scanID, Type: models.ScanDiscovery, Target: n.Subnet,
				Status: models.ScanRunning, StartedAt: &now,
			}
			s.Scans.Create(context.Background(), job)
			runScan(context.Background(), scanID, models.ScanDiscovery, n.Subnet)
		}
	})
	sched.Start()
	defer sched.Stop()

	// HTTP server
	scanHandler := handlers.NewScanHandler(s.Scans)
	scanHandler.ScanTrigger = runScan
	configHandler := handlers.NewConfigHandler(configRepo)
	authHandler := handlers.NewAuthHandler(configRepo, s.Sessions)
	alertHandler := handlers.NewAlertHandler(s.Alerts)
	router := api.NewRouter(s, hub, scanHandler, configHandler, authHandler, alertHandler, version)
	router.Handle("/*", staticHandler())
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		log.Printf("NetMap v%s starting on http://localhost:%d", version, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func parseScanInterval(s string) (time.Duration, error) {
	switch s {
	case "1m":
		return time.Minute, nil
	case "5m":
		return 5 * time.Minute, nil
	case "15m":
		return 15 * time.Minute, nil
	case "1h":
		return time.Hour, nil
	case "off":
		return 0, nil
	}
	return 0, fmt.Errorf("unknown interval: %s", s)
}

// markOfflineInSubnet marks devices whose IPs fall inside the scanned subnet
// as offline if they were not present in the scan results.
func markOfflineInSubnet(ctx context.Context, s *store.Store, bus *eventbus.EventBus, subnet string, found []models.HostResult, now time.Time) {
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return
	}

	// Build a set of IPs that responded.
	seenIPs := make(map[string]bool, len(found))
	for _, h := range found {
		seenIPs[h.IP] = true
	}

	// List all devices (high limit to get everything).
	result, err := s.Devices.List(ctx, models.ListParams{Limit: 10000, Page: 1})
	if err != nil {
		return
	}

	for i := range result.Items {
		d := &result.Items[i]
		if d.Status != models.StatusOnline {
			continue
		}
		// Check if any of the device's IPs fall within the subnet.
		inSubnet := false
		for _, ipStr := range d.IPAddresses {
			if ip := net.ParseIP(ipStr); ip != nil && ipNet.Contains(ip) {
				inSubnet = true
				break
			}
		}
		if !inSubnet {
			continue
		}
		// Check if any of the device's IPs were seen in the scan.
		wasSeen := false
		for _, ipStr := range d.IPAddresses {
			if seenIPs[ipStr] {
				wasSeen = true
				break
			}
		}
		if wasSeen {
			continue
		}
		// Device is in the subnet but didn't respond — mark offline.
		d.Status = models.StatusOffline
		s.Devices.Update(ctx, d)
		bus.Publish(models.Event{Type: models.EventDeviceUpdated, Payload: d, Timestamp: now})
	}
}

func parsePortRanges(s string) []int {
	parts := strings.Split(s, ",")
	var ports []int
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil && n > 0 && n <= 65535 {
			ports = append(ports, n)
		}
	}
	return ports
}
