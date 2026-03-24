import { create } from "zustand";
import { api } from "../lib/api";
import type { ScanJob } from "../lib/types";

export interface ActiveScan {
  id: string;
  target: string;
  hostsScanned: number;
  hostsTotal: number;
  hostsFound: number;
  percent: number;
  etaSeconds: number;
  newDevicesCount: number;
}

interface ScanState {
  scans: ScanJob[];
  scanning: boolean;
  activeScan: ActiveScan | null;
  popoverOpen: boolean;
  popoverMode: "progress" | "complete" | null;

  fetch: () => Promise<void>;
  triggerScan: (type: "discovery" | "port" | "full", target: string) => Promise<void>;
  startScan: (networkSubnet: string) => Promise<void>;
  cancelScan: () => Promise<void>;
  setActiveScan: (scan: Partial<ActiveScan>) => void;
  clearActiveScan: () => void;
  setPopover: (open: boolean, mode: "progress" | "complete" | null) => void;
  incrementNewDevices: () => void;
}

export const useScanStore = create<ScanState>((set, get) => ({
  scans: [],
  scanning: false,
  activeScan: null,
  popoverOpen: false,
  popoverMode: null,

  fetch: async () => {
    const result = await api.scans.list();
    set({ scans: result.items || [] });
  },

  triggerScan: async (type, target) => {
    set({ scanning: true });
    try {
      await fetch("/api/v1/scans", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type, target }),
      });
    } finally {
      setTimeout(() => set({ scanning: false }), 2000);
    }
  },

  startScan: async (networkSubnet) => {
    const data = await api.scans.trigger("discovery", networkSubnet);
    set({
      activeScan: {
        id: data.id,
        target: networkSubnet,
        hostsScanned: 0,
        hostsTotal: 0,
        hostsFound: 0,
        percent: 0,
        etaSeconds: 0,
        newDevicesCount: 0,
      },
      popoverOpen: true,
      popoverMode: "progress",
      scanning: true,
    });
  },

  cancelScan: async () => {
    const { activeScan } = get();
    if (!activeScan) return;
    await api.scans.cancel(activeScan.id);
    set({ activeScan: null, popoverOpen: false, popoverMode: null, scanning: false });
  },

  setActiveScan: (partial) => {
    set((s) => ({
      activeScan: s.activeScan ? { ...s.activeScan, ...partial } : null,
    }));
  },

  clearActiveScan: () => {
    set({ activeScan: null, scanning: false });
  },

  setPopover: (open, mode) => {
    set({ popoverOpen: open, popoverMode: mode });
  },

  incrementNewDevices: () => {
    set((s) => ({
      activeScan: s.activeScan
        ? { ...s.activeScan, newDevicesCount: s.activeScan.newDevicesCount + 1 }
        : null,
    }));
  },
}));
