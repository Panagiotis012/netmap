import { useEffect, useState } from "react";
import { api } from "../../lib/api";
import type { SystemStatus } from "../../lib/types";

export function General() {
  const [status, setStatus] = useState<SystemStatus | null>(null);

  useEffect(() => {
    api.system.status().then(setStatus).catch(() => {/* status is optional info */});
  }, []);

  const row = (label: string, value: string) => (
    <div style={{ display: "flex", alignItems: "center", padding: "10px 0", borderBottom: "1px solid #2a2e3a" }}>
      <span style={{ width: "140px", fontSize: "12px", color: "#71717a" }}>{label}</span>
      <span style={{ fontSize: "13px", color: "#e4e4e7", fontFamily: "monospace" }}>{value}</span>
    </div>
  );

  return (
    <div style={{ padding: "24px", maxWidth: "480px" }}>
      <h2 style={{ color: "#f4f4f5", margin: "0 0 20px", fontSize: "16px" }}>General</h2>
      <div style={{ background: "#1a1d27", borderRadius: "8px", padding: "0 16px" }}>
        {row("Version", status?.version ?? "—")}
        {status?.started_at && row("Started at", new Date(status.started_at).toLocaleString())}
        {status?.db_path && row("DB Path", status.db_path)}
        {row("Devices online", String(status?.devices_online ?? "—"))}
        {row("Devices total", String(status?.devices_total ?? "—"))}
      </div>
      <p style={{ fontSize: "11px", color: "#71717a", marginTop: "16px" }}>
        Auth and themes are planned for Phase 2b.
      </p>
    </div>
  );
}
