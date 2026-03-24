# NetMap

Network mapping and device discovery tool. Scans your local network, tracks devices, shows live status on an interactive map.

## Features

- **Network scanning** — ARP + TCP ping sweep discovers live hosts; port scan finds open services
- **Live map** — Cytoscape graph with drag-to-reposition nodes, click to inspect
- **Device panel** — IP, MAC, OS, latency, open ports, tags; inline edit hostname/OS; delete
- **Scan popover** — Real-time progress bar via WebSocket; cancel in-flight scans
- **Scans history** — Full history with expandable per-host results
- **Alerts feed** — Device discovered/offline, scan started/completed events
- **Device list** — Searchable table with CSV export
- **Settings** — Manage networks, configure scan interval/workers/ports, system info
- **Auto-offline** — Devices not seen in a scan are marked offline automatically
- **Single binary** — React frontend embedded in the Go binary via `go:embed`

## Quick start

```bash
# Build frontend
cd web && npm install && npm run build && cd ..

# Build and run
go build -o netmap ./cmd/netmap
sudo ./netmap --port 8080
```

Open http://localhost:8080

> Scanning requires raw socket access — run with `sudo` or grant `CAP_NET_RAW`.

## Usage

```
./netmap [flags]

Flags:
  --port      HTTP port (default 8080)
  --db        SQLite database path (default ~/.netmap/netmap.db)
  --workers   Parallel scan workers (default 50, overridden by DB config)
  --interval  Auto-scan interval: 1m|5m|15m|1h|off (default 5m)
```

## API

```
GET  /api/v1/devices          List devices
PUT  /api/v1/devices/:id      Update device (hostname, os, tags, map_x, map_y)
DELETE /api/v1/devices/:id    Delete device

GET  /api/v1/networks         List networks
POST /api/v1/networks         Add network
PUT  /api/v1/networks/:id     Update network
DELETE /api/v1/networks/:id   Delete network

POST /api/v1/scans            Trigger scan { type, target }
GET  /api/v1/scans            List scan history
DELETE /api/v1/scans/:id      Cancel running scan

GET  /api/v1/system/config    Get config
PUT  /api/v1/system/config    Update config (scan_interval, scan_workers, port_ranges)
GET  /api/v1/system/status    System status

GET  /api/v1/ws               WebSocket (scan.progress, scan.completed, device.discovered, device.updated)
```

## Stack

- **Backend** — Go 1.22, Chi router, SQLite (mattn/go-sqlite3), gorilla/websocket
- **Frontend** — React 18, TypeScript, Vite, Zustand, Cytoscape.js, Framer Motion
