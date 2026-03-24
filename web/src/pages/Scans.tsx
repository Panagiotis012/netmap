import { useEffect, useState } from "react";
import { useScanStore } from "../stores/scanStore";
import { useNetworkStore } from "../stores/networkStore";
import type { ScanJob, ScanType } from "../lib/types";

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    completed: "#2dd4bf", running: "#f59e0b", failed: "#ef4444", cancelled: "#71717a", pending: "#a1a1aa",
  };
  return (
    <span style={{ color: colors[status] ?? "#a1a1aa", fontSize: "12px", fontWeight: 600 }}>
      {status}
    </span>
  );
}

function ScanRow({ scan }: { scan: ScanJob }) {
  const [expanded, setExpanded] = useState(false);
  const activeScan = useScanStore((s) => s.activeScan);
  const isRunning = scan.status === "running" || (activeScan?.id === scan.id);

  return (
    <>
      <tr
        onClick={() => setExpanded(!expanded)}
        style={{ cursor: "pointer", borderBottom: "1px solid #2a2e3a" }}
      >
        <td style={tdStyle}>{new Date(scan.started_at).toLocaleString()}</td>
        <td style={tdStyle}>{scan.target}</td>
        <td style={tdStyle}>{scan.type}</td>
        <td style={tdStyle}><StatusBadge status={scan.status} /></td>
        <td style={tdStyle}>{scan.results?.stats?.hosts_up ?? (isRunning ? activeScan?.hostsFound ?? "—" : "—")}</td>
        <td style={tdStyle}>{scan.results?.stats?.duration_ms ? `${(scan.results.stats.duration_ms / 1000).toFixed(1)}s` : "—"}</td>
      </tr>
      {isRunning && (
        <tr>
          <td colSpan={6} style={{ padding: "4px 12px", background: "#0f1117" }}>
            <div style={{ background: "#2a2e3a", borderRadius: "4px", height: "4px", overflow: "hidden" }}>
              <div style={{ background: "#f59e0b", height: "100%", width: `${activeScan?.percent ?? 0}%`, transition: "width 0.3s" }} />
            </div>
          </td>
        </tr>
      )}
      {expanded && scan.results?.hosts && scan.results.hosts.length > 0 && (
        <tr>
          <td colSpan={6} style={{ padding: "8px 16px", background: "#0f1117" }}>
            <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "12px" }}>
              <thead>
                <tr>
                  {["IP", "MAC", "Hostname", "Ports"].map(h => (
                    <th key={h} style={{ textAlign: "left", color: "#71717a", padding: "4px 8px", fontWeight: 500 }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {scan.results.hosts.map((host) => (
                  <tr key={host.ip} style={{ borderBottom: "1px solid #1e2130" }}>
                    <td style={{ padding: "4px 8px", color: "#e4e4e7" }}>{host.ip}</td>
                    <td style={{ padding: "4px 8px", color: "#71717a" }}>{host.mac || "—"}</td>
                    <td style={{ padding: "4px 8px", color: "#a1a1aa" }}>{host.hostname || "—"}</td>
                    <td style={{ padding: "4px 8px", color: "#71717a" }}>
                      {host.ports?.map(p => p.number).join(", ") || "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </td>
        </tr>
      )}
    </>
  );
}

const tdStyle: React.CSSProperties = { padding: "10px 12px", color: "#e4e4e7", fontSize: "13px" };

export function Scans() {
  const scans = useScanStore((s) => s.scans);
  const scanning = useScanStore((s) => s.scanning);
  const triggerScan = useScanStore((s) => s.triggerScan);
  const fetchScans = useScanStore((s) => s.fetch);
  const networks = useNetworkStore((s) => s.networks);
  const fetchNetworks = useNetworkStore((s) => s.fetch);

  const [scanType, setScanType] = useState<ScanType>("discovery");
  // selectTarget: the value from the network dropdown ("" = custom)
  const [selectTarget, setSelectTarget] = useState("");
  // customSubnet: text typed when "Custom..." is selected
  const [customSubnet, setCustomSubnet] = useState("");
  const [runError, setRunError] = useState<string | null>(null);
  const [page, setPage] = useState(0);
  const PAGE_SIZE = 20;

  useEffect(() => {
    fetchScans();
    fetchNetworks();
  }, [fetchScans, fetchNetworks]);

  const effectiveTarget = selectTarget !== "" ? selectTarget : customSubnet;

  const handleRun = async () => {
    const t = effectiveTarget || networks[0]?.subnet;
    if (!t) return;
    setRunError(null);
    try {
      await triggerScan(scanType, t);
      fetchScans();
    } catch (err) {
      setRunError(err instanceof Error ? err.message : "Scan failed");
    }
  };

  const totalPages = Math.ceil(scans.length / PAGE_SIZE);
  const pageItems = scans.slice(page * PAGE_SIZE, (page + 1) * PAGE_SIZE);

  return (
    <div style={{ padding: "24px", flex: 1, overflow: "auto" }}>
      <h2 style={{ color: "#f4f4f5", margin: "0 0 20px", fontSize: "18px" }}>Scans</h2>

      {/* Manual trigger panel */}
      <div style={{ background: "#1a1d27", borderRadius: "8px", padding: "16px", marginBottom: "24px", display: "flex", gap: "10px", alignItems: "flex-end" }}>
        <div>
          <label style={{ display: "block", fontSize: "11px", color: "#71717a", marginBottom: "4px" }}>Network / Target</label>
          <select
            value={selectTarget}
            onChange={(e) => setSelectTarget(e.target.value)}
            style={{ background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px", color: "#e4e4e7", padding: "6px 10px", fontSize: "13px" }}
          >
            {networks.map((n) => (
              <option key={n.id} value={n.subnet}>{n.name} ({n.subnet})</option>
            ))}
            <option value="">Custom...</option>
          </select>
        </div>
        {selectTarget === "" && (
          <div>
            <label style={{ display: "block", fontSize: "11px", color: "#71717a", marginBottom: "4px" }}>Subnet</label>
            <input
              type="text"
              value={customSubnet}
              placeholder="192.168.1.0/24"
              onChange={(e) => setCustomSubnet(e.target.value)}
              style={{ background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px", color: "#e4e4e7", padding: "6px 10px", fontSize: "13px" }}
            />
          </div>
        )}
        <div>
          <label style={{ display: "block", fontSize: "11px", color: "#71717a", marginBottom: "4px" }}>Type</label>
          <select
            value={scanType}
            onChange={(e) => setScanType(e.target.value as ScanType)}
            style={{ background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px", color: "#e4e4e7", padding: "6px 10px", fontSize: "13px" }}
          >
            <option value="discovery">Discovery</option>
            <option value="port">Port</option>
            <option value="full">Full</option>
          </select>
        </div>
        <div style={{ display: "flex", flexDirection: "column", gap: "4px" }}>
          <button
            onClick={handleRun}
            disabled={scanning}
            style={{ padding: "7px 16px", borderRadius: "6px", border: "none", background: scanning ? "#2a2e3a" : "#2dd4bf", color: scanning ? "#71717a" : "#0f1117", cursor: scanning ? "not-allowed" : "pointer", fontWeight: 600, fontSize: "13px" }}
          >
            {scanning ? "Running..." : "Run Scan"}
          </button>
          {runError && <span style={{ fontSize: "11px", color: "#ef4444" }}>{runError}</span>}
        </div>
      </div>

      {/* History table */}
      <div style={{ background: "#1a1d27", borderRadius: "8px", overflow: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ borderBottom: "1px solid #2a2e3a" }}>
              {["Date", "Target", "Type", "Status", "Devices", "Duration"].map(h => (
                <th key={h} style={{ textAlign: "left", padding: "10px 12px", fontSize: "12px", color: "#71717a", fontWeight: 500 }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {pageItems.length === 0 ? (
              <tr><td colSpan={6} style={{ padding: "32px", textAlign: "center", color: "#71717a", fontSize: "13px" }}>No scans yet</td></tr>
            ) : (
              pageItems.map((scan) => <ScanRow key={scan.id} scan={scan} />)
            )}
          </tbody>
        </table>
      </div>
      {totalPages > 1 && (
        <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end", marginTop: "12px", alignItems: "center" }}>
          <button
            onClick={() => setPage((p) => Math.max(0, p - 1))}
            disabled={page === 0}
            style={{ padding: "4px 10px", borderRadius: "5px", border: "1px solid #2a2e3a", background: "transparent", color: page === 0 ? "#71717a" : "#a1a1aa", cursor: page === 0 ? "not-allowed" : "pointer", fontSize: "12px" }}
          >
            ← Prev
          </button>
          <span style={{ fontSize: "12px", color: "#71717a" }}>{page + 1} / {totalPages}</span>
          <button
            onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
            disabled={page >= totalPages - 1}
            style={{ padding: "4px 10px", borderRadius: "5px", border: "1px solid #2a2e3a", background: "transparent", color: page >= totalPages - 1 ? "#71717a" : "#a1a1aa", cursor: page >= totalPages - 1 ? "not-allowed" : "pointer", fontSize: "12px" }}
          >
            Next →
          </button>
        </div>
      )}
    </div>
  );
}
