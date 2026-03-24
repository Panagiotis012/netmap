import { useEffect } from "react";
import { wsClient } from "../../lib/ws";
import { useScanStore } from "../../stores/scanStore";
import { useDeviceStore } from "../../stores/deviceStore";
import { useAlertsStore } from "../../stores/alertsStore";
import type { Device } from "../../lib/types";

interface ScanProgressPayload {
  scan_id: string;
  hosts_scanned: number;
  hosts_total: number;
  hosts_found: number;
  percent: number;
  eta_seconds: number;
}

interface ScanStartedPayload {
  id: string;
  target: string;
}

export function useScanProgress() {
  // Read actions from deviceStore outside selector to avoid stale-closure lint
  const upsert = useDeviceStore((s) => s.upsert);

  useEffect(() => {
    let dismissTimer: ReturnType<typeof setTimeout> | null = null;

    const unsubProgress = wsClient.on("scan.progress", (e) => {
      const p = e.payload as ScanProgressPayload;
      useScanStore.getState().setActiveScan({
        hostsScanned: p.hosts_scanned,
        hostsTotal: p.hosts_total,
        hostsFound: p.hosts_found,
        percent: p.percent,
        etaSeconds: p.eta_seconds,
      });
    });

    const unsubStarted = wsClient.on("scan.started", (e) => {
      const p = e.payload as ScanStartedPayload;
      useAlertsStore.getState().addAlert({
        type: "scan_started",
        message: `Scan started: ${p.id}`,
        timestamp: new Date().toISOString(),
        scanId: p.id,
      });
    });

    const unsubCompleted = wsClient.on("scan.completed", () => {
      // Cancel any pending dismiss timer from a previous event before creating a new one
      if (dismissTimer !== null) clearTimeout(dismissTimer);
      const store = useScanStore.getState();
      store.setPopover(true, "complete");
      // scanning is done — re-enable the button immediately
      useScanStore.setState({ scanning: false });
      store.fetch();
      // Keep activeScan alive so the complete popover can display newDevicesCount.
      // Clear it after 8s auto-dismiss, but only if still in complete mode.
      dismissTimer = setTimeout(() => {
        if (useScanStore.getState().popoverMode === "complete") {
          useScanStore.getState().clearActiveScan();
          useScanStore.getState().setPopover(false, null);
        }
        dismissTimer = null;
      }, 8000);

      useAlertsStore.getState().addAlert({
        type: "scan_completed",
        message: "Scan complete",
        timestamp: new Date().toISOString(),
      });
    });

    const unsubDiscovered = wsClient.on("device.discovered", (e) => {
      const device = e.payload as Device;
      upsert(device);
      useScanStore.getState().incrementNewDevices();

      const label = device.hostname || device.ip_addresses?.[0] || "Unknown";
      useAlertsStore.getState().addAlert({
        type: "device_discovered",
        message: `New device: ${label}`,
        timestamp: new Date().toISOString(),
        deviceId: device.id,
      });
    });

    return () => {
      unsubProgress();
      unsubStarted();
      unsubCompleted();
      unsubDiscovered();
      if (dismissTimer !== null) clearTimeout(dismissTimer);
    };
  }, [upsert]);
}
