import { create } from "zustand";

interface AuthState {
  setup: boolean;       // password has been configured
  authenticated: boolean;
  loading: boolean;
  fetchStatus: () => Promise<void>;
  login: (password: string) => Promise<void>;
  setup_password: (password: string) => Promise<void>;
  logout: () => Promise<void>;
}

async function authRequest(path: string, body?: object) {
  const res = await fetch(`/api/v1/auth${path}`, {
    method: body ? "POST" : "GET",
    headers: body ? { "Content-Type": "application/json" } : undefined,
    body: body ? JSON.stringify(body) : undefined,
    credentials: "same-origin",
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error((err as { error: string }).error || res.statusText);
  }
  if (res.status === 204) return {};
  return res.json();
}

export const useAuthStore = create<AuthState>((set) => ({
  setup: false,
  authenticated: false,
  loading: true,

  fetchStatus: async () => {
    try {
      const data = await authRequest("/status");
      set({ setup: data.setup, authenticated: data.authenticated, loading: false });
    } catch {
      set({ loading: false });
    }
  },

  login: async (password: string) => {
    const data = await authRequest("/login", { password });
    set({ authenticated: data.authenticated });
  },

  setup_password: async (password: string) => {
    const data = await authRequest("/setup", { password });
    set({ setup: true, authenticated: data.authenticated });
  },

  logout: async () => {
    await authRequest("/logout", {});
    set({ authenticated: false });
  },
}));
