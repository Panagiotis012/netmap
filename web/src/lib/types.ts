export type DeviceStatus = "online" | "offline" | "unknown";
export type DiscoveryMethod = "scan" | "agent" | "manual";
export type ScanType = "discovery" | "port" | "full";
export type ScanStatus = "pending" | "running" | "completed" | "failed" | "cancelled";

export interface Device {
  id: string;
  hostname: string;
  ip_addresses: string[];
  mac_addresses: string[];
  os: string;
  status: DeviceStatus;
  discovery_method: DiscoveryMethod;
  first_seen_at: string;
  last_seen_at: string;
  tags: string[];
  ports?: PortResult[];
  latency_ms?: number;
  group_id?: string;
  metadata?: Record<string, unknown>;
  map_x?: number;
  map_y?: number;
  network_id?: string;
}

export interface Network {
  id: string;
  name: string;
  subnet: string;
  gateway: string; // may be empty string if not configured
}

export interface ScanJob {
  id: string;
  type: ScanType;
  target: string;
  status: ScanStatus;
  started_at: string;
  completed_at?: string;
  results?: ScanResults;
}

export interface ScanResults {
  hosts: HostResult[];
  stats: ScanStats;
}

export interface HostResult {
  ip: string;
  mac: string;
  hostname: string;
  latency_ms: number;
  ports?: PortResult[];
  os_guess?: string;
  status: string;
}

export interface PortResult {
  number: number;
  protocol: string;
  service: string;
  state: string;
}

export interface ScanStats {
  hosts_scanned: number;
  hosts_up: number;
  duration_ms: number;
}

export interface ListResult<T> {
  items: T[];
  total: number;
  page: number;
  total_pages: number;
}

export interface WSEvent {
  type: string;
  payload: unknown;
  timestamp: string;
}

export interface SystemStatus {
  version: string;
  db_path?: string;
  started_at?: string;
  gateway?: string;
  devices_online: number;
  devices_offline: number;
  devices_unknown: number;
  devices_total: number;
}

export interface NetMapConfig {
  scan_interval: string;
  scan_workers: number;
  port_ranges: string;
}
