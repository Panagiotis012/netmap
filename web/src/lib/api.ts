import type { Device, Network, ScanJob, ScanType, SystemStatus, ListResult } from "./types";

const BASE = "/api/v1";

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error: string }).error || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  devices: {
    list: (params?: Record<string, string>) =>
      request<ListResult<Device>>(`/devices?${new URLSearchParams(params)}`),
    get: (id: string) => request<Device>(`/devices/${id}`),
    create: (data: Partial<Device>) =>
      request<Device>("/devices", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: Partial<Device>) =>
      request<Device>(`/devices/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/devices/${id}`, { method: "DELETE" }),
  },
  networks: {
    list: () => request<Network[]>("/networks"),
    create: (data: Partial<Network>) =>
      request<Network>("/networks", { method: "POST", body: JSON.stringify(data) }),
    update: (id: string, data: Partial<Network>) =>
      request<Network>(`/networks/${id}`, { method: "PUT", body: JSON.stringify(data) }),
    delete: (id: string) => request<void>(`/networks/${id}`, { method: "DELETE" }),
  },
  scans: {
    list: () => request<ListResult<ScanJob>>("/scans"),
    get: (id: string) => request<ScanJob>(`/scans/${id}`),
    trigger: (type: ScanType, target: string) =>
      request<{ id: string; status: string }>("/scans", {
        method: "POST",
        body: JSON.stringify({ type, target }),
      }),
    cancel: (id: string) => request<void>(`/scans/${id}`, { method: "DELETE" }),
  },
  system: {
    status: () => request<SystemStatus>("/system/status"),
  },
};
