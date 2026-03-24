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
  const setActiveScan = useScanStore((s) => s.setActiveScan);
  const clearActiveScan = useScanStore((s) => s.clearActiveScan);
  const setPopover = useScanStore((s) => s.setPopover);
  const incrementNewDevices = useScanStore((s) => s.incrementNewDevices);
  const upsert = useDeviceStore((s) => s.upsert);
  const fetchScans = useScanStore((s) => s.fetch);

  useEffect(() => {
    let dismissTimer: ReturnType<typeof setTimeout> | null = null;

    const unsubProgress = wsClient.on("scan.progress", (e) => {
      const p = e.payload as ScanProgressPayload;
      setActiveScan({
        hostsScanned: p.hosts_scanned,
        hostsTotal: p.hosts_total,
        hostsFound: p.hosts_found,
        percent: p.percent,
        etaSeconds: p.eta_seconds,
      });
    });

    const unsubCompleted = wsClient.on("scan.completed", () => {
      setPopover(true, "complete");
      fetchScans();
      // Auto-dismiss after 8s; clear activeScan with the popover so complete
      // mode can still display the newDevicesCount until then.
      dismissTimer = setTimeout(() => {
        clearActiveScan();
        setPopover(false, null);
      }, 8000);
    });

    const unsubDiscovered = wsClient.on("device.discovered", (e) => {
      upsert(e.payload as Device);
      incrementNewDevices();
    });

    return () => {
      unsubProgress();
      unsubCompleted();
      unsubDiscovered();
      if (dismissTimer !== null) clearTimeout(dismissTimer);
    };
  }, []);
}
