import { create } from "zustand";
import type { Device } from "../lib/types";
import { api } from "../lib/api";

interface DeviceState {
  devices: Device[];
  loading: boolean;
  error: string | null;
  fetch: () => Promise<void>;
  upsert: (device: Device) => void;
  remove: (id: string) => void;
  updatePosition: (id: string, x: number, y: number) => void;
}

export const useDeviceStore = create<DeviceState>((set) => ({
  devices: [],
  loading: false,
  error: null,

  fetch: async () => {
    set({ loading: true, error: null });
    try {
      const result = await api.devices.list({ limit: "1000" });
      set({ devices: result.items || [], loading: false });
    } catch (e) {
      set({ error: (e as Error).message, loading: false });
    }
  },

  upsert: (device) => {
    set((state) => {
      const idx = state.devices.findIndex((d) => d.id === device.id);
      if (idx >= 0) {
        const updated = [...state.devices];
        updated[idx] = device;
        return { devices: updated };
      }
      return { devices: [...state.devices, device] };
    });
  },

  remove: (id) => {
    set((state) => ({
      devices: state.devices.filter((d) => d.id !== id),
    }));
  },

  updatePosition: (id, x, y) => {
    set((state) => ({
      devices: state.devices.map((d) =>
        d.id === id ? { ...d, map_x: x, map_y: y } : d
      ),
    }));
    api.devices.update(id, { map_x: x, map_y: y } as Partial<Device>);
  },
}));
