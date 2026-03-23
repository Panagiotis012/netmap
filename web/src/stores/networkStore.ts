import { create } from "zustand";
import { api } from "../lib/api";
import type { Network } from "../lib/types";

interface NetworkState {
  networks: Network[];
  fetch: () => Promise<void>;
}

export const useNetworkStore = create<NetworkState>((set) => ({
  networks: [],
  fetch: async () => {
    const data = await api.networks.list();
    set({ networks: Array.isArray(data) ? data : [] });
  },
}));
