import { useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, CheckCircle, XCircle, Clock } from "lucide-react";
import {
  ResponsiveContainer,
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
} from "recharts";
import { useMonitorStore } from "../stores/monitorStore";

function statusColor(s: string) {
  if (s === "up") return "#2dd4bf";
  if (s === "down") return "#ef4444";
  return "#71717a";
}

function fmt(iso: string) {
  const d = new Date(iso);
  return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

export function MonitorDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { monitors, checks, fetch, fetchChecks } = useMonitorStore();

  useEffect(() => {
    fetch();
    if (id) fetchChecks(id);
  }, [id]);

  const monitor = monitors.find((m) => m.id === id);
  const history = id ? (checks[id] ?? []) : [];

  if (!monitor) {
    return (
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        Loading…
      </div>
    );
  }

  const chartData = [...history].reverse().map((c) => ({
    time: fmt(c.checked_at),
    ms: c.response_time_ms,
    status: c.status,
  }));

  const last = history[0];
  const avgMs = history.length ? Math.round(history.reduce((s, c) => s + c.response_time_ms, 0) / history.length) : 0;
  const downCount = history.filter((c) => c.status === "down").length;

  return (
    <div style={{ flex: 1, overflowY: "auto", padding: "24px", backgroundColor: "#0f1117" }}>
      {/* Back */}
      <button
        onClick={() => navigate("/monitors")}
        style={{ display: "flex", alignItems: "center", gap: "6px", background: "transparent", border: "none", color: "#71717a", cursor: "pointer", fontSize: "13px", marginBottom: "20px", padding: 0 }}
      >
        <ArrowLeft size={14} /> Back to Monitors
      </button>

      {/* Header */}
      <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "24px" }}>
        <div style={{ width: 10, height: 10, borderRadius: "50%", backgroundColor: statusColor(monitor.status), flexShrink: 0 }} />
        <h1 style={{ margin: 0, fontSize: "20px", fontWeight: 600, color: "#f4f4f5" }}>{monitor.name}</h1>
        <span style={{ fontSize: "12px", color: "#52525b", backgroundColor: "#1e2130", padding: "2px 8px", borderRadius: "6px", border: "1px solid #2a2e3a" }}>
          {monitor.type.toUpperCase()}
        </span>
        <span style={{ fontSize: "12px", fontWeight: 600, color: statusColor(monitor.status) }}>
          {monitor.status.toUpperCase()}
        </span>
      </div>

      {/* Stat cards */}
      <div style={{ display: "grid", gridTemplateColumns: "repeat(4, 1fr)", gap: "12px", marginBottom: "24px" }}>
        {[
          { label: "24h Uptime", value: `${monitor.uptime_day.toFixed(1)}%`, color: monitor.uptime_day >= 99 ? "#2dd4bf" : monitor.uptime_day >= 90 ? "#f59e0b" : "#ef4444" },
          { label: "7d Uptime", value: `${monitor.uptime_week.toFixed(1)}%`, color: monitor.uptime_week >= 99 ? "#2dd4bf" : monitor.uptime_week >= 90 ? "#f59e0b" : "#ef4444" },
          { label: "Avg Response", value: avgMs ? `${avgMs}ms` : "—", color: "#a1a1aa" },
          { label: "Incidents (last 100)", value: String(downCount), color: downCount > 0 ? "#ef4444" : "#2dd4bf" },
        ].map((card) => (
          <div key={card.label} style={{ backgroundColor: "#1e2130", borderRadius: "10px", border: "1px solid #2a2e3a", padding: "16px" }}>
            <div style={{ fontSize: "12px", color: "#52525b", marginBottom: "6px" }}>{card.label}</div>
            <div style={{ fontSize: "22px", fontWeight: 700, color: card.color }}>{card.value}</div>
          </div>
        ))}
      </div>

      {/* Response time chart */}
      {chartData.length > 0 && (
        <div style={{ backgroundColor: "#1e2130", borderRadius: "10px", border: "1px solid #2a2e3a", padding: "20px", marginBottom: "24px" }}>
          <h2 style={{ margin: "0 0 16px", fontSize: "14px", color: "#a1a1aa", fontWeight: 500 }}>Response Time (ms)</h2>
          <ResponsiveContainer width="100%" height={180}>
            <AreaChart data={chartData} margin={{ top: 4, right: 4, bottom: 0, left: 0 }}>
              <defs>
                <linearGradient id="rtGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#2dd4bf" stopOpacity={0.25} />
                  <stop offset="95%" stopColor="#2dd4bf" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#2a2e3a" />
              <XAxis dataKey="time" tick={{ fill: "#52525b", fontSize: 10 }} tickLine={false} axisLine={false} interval="preserveStartEnd" />
              <YAxis tick={{ fill: "#52525b", fontSize: 10 }} tickLine={false} axisLine={false} width={36} />
              <Tooltip
                contentStyle={{ background: "#1a1d27", border: "1px solid #2a2e3a", borderRadius: "6px", fontSize: "12px" }}
                labelStyle={{ color: "#a1a1aa" }}
                itemStyle={{ color: "#2dd4bf" }}
              />
              <Area type="monotone" dataKey="ms" stroke="#2dd4bf" strokeWidth={1.5} fill="url(#rtGrad)" dot={false} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      )}

      {/* Recent checks */}
      <div style={{ backgroundColor: "#1e2130", borderRadius: "10px", border: "1px solid #2a2e3a", overflow: "hidden" }}>
        <div style={{ padding: "14px 16px", borderBottom: "1px solid #2a2e3a" }}>
          <h2 style={{ margin: 0, fontSize: "14px", color: "#a1a1aa", fontWeight: 500 }}>Recent Checks</h2>
        </div>
        {history.length === 0 ? (
          <div style={{ padding: "32px", color: "#52525b", fontSize: "13px", textAlign: "center" }}>No checks yet</div>
        ) : (
          history.slice(0, 50).map((c, i) => (
            <div
              key={c.id}
              style={{ display: "flex", alignItems: "center", gap: "12px", padding: "10px 16px", borderBottom: i < Math.min(history.length, 50) - 1 ? "1px solid #2a2e3a" : "none", fontSize: "13px" }}
            >
              {c.status === "up"
                ? <CheckCircle size={14} color="#2dd4bf" />
                : <XCircle size={14} color="#ef4444" />}
              <span style={{ color: c.status === "up" ? "#2dd4bf" : "#ef4444", width: "36px", fontWeight: 600, fontSize: "11px" }}>
                {c.status.toUpperCase()}
              </span>
              <span style={{ color: "#a1a1aa", minWidth: "60px", display: "flex", alignItems: "center", gap: "4px" }}>
                <Clock size={11} />{c.response_time_ms}ms
              </span>
              {c.status_code ? <span style={{ color: "#52525b" }}>HTTP {c.status_code}</span> : null}
              {c.error ? <span style={{ color: "#ef4444", flex: 1, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>{c.error}</span> : null}
              <span style={{ color: "#3f3f46", marginLeft: "auto", flexShrink: 0 }}>
                {new Date(c.checked_at).toLocaleString()}
              </span>
            </div>
          ))
        )}
      </div>

      {/* Monitor config */}
      <div style={{ marginTop: "24px", backgroundColor: "#1e2130", borderRadius: "10px", border: "1px solid #2a2e3a", padding: "16px" }}>
        <h2 style={{ margin: "0 0 12px", fontSize: "14px", color: "#a1a1aa", fontWeight: 500 }}>Configuration</h2>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "8px 24px", fontSize: "13px" }}>
          {[
            ["Target", monitor.url || monitor.host || "—"],
            ["Interval", `${monitor.interval}s`],
            ["Timeout", `${monitor.timeout}s`],
            ["Method", monitor.method || "—"],
            ["Expected Status", monitor.expected_status || "—"],
            ["Keyword", monitor.keyword || "—"],
            ["Webhook", monitor.notify_webhook ? "Configured" : "—"],
            ["Active", monitor.active ? "Yes" : "No"],
            ["Last Checked", last ? new Date(last.checked_at).toLocaleString() : "Never"],
          ].map(([k, v]) => (
            <div key={k} style={{ display: "flex", gap: "8px" }}>
              <span style={{ color: "#52525b", minWidth: "120px" }}>{k}</span>
              <span style={{ color: "#a1a1aa" }}>{v}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
