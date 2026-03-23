import { describe, it, expect, vi, beforeEach } from "vitest";
import { useScanStore } from "./scanStore";

// Mock the api module so tests don't make real HTTP calls
vi.mock("../lib/api", () => ({
  api: {
    scans: {
      trigger: vi.fn(),
      cancel: vi.fn(),
    },
  },
}));

import { api } from "../lib/api";

beforeEach(() => {
  useScanStore.setState({
    activeScan: null,
    popoverOpen: false,
    popoverMode: null,
    scanning: false,
    scans: [],
  });
  vi.clearAllMocks();
});

describe("startScan", () => {
  it("sets popoverMode to 'progress' and popoverOpen to true on success", async () => {
    vi.mocked(api.scans.trigger).mockResolvedValue({ id: "scan-123", status: "running" });

    await useScanStore.getState().startScan("192.168.1.0/24");

    const state = useScanStore.getState();
    expect(state.popoverMode).toBe("progress");
    expect(state.popoverOpen).toBe(true);
    expect(state.activeScan?.id).toBe("scan-123");
    expect(state.activeScan?.target).toBe("192.168.1.0/24");
  });

  it("throws on api error", async () => {
    vi.mocked(api.scans.trigger).mockRejectedValue(new Error("scan trigger failed"));

    await expect(useScanStore.getState().startScan("192.168.1.0/24")).rejects.toThrow();
  });
});

describe("cancelScan", () => {
  it("calls cancel api and clears activeScan + popover state", async () => {
    vi.mocked(api.scans.cancel).mockResolvedValue(undefined);

    useScanStore.setState({
      activeScan: {
        id: "scan-abc", target: "192.168.1.0/24",
        hostsScanned: 5, hostsTotal: 254, hostsFound: 2,
        percent: 2, etaSeconds: 30, newDevicesCount: 0,
      },
      scanning: true,
      popoverOpen: true,
      popoverMode: "progress",
    });

    await useScanStore.getState().cancelScan();

    expect(api.scans.cancel).toHaveBeenCalledWith("scan-abc");
    const state = useScanStore.getState();
    expect(state.activeScan).toBeNull();
    expect(state.scanning).toBe(false);
    expect(state.popoverOpen).toBe(false);
    expect(state.popoverMode).toBeNull();
  });

  it("does nothing if no activeScan", async () => {
    await useScanStore.getState().cancelScan();
    expect(api.scans.cancel).not.toHaveBeenCalled();
  });
});
