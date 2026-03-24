import { create } from "zustand";

export interface Alert {
  id: string;
  type: "device_discovered" | "device_updated" | "scan_started" | "scan_completed" | "scan_failed";
  message: string;
  timestamp: string; // ISO string
  deviceId?: string;
  scanId?: string;
}

interface AlertsState {
  alerts: Alert[];   // newest first, max 100
  unread: number;    // count since last markRead
  addAlert: (alert: Omit<Alert, "id">) => void;
  markRead: () => void;
  clear: () => void;
}

export const useAlertsStore = create<AlertsState>((set) => ({
  alerts: [],
  unread: 0,

  addAlert: (alert) => {
    const newAlert: Alert = { ...alert, id: crypto.randomUUID() };
    set((s) => ({
      alerts: [newAlert, ...s.alerts].slice(0, 100),
      unread: s.unread + 1,
    }));
  },

  markRead: () => {
    set({ unread: 0 });
  },

  clear: () => {
    set({ alerts: [], unread: 0 });
  },
}));
