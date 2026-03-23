import { describe, it, expect, vi, beforeEach } from "vitest";
import { renderHook } from "@testing-library/react";
import { useScanStore } from "../../stores/scanStore";

// Capture wsClient callbacks without referencing the imported symbol
// inside the vi.mock factory (which runs before imports are resolved).
const mockHandlers: Record<string, ((e: any) => void)> = {};

vi.mock("../../lib/ws", () => ({
  wsClient: {
    on: vi.fn((event: string, cb: (e: any) => void) => {
      mockHandlers[event] = cb;
      return vi.fn(); // unsubscribe fn
    }),
  },
}));

vi.mock("../../stores/deviceStore", () => ({
  useDeviceStore: vi.fn((selector: any) => selector({ upsert: vi.fn() })),
}));

beforeEach(() => {
  // Reset handlers and store state
  Object.keys(mockHandlers).forEach((k) => delete mockHandlers[k]);
  useScanStore.setState({
    activeScan: {
      id: "scan-1", target: "10.0.0.0/24",
      hostsScanned: 0, hostsTotal: 0, hostsFound: 0,
      percent: 0, etaSeconds: 0, newDevicesCount: 0,
    },
    popoverOpen: true,
    popoverMode: "progress",
    scanning: true,
  });
});

describe("useScanProgress", () => {
  it("updates activeScan when scan.progress event fires", async () => {
    const { useScanProgress } = await import("./useScanProgress");
    renderHook(() => useScanProgress());

    mockHandlers["scan.progress"]?.({
      payload: {
        scan_id: "scan-1",
        hosts_scanned: 50, hosts_total: 254,
        hosts_found: 5, percent: 20, eta_seconds: 15,
      },
    });

    const state = useScanStore.getState();
    expect(state.activeScan?.hostsScanned).toBe(50);
    expect(state.activeScan?.percent).toBe(20);
    expect(state.activeScan?.etaSeconds).toBe(15);
  });

  it("clears activeScan and transitions to complete mode on scan.completed event", async () => {
    const { useScanProgress } = await import("./useScanProgress");
    renderHook(() => useScanProgress());

    mockHandlers["scan.completed"]?.({ payload: {} });

    const state = useScanStore.getState();
    expect(state.activeScan).toBeNull();
    expect(state.popoverMode).toBe("complete");
  });
});
