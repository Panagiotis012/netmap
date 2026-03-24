import { useState, useEffect } from "react";
import type { Monitor } from "../lib/types";

function statusColor(s: string) {
  if (s === "up") return "#2dd4bf";
  if (s === "down") return "#ef4444";
  return "#71717a";
}

function uptimeBar(pct: number) {
  const color = pct >= 99 ? "#2dd4bf" : pct >= 90 ? "#f59e0b" : "#ef4444";
  return (
    <div style={{ display: "flex", alignItems: "center", gap: "8px" }}>
      <div style={{ flex: 1, height: "6px", backgroundColor: "#2a2e3a", borderRadius: "3px", overflow: "hidden" }}>
        <div style={{ width: `${pct}%`, height: "100%", backgroundColor: color, borderRadius: "3px", transition: "width 0.5s" }} />
      </div>
      <span style={{ fontSize: "12px", color, fontWeight: 600, minWidth: "44px", textAlign: "right" }}>{pct.toFixed(1)}%</span>
    </div>
  );
}

export function StatusPage() {
  const [monitors, setMonitors] = useState<Monitor[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetch("/api/v1/status-page")
      .then((r) => r.json())
      .then((data: Monitor[]) => setMonitors(data ?? []))
      .catch(() => setError("Failed to load status data"))
      .finally(() => setLoading(false));
  }, []);

  const allUp = monitors.length > 0 && monitors.every((m) => m.status === "up");
  const anyDown = monitors.some((m) => m.status === "down");

  return (
    <div style={{ minHeight: "100vh", backgroundColor: "#0f1117", display: "flex", flexDirection: "column", alignItems: "center", padding: "48px 16px", fontFamily: "system-ui, sans-serif" }}>
      <div style={{ width: "100%", maxWidth: "640px" }}>
        {/* Brand */}
        <div style={{ textAlign: "center", marginBottom: "40px" }}>
          <h1 style={{ margin: "0 0 8px", fontSize: "28px", fontWeight: 700, color: "#f4f4f5", letterSpacing: "-0.02em" }}>
            Status
          </h1>
          <p style={{ margin: 0, fontSize: "14px", color: "#71717a" }}>
            {new Date().toLocaleDateString(undefined, { dateStyle: "long" })}
          </p>
        </div>

        {/* Overall status banner */}
        {!loading && !error && (
          <div style={{
            borderRadius: "12px",
            border: `1px solid ${anyDown ? "#ef4444" : allUp ? "#2dd4bf" : "#2a2e3a"}`,
            backgroundColor: anyDown ? "rgba(239,68,68,0.08)" : allUp ? "rgba(45,212,191,0.08)" : "#1e2130",
            padding: "16px 20px",
            marginBottom: "24px",
            display: "flex",
            alignItems: "center",
            gap: "12px",
          }}>
            <div style={{ width: 10, height: 10, borderRadius: "50%", backgroundColor: anyDown ? "#ef4444" : allUp ? "#2dd4bf" : "#71717a", flexShrink: 0 }} />
            <span style={{ fontSize: "15px", fontWeight: 600, color: anyDown ? "#ef4444" : allUp ? "#2dd4bf" : "#a1a1aa" }}>
              {monitors.length === 0 ? "No monitors configured" : anyDown ? "Some services are down" : allUp ? "All systems operational" : "Checking…"}
            </span>
          </div>
        )}

        {loading && (
          <div style={{ color: "#71717a", textAlign: "center", padding: "48px 0" }}>Loading…</div>
        )}
        {error && (
          <div style={{ color: "#ef4444", textAlign: "center", padding: "48px 0" }}>{error}</div>
        )}

        {/* Monitor list */}
        {monitors.length > 0 && (
          <div style={{ backgroundColor: "#1e2130", borderRadius: "12px", border: "1px solid #2a2e3a", overflow: "hidden" }}>
            {monitors.map((m, i) => (
              <div
                key={m.id}
                style={{ padding: "16px 20px", borderBottom: i < monitors.length - 1 ? "1px solid #2a2e3a" : "none" }}
              >
                <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "10px" }}>
                  <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                    <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: statusColor(m.status) }} />
                    <span style={{ fontSize: "14px", fontWeight: 500, color: "#f4f4f5" }}>{m.name}</span>
                    <span style={{ fontSize: "11px", color: "#52525b", backgroundColor: "#0f1117", padding: "1px 6px", borderRadius: "4px" }}>
                      {m.type.toUpperCase()}
                    </span>
                  </div>
                  <span style={{ fontSize: "12px", fontWeight: 700, color: statusColor(m.status) }}>
                    {m.status.toUpperCase()}
                  </span>
                </div>
                <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px" }}>
                  <div>
                    <div style={{ fontSize: "11px", color: "#52525b", marginBottom: "4px" }}>Last 24 hours</div>
                    {uptimeBar(m.uptime_day)}
                  </div>
                  <div>
                    <div style={{ fontSize: "11px", color: "#52525b", marginBottom: "4px" }}>Last 7 days</div>
                    {uptimeBar(m.uptime_week)}
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}

        <p style={{ marginTop: "32px", textAlign: "center", fontSize: "12px", color: "#3f3f46" }}>
          Powered by NetMap
        </p>
      </div>
    </div>
  );
}
