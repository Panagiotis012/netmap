import { useDeviceStore } from "../stores/deviceStore";

export function DeviceList() {
  const devices = useDeviceStore((s) => s.devices);
  const loading = useDeviceStore((s) => s.loading);

  const online = devices.filter((d) => d.status === "online");
  const offline = devices.filter((d) => d.status === "offline");
  const sorted = [...online, ...offline];

  const statusColor = (s: string) =>
    s === "online" ? "#2dd4bf" : s === "offline" ? "#ef4444" : "#71717a";

  if (loading) {
    return (
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        Loading…
      </div>
    );
  }

  if (devices.length === 0) {
    return (
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        No devices yet — run a scan to discover your network.
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflow: "auto", padding: "24px" }}>
      <h2 style={{ color: "#f4f4f5", margin: "0 0 20px", fontSize: "16px" }}>
        Devices <span style={{ color: "#71717a", fontWeight: 400 }}>({devices.length})</span>
      </h2>
      <table style={{ width: "100%", borderCollapse: "collapse", fontSize: "13px" }}>
        <thead>
          <tr style={{ borderBottom: "1px solid #2a2e3a" }}>
            {["Status", "Hostname", "IP", "MAC", "OS", "Last Seen"].map((h) => (
              <th key={h} style={{ textAlign: "left", padding: "6px 12px", color: "#71717a", fontWeight: 500 }}>{h}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {sorted.map((d) => (
            <tr key={d.id} style={{ borderBottom: "1px solid #1e2130" }}>
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
    </div>
  );
}
