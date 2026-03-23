import { create } from "zustand";

interface UIState {
  selectedDeviceId: string | null;
  panelOpen: boolean;
  commandPaletteOpen: boolean;
  selectDevice: (id: string | null) => void;
  togglePanel: (open?: boolean) => void;
  toggleCommandPalette: (open?: boolean) => void;
}

export const useUIStore = create<UIState>((set) => ({
  selectedDeviceId: null,
  panelOpen: false,
  commandPaletteOpen: false,

  selectDevice: (id) =>
    set({ selectedDeviceId: id, panelOpen: id !== null }),

  togglePanel: (open) =>
    set((s) => ({ panelOpen: open ?? !s.panelOpen })),

  toggleCommandPalette: (open) =>
    set((s) => ({ commandPaletteOpen: open ?? !s.commandPaletteOpen })),
}));
