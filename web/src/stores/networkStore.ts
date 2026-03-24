import { create } from "zustand";
import { api } from "../lib/api";
import type { Network } from "../lib/types";

interface NetworkState {
  networks: Network[];
  error: string | null;
  fetch: () => Promise<void>;
}

export const useNetworkStore = create<NetworkState>((set) => ({
  networks: [],
  error: null,
  fetch: async () => {
    try {
      const data = await api.networks.list();
      set({ networks: Array.isArray(data) ? data : [], error: null });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : "Failed to load networks" });
    }
  },
}));
