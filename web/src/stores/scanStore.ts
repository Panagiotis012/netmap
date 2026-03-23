import { create } from "zustand";
import type { ScanJob } from "../lib/types";
import { api } from "../lib/api";

interface ScanState {
  scans: ScanJob[];
  scanning: boolean;
  fetch: () => Promise<void>;
  triggerScan: (type: "discovery" | "port" | "full", target: string) => Promise<void>;
}

export const useScanStore = create<ScanState>((set) => ({
  scans: [],
  scanning: false,

  fetch: async () => {
    const result = await api.scans.list();
    set({ scans: result.items || [] });
  },

  triggerScan: async (type, target) => {
    set({ scanning: true });
    try {
      await api.scans.trigger(type, target);
    } finally {
      setTimeout(() => set({ scanning: false }), 2000);
    }
  },
}));
