import { describe, it, expect, vi, beforeEach } from "vitest";
import { useScanStore } from "./scanStore";

beforeEach(() => {
  useScanStore.setState({
    activeScan: null,
    popoverOpen: false,
    popoverMode: null,
    scanning: false,
    scans: [],
  });
});

describe("startScan", () => {
  it("sets popoverMode to 'progress' and popoverOpen to true on success", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ id: "scan-123", status: "running" }),
    }) as any;

    await useScanStore.getState().startScan("192.168.1.0/24");

    const state = useScanStore.getState();
    expect(state.popoverMode).toBe("progress");
    expect(state.popoverOpen).toBe(true);
    expect(state.activeScan?.id).toBe("scan-123");
    expect(state.activeScan?.target).toBe("192.168.1.0/24");
  });

  it("throws on non-ok response", async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      json: async () => ({}),
    }) as any;

    await expect(useScanStore.getState().startScan("192.168.1.0/24")).rejects.toThrow();
  });
});

describe("cancelScan", () => {
  it("calls DELETE /api/v1/scans/{id} and clears activeScan", async () => {
    useScanStore.setState({
      activeScan: {
        id: "scan-abc", target: "192.168.1.0/24",
        hostsScanned: 5, hostsTotal: 254, hostsFound: 2,
        percent: 2, etaSeconds: 30, newDevicesCount: 0,
      },
      scanning: true,
    });

    const deleteMock = vi.fn().mockResolvedValue({ ok: true });
    global.fetch = deleteMock as any;

    await useScanStore.getState().cancelScan();

    expect(deleteMock).toHaveBeenCalledWith(
      "/api/v1/scans/scan-abc",
      { method: "DELETE" }
    );
    expect(useScanStore.getState().activeScan).toBeNull();
    expect(useScanStore.getState().scanning).toBe(false);
  });

  it("does nothing if no activeScan", async () => {
    const fetchMock = vi.fn();
    global.fetch = fetchMock as any;
    await useScanStore.getState().cancelScan();
    expect(fetchMock).not.toHaveBeenCalled();
  });
});
