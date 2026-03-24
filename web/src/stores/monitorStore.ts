import { create } from "zustand";
import { api } from "../lib/api";
import type { Monitor, MonitorCheck } from "../lib/types";

interface MonitorState {
  monitors: Monitor[];
  checks: Record<string, MonitorCheck[]>;
  loading: boolean;
  fetch: () => Promise<void>;
  fetchChecks: (id: string) => Promise<void>;
  create: (data: Partial<Monitor>) => Promise<Monitor>;
  update: (id: string, data: Partial<Monitor>) => Promise<Monitor>;
  remove: (id: string) => Promise<void>;
}

export const useMonitorStore = create<MonitorState>((set) => ({
  monitors: [],
  checks: {},
  loading: false,

  fetch: async () => {
    set({ loading: true });
    try {
      const monitors = await api.monitors.list();
      set({ monitors });
    } finally {
      set({ loading: false });
    }
  },

  fetchChecks: async (id: string) => {
    const checks = await api.monitors.checks(id);
    set((s) => ({ checks: { ...s.checks, [id]: checks } }));
  },

  create: async (data) => {
    const monitor = await api.monitors.create(data);
    set((s) => ({ monitors: [...s.monitors, monitor] }));
    return monitor;
  },

  update: async (id, data) => {
    const monitor = await api.monitors.update(id, data);
    set((s) => ({ monitors: s.monitors.map((m) => (m.id === id ? monitor : m)) }));
    return monitor;
  },

  remove: async (id) => {
    await api.monitors.delete(id);
    set((s) => ({ monitors: s.monitors.filter((m) => m.id !== id) }));
  },
}));
