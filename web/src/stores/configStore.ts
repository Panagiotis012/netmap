import { create } from "zustand";
import type { NetMapConfig } from "../lib/types";

interface ConfigState extends NetMapConfig {
  fetch: () => Promise<void>;
  save: (partial: Partial<NetMapConfig>) => Promise<void>;
}

export const useConfigStore = create<ConfigState>((set) => ({
  scan_interval: "5m",
  scan_workers: 50,
  port_ranges: "22,80,443,8080,8443",

  fetch: async () => {
    const raw = await fetch("/api/v1/system/config").then((r) => r.json());
    set({
      scan_interval: raw.scan_interval,
      scan_workers: parseInt(raw.scan_workers, 10),
      port_ranges: raw.port_ranges,
    });
  },

  save: async (partial) => {
    const payload: Record<string, string> = {};
    if (partial.scan_interval !== undefined) payload.scan_interval = partial.scan_interval;
    if (partial.scan_workers !== undefined) payload.scan_workers = String(partial.scan_workers);
    if (partial.port_ranges !== undefined) payload.port_ranges = partial.port_ranges;

    const raw = await fetch("/api/v1/system/config", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    }).then((r) => {
      if (!r.ok) throw new Error("config save failed");
      return r.json();
    });
    set({
      scan_interval: raw.scan_interval,
      scan_workers: parseInt(raw.scan_workers, 10),
      port_ranges: raw.port_ranges,
    });
  },
}));
