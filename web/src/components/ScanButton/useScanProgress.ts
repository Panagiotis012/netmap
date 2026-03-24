import { useEffect } from "react";
import { wsClient } from "../../lib/ws";
import { useScanStore } from "../../stores/scanStore";
import { useDeviceStore } from "../../stores/deviceStore";
import type { Device } from "../../lib/types";

interface ScanProgressPayload {
  scan_id: string;
  hosts_scanned: number;
  hosts_total: number;
  hosts_found: number;
  percent: number;
  eta_seconds: number;
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
    });

    const unsubDiscovered = wsClient.on("device.discovered", (e) => {
      upsert(e.payload as Device);
      useScanStore.getState().incrementNewDevices();
    });

    return () => {
      unsubProgress();
      unsubCompleted();
      unsubDiscovered();
      if (dismissTimer !== null) clearTimeout(dismissTimer);
    };
  }, [upsert]);
}
