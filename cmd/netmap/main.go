package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
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

	// Scanner — NOTE: NewNetworkProber takes (timeout, workers)
	prober := scanner.NewNetworkProber(2*time.Second, cfg.ScanWorkers)
	sc := scanner.NewScanner(prober, cfg.ScanWorkers)

	runScan := func(scanType models.ScanType, target string) {
		scanID := uuid.New().String()
		now := time.Now()
		job := &models.ScanJob{
			ID: scanID, Type: scanType, Target: target,
			Status: models.ScanRunning, StartedAt: &now, // NOTE: pointer
		}
		s.Scans.Create(context.Background(), job)

		bus.Publish(models.Event{Type: models.EventScanStarted, Payload: job, Timestamp: now})

		results, err := sc.Scan(context.Background(), target, scanType, nil)
		completed := time.Now()
		if err != nil {
			job.Status = models.ScanFailed
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

		// Upsert discovered devices (dedup: MAC → hostname → stable IP)
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
				continue // skip on transient lookup error to avoid duplicate devices
			}
			if existing == nil {
				// New device
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
				}
				s.Devices.Create(context.Background(), device)
				bus.Publish(models.Event{Type: models.EventDeviceDiscovered, Payload: device, Timestamp: now})
			} else {
				existing.LastSeenAt = now
				existing.Status = models.StatusOnline
				if host.IP != "" && !contains(existing.IPAddresses, host.IP) {
					existing.IPAddresses = append(existing.IPAddresses, host.IP)
				}
				s.Devices.Update(context.Background(), existing)
				bus.Publish(models.Event{Type: models.EventDeviceUpdated, Payload: existing, Timestamp: now})
			}
		}
	}

	// Scheduler
	sched := scanner.NewScheduler(cfg.ScanInterval, func() {
		nets, _ := s.Networks.List(context.Background())
		for _, n := range nets {
			runScan(models.ScanDiscovery, n.Subnet)
		}
	})
	sched.Start()
	defer sched.Stop()

	// HTTP server
	scanHandler := handlers.NewScanHandler(s.Scans)
	scanHandler.ScanTrigger = runScan
	router := api.NewRouter(s, hub, scanHandler)
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
