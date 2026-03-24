import { create } from "zustand";

export interface Alert {
  id: string;
  type: "device_discovered" | "device_updated" | "device_offline" | "scan_started" | "scan_completed" | "scan_failed";
  message: string;
  timestamp: string; // ISO string
  device_id?: string;
  scan_id?: string;
  read: boolean;
}

interface AlertsState {
  alerts: Alert[];
  unread: number;
  fetch: () => Promise<void>;
  addAlert: (alert: Omit<Alert, "id" | "read">) => void;
  markRead: () => Promise<void>;
  clear: () => Promise<void>;
}

export const useAlertsStore = create<AlertsState>((set) => ({
  alerts: [],
  unread: 0,

  fetch: async () => {
    try {
      const res = await fetch("/api/v1/alerts", { credentials: "same-origin" });
      if (!res.ok) return;
      const data = await res.json();
      set({ alerts: data.alerts ?? [], unread: data.unread ?? 0 });
    } catch {
      // ignore — backend may not be ready yet
    }
  },

  addAlert: (alert) => {
    const newAlert: Alert = {
      ...alert,
      id: crypto.randomUUID(),
      read: false,
    };
    // Persist to backend
    fetch("/api/v1/alerts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        id: newAlert.id,
        type: newAlert.type,
        message: newAlert.message,
        timestamp: newAlert.timestamp,
        device_id: newAlert.device_id ?? "",
        scan_id: newAlert.scan_id ?? "",
      }),
      credentials: "same-origin",
    }).catch(() => {/* ignore */});

    set((s) => ({
      alerts: [newAlert, ...s.alerts].slice(0, 200),
      unread: s.unread + 1,
    }));
  },

  markRead: async () => {
    try {
      await fetch("/api/v1/alerts/read", { method: "POST", credentials: "same-origin" });
    } catch {/* ignore */}
    set({ unread: 0 });
  },

  clear: async () => {
    try {
      await fetch("/api/v1/alerts", { method: "DELETE", credentials: "same-origin" });
    } catch {/* ignore */}
    set({ alerts: [], unread: 0 });
  },
}));
