import { useNavigate } from "react-router-dom";
import { useScanStore } from "../../stores/scanStore";
import { X } from "lucide-react";

export function ScanPopover() {
  const activeScan = useScanStore((s) => s.activeScan);
  const popoverOpen = useScanStore((s) => s.popoverOpen);
  const popoverMode = useScanStore((s) => s.popoverMode);
  const cancelScan = useScanStore((s) => s.cancelScan);
  const setPopover = useScanStore((s) => s.setPopover);
  const navigate = useNavigate();

  if (!popoverOpen) return null;

  const handleStop = async () => {
    await cancelScan();
  };

  if (popoverMode === "progress" && activeScan) {
    return (
      <div style={{
        position: "absolute", top: "52px", right: "16px", zIndex: 100,
        background: "#1e2130", border: "1px solid #2a2e3a", borderRadius: "8px",
        padding: "14px", width: "260px", boxShadow: "0 8px 24px rgba(0,0,0,0.4)",
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", marginBottom: "8px" }}>
          <span style={{ fontSize: "13px", fontWeight: 600, color: "#f4f4f5" }}>Discovery Scan</span>
          <span style={{ fontSize: "11px", color: "#2dd4bf" }}>● running</span>
        </div>
        <div style={{ fontSize: "11px", color: "#71717a", marginBottom: "8px" }}>
          Target: {activeScan.target}
        </div>
        <div style={{ background: "#0f1117", borderRadius: "4px", height: "6px", marginBottom: "6px" }}>
          <div style={{
            background: "#2dd4bf", borderRadius: "4px", height: "100%",
            width: `${activeScan.percent}%`, transition: "width 0.3s",
          }} />
        </div>
        <div style={{ display: "flex", justifyContent: "space-between", fontSize: "11px", color: "#a1a1aa", marginBottom: "10px" }}>
          <span>{activeScan.hostsScanned}/{activeScan.hostsTotal} hosts · {activeScan.hostsFound} found</span>
          <span>{activeScan.percent}%</span>
        </div>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
          {activeScan.etaSeconds > 0 && (
            <span style={{ fontSize: "11px", color: "#71717a" }}>ETA ~{activeScan.etaSeconds}s</span>
          )}
          <button onClick={handleStop} style={{
            marginLeft: "auto", padding: "4px 10px", borderRadius: "5px",
            background: "transparent", border: "1px solid #ef4444",
            color: "#ef4444", cursor: "pointer", fontSize: "12px",
          }}>
            Stop
          </button>
        </div>
      </div>
    );
  }

  if (popoverMode === "complete") {
    const newCount = activeScan?.newDevicesCount ?? 0;
    return (
      <div style={{
        position: "absolute", top: "52px", right: "16px", zIndex: 100,
        background: "#1e2130", border: "1px solid #2dd4bf40", borderRadius: "8px",
        padding: "14px", width: "260px", boxShadow: "0 8px 24px rgba(0,0,0,0.4)",
      }}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "10px" }}>
          <span style={{ fontSize: "13px", fontWeight: 600, color: "#f4f4f5" }}>✓ Scan complete</span>
          <button onClick={() => setPopover(false, null)} style={{ background: "none", border: "none", cursor: "pointer", color: "#71717a" }}>
            <X size={14} />
          </button>
        </div>
        {newCount > 0 && (
          <div style={{ fontSize: "12px", color: "#a1a1aa", marginBottom: "10px" }}>
            <span style={{ color: "#2dd4bf", fontWeight: 600 }}>{newCount} new</span> device{newCount !== 1 ? "s" : ""} discovered
          </div>
        )}
        <button
          onClick={() => { setPopover(false, null); navigate("/scans"); }}
          style={{ background: "none", border: "none", cursor: "pointer", color: "#2dd4bf", fontSize: "12px", padding: 0 }}
        >
          δες ιστορικό →
        </button>
      </div>
    );
  }

  return null;
}
