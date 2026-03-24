import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDeviceStore } from "../stores/deviceStore";
import { useUIStore } from "../stores/uiStore";

export function DeviceList() {
  const devices = useDeviceStore((s) => s.devices);
  const loading = useDeviceStore((s) => s.loading);
  const selectDevice = useUIStore((s) => s.selectDevice);
  const navigate = useNavigate();
  const [search, setSearch] = useState("");

  const q = search.toLowerCase();
  const filtered = devices
    .filter((d) =>
      !q ||
      d.hostname?.toLowerCase().includes(q) ||
      d.ip_addresses?.some((ip) => ip.includes(q)) ||
      d.mac_addresses?.some((m) => m.toLowerCase().includes(q)) ||
      d.os?.toLowerCase().includes(q)
    )
    .sort((a, b) => {
      if (a.status === b.status) return 0;
      return a.status === "online" ? -1 : 1;
    });

  const statusColor = (s: string) =>
    s === "online" ? "#2dd4bf" : s === "offline" ? "#ef4444" : "#71717a";

  const handleRowClick = (id: string) => {
    selectDevice(id);
    navigate("/");
  };

  if (loading) {
    return (
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        Loading…
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflow: "auto", padding: "24px" }}>
      <div style={{ display: "flex", alignItems: "center", gap: "16px", marginBottom: "20px" }}>
        <h2 style={{ color: "#f4f4f5", margin: 0, fontSize: "16px" }}>
          Devices <span style={{ color: "#71717a", fontWeight: 400 }}>({devices.length})</span>
        </h2>
        <input
          type="text"
          placeholder="Search hostname, IP, MAC, OS…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          style={{
            background: "#1e2130", border: "1px solid #2a2e3a", borderRadius: "6px",
            color: "#e4e4e7", padding: "5px 10px", fontSize: "13px", width: "260px",
            outline: "none",
          }}
        />
      </div>

      {devices.length === 0 ? (
        <div style={{ color: "#71717a", padding: "48px 0", textAlign: "center" }}>
          No devices yet — run a scan to discover your network.
        </div>
      ) : filtered.length === 0 ? (
        <div style={{ color: "#71717a", padding: "48px 0", textAlign: "center" }}>
          No devices match "{search}"
        </div>
      ) : (
        <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "13px" }}>
          <thead>
            <tr style={{ borderBottom: "1px solid #2a2e3a" }}>
              {["Status", "Hostname", "IP", "MAC", "OS", "Last Seen"].map((h) => (
                <th key={h} style={{ textAlign: "left", padding: "6px 12px", color: "#71717a", fontWeight: 500 }}>{h}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {filtered.map((d) => (
              <tr
                key={d.id}
                onClick={() => handleRowClick(d.id)}
                style={{ borderBottom: "1px solid #1e2130", cursor: "pointer" }}
                onMouseEnter={(e) => (e.currentTarget.style.background = "#1e2130")}
                onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
              >
                <td style={{ padding: "8px 12px" }}>
                  <span style={{ color: statusColor(d.status), fontSize: "11px", fontWeight: 600 }}>
                    ● {d.status}
                  </span>
                </td>
                <td style={{ padding: "8px 12px", color: "#f4f4f5" }}>
                  {d.hostname || <span style={{ color: "#52525b" }}>—</span>}
                </td>
                <td style={{ padding: "8px 12px", color: "#a1a1aa", fontFamily: "monospace" }}>
                  {d.ip_addresses?.[0] ?? "—"}
                </td>
                <td style={{ padding: "8px 12px", color: "#52525b", fontFamily: "monospace", fontSize: "12px" }}>
                  {d.mac_addresses?.[0] ?? "—"}
                </td>
                <td style={{ padding: "8px 12px", color: "#71717a" }}>
                  {d.os || <span style={{ color: "#3f3f46" }}>—</span>}
                </td>
                <td style={{ padding: "8px 12px", color: "#52525b", fontSize: "12px" }}>
                  {d.last_seen_at ? new Date(d.last_seen_at).toLocaleString() : "—"}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
