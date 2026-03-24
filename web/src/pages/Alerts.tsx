import { useEffect } from "react";
import { Radar, Monitor, AlertCircle, CheckCircle, RotateCcw, WifiOff } from "lucide-react";
import { useAlertsStore } from "../stores/alertsStore";
import type { Alert } from "../stores/alertsStore";

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const m = Math.floor(diff / 60000);
  if (m < 1) return "just now";
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  return new Date(iso).toLocaleDateString();
}

function AlertIcon({ type }: { type: Alert["type"] }) {
  const base: React.CSSProperties = { flexShrink: 0, width: 32, height: 32, borderRadius: "8px", display: "flex", alignItems: "center", justifyContent: "center" };

  switch (type) {
    case "device_discovered":
      return (
        <div style={{ ...base, backgroundColor: "rgba(45,212,191,0.12)" }}>
          <Monitor size={16} strokeWidth={1.5} color="#2dd4bf" />
        </div>
      );
    case "device_updated":
      return (
        <div style={{ ...base, backgroundColor: "rgba(45,212,191,0.08)" }}>
          <RotateCcw size={16} strokeWidth={1.5} color="#2dd4bf" />
        </div>
      );
    case "device_offline":
      return (
        <div style={{ ...base, backgroundColor: "rgba(239,68,68,0.12)" }}>
          <WifiOff size={16} strokeWidth={1.5} color="#ef4444" />
        </div>
      );
    case "scan_started":
      return (
        <div style={{ ...base, backgroundColor: "rgba(99,102,241,0.12)" }}>
          <Radar size={16} strokeWidth={1.5} color="#818cf8" />
        </div>
      );
    case "scan_completed":
      return (
        <div style={{ ...base, backgroundColor: "rgba(45,212,191,0.12)" }}>
          <CheckCircle size={16} strokeWidth={1.5} color="#2dd4bf" />
        </div>
      );
    case "scan_failed":
      return (
        <div style={{ ...base, backgroundColor: "rgba(239,68,68,0.12)" }}>
          <AlertCircle size={16} strokeWidth={1.5} color="#ef4444" />
        </div>
      );
    default:
      return (
        <div style={{ ...base, backgroundColor: "rgba(113,113,122,0.12)" }}>
          <AlertCircle size={16} strokeWidth={1.5} color="#71717a" />
        </div>
      );
  }
}

export function Alerts() {
  const alerts = useAlertsStore((s) => s.alerts);
  const fetch = useAlertsStore((s) => s.fetch);
  const markRead = useAlertsStore((s) => s.markRead);
  const clear = useAlertsStore((s) => s.clear);

  useEffect(() => {
    fetch();
    markRead();
  }, []);

  return (
    <div style={{ flex: 1, overflowY: "auto", backgroundColor: "#0f1117", padding: "24px" }}>
      {/* Header */}
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "20px" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <h1 style={{ margin: 0, fontSize: "18px", fontWeight: 600, color: "#f4f4f5" }}>Alerts</h1>
          {alerts.length > 0 && (
            <span style={{
              fontSize: "12px",
              fontWeight: 500,
              color: "#71717a",
              backgroundColor: "#1e2130",
              padding: "2px 8px",
              borderRadius: "12px",
              border: "1px solid #2a2e3a",
            }}>
              {alerts.length}
            </span>
          )}
        </div>
        {alerts.length > 0 && (
          <button
            onClick={clear}
            style={{
              background: "transparent",
              border: "1px solid #2a2e3a",
              borderRadius: "6px",
              color: "#71717a",
              cursor: "pointer",
              fontSize: "13px",
              padding: "5px 12px",
              transition: "all 0.15s",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLButtonElement).style.borderColor = "#ef4444";
              (e.currentTarget as HTMLButtonElement).style.color = "#ef4444";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLButtonElement).style.borderColor = "#2a2e3a";
              (e.currentTarget as HTMLButtonElement).style.color = "#71717a";
            }}
          >
            Clear all
          </button>
        )}
      </div>

      {/* Empty state */}
      {alerts.length === 0 && (
        <div style={{
          display: "flex",
          flexDirection: "column",
          alignItems: "center",
          justifyContent: "center",
          gap: "12px",
          padding: "80px 32px",
          color: "#71717a",
          textAlign: "center",
        }}>
          <AlertCircle size={36} strokeWidth={1} color="#2a2e3a" />
          <p style={{ margin: 0, fontSize: "14px", lineHeight: "1.6", maxWidth: "320px" }}>
            No alerts yet — events from scans and device discovery will appear here.
          </p>
        </div>
      )}

      {/* Alert list */}
      {alerts.length > 0 && (
        <div style={{
          backgroundColor: "#1e2130",
          borderRadius: "10px",
          border: "1px solid #2a2e3a",
          overflow: "hidden",
        }}>
          {alerts.map((alert, index) => (
            <div
              key={alert.id}
              style={{
                display: "flex",
                alignItems: "center",
                gap: "12px",
                padding: "12px 16px",
                borderBottom: index < alerts.length - 1 ? "1px solid #2a2e3a" : "none",
              }}
            >
              <AlertIcon type={alert.type} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <p style={{
                  margin: 0,
                  fontSize: "14px",
                  color: "#f4f4f5",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}>
                  {alert.message}
                </p>
              </div>
              <span style={{
                fontSize: "12px",
                color: "#71717a",
                flexShrink: 0,
                whiteSpace: "nowrap",
              }}>
                {relativeTime(alert.timestamp)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
