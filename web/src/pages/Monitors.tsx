import { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Plus, Trash2, Activity, Edit2 } from "lucide-react";
import { useMonitorStore } from "../stores/monitorStore";
import type { Monitor, MonitorType } from "../lib/types";

function statusColor(s: string) {
  if (s === "up") return "#2dd4bf";
  if (s === "down") return "#ef4444";
  return "#71717a";
}

function uptimeBadge(pct: number) {
  const color = pct >= 99 ? "#2dd4bf" : pct >= 90 ? "#f59e0b" : "#ef4444";
  return <span style={{ color, fontWeight: 600, fontSize: "12px" }}>{pct.toFixed(1)}%</span>;
}

const defaultForm = (): Partial<Monitor> => ({
  name: "",
  type: "http",
  url: "",
  host: "",
  port: 80,
  interval: 60,
  timeout: 10,
  method: "GET",
  expected_status: 200,
  keyword: "",
  active: true,
  notify_webhook: "",
});

function MonitorModal({ monitor, onClose, onSave }: {
  monitor: Partial<Monitor> | null;
  onClose: () => void;
  onSave: (data: Partial<Monitor>) => Promise<void>;
}) {
  const [form, setForm] = useState<Partial<Monitor>>(monitor ?? defaultForm());
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState("");

  const set = (k: keyof Monitor, v: unknown) => setForm((f) => ({ ...f, [k]: v }));

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    setError("");
    try {
      await onSave(form);
      onClose();
    } catch (err) {
      setError((err as Error).message);
    } finally {
      setSaving(false);
    }
  };

  const overlay: React.CSSProperties = {
    position: "fixed", inset: 0, backgroundColor: "rgba(0,0,0,0.6)",
    display: "flex", alignItems: "center", justifyContent: "center", zIndex: 1000,
  };
  const dialog: React.CSSProperties = {
    background: "#1a1d27", border: "1px solid #2a2e3a", borderRadius: "12px",
    padding: "24px", width: "480px", maxHeight: "90vh", overflowY: "auto",
  };
  const label: React.CSSProperties = { fontSize: "12px", color: "#71717a", display: "block", marginBottom: "4px" };
  const input: React.CSSProperties = {
    width: "100%", background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px",
    color: "#e4e4e7", padding: "6px 10px", fontSize: "13px", outline: "none", boxSizing: "border-box",
  };
  const row: React.CSSProperties = { marginBottom: "14px" };

  return (
    <div style={overlay} onClick={(e) => e.target === e.currentTarget && onClose()}>
      <div style={dialog}>
        <h3 style={{ margin: "0 0 20px", fontSize: "16px", color: "#f4f4f5" }}>
          {monitor?.id ? "Edit Monitor" : "New Monitor"}
        </h3>
        <form onSubmit={handleSubmit}>
          <div style={row}>
            <label style={label}>Name *</label>
            <input style={input} value={form.name ?? ""} onChange={(e) => set("name", e.target.value)} required />
          </div>
          <div style={row}>
            <label style={label}>Type</label>
            <select style={{ ...input, cursor: "pointer" }} value={form.type ?? "http"} onChange={(e) => set("type", e.target.value as MonitorType)}>
              <option value="http">HTTP(S)</option>
              <option value="tcp">TCP Port</option>
              <option value="ping">Ping</option>
            </select>
          </div>
          {form.type === "http" && (
            <>
              <div style={row}>
                <label style={label}>URL *</label>
                <input style={input} value={form.url ?? ""} onChange={(e) => set("url", e.target.value)} placeholder="https://example.com" required />
              </div>
              <div style={{ display: "flex", gap: "12px", marginBottom: "14px" }}>
                <div style={{ flex: 1 }}>
                  <label style={label}>Method</label>
                  <select style={{ ...input, cursor: "pointer" }} value={form.method ?? "GET"} onChange={(e) => set("method", e.target.value)}>
                    <option>GET</option>
                    <option>POST</option>
                    <option>HEAD</option>
                  </select>
                </div>
                <div style={{ flex: 1 }}>
                  <label style={label}>Expected Status</label>
                  <input style={input} type="number" value={form.expected_status ?? 200} onChange={(e) => set("expected_status", Number(e.target.value))} />
                </div>
              </div>
              <div style={row}>
                <label style={label}>Keyword (optional body check)</label>
                <input style={input} value={form.keyword ?? ""} onChange={(e) => set("keyword", e.target.value)} placeholder="e.g. OK" />
              </div>
            </>
          )}
          {(form.type === "tcp" || form.type === "ping") && (
            <div style={{ display: "flex", gap: "12px", marginBottom: "14px" }}>
              <div style={{ flex: 1 }}>
                <label style={label}>Host *</label>
                <input style={input} value={form.host ?? ""} onChange={(e) => set("host", e.target.value)} placeholder="192.168.1.1" required />
              </div>
              {form.type === "tcp" && (
                <div style={{ width: "100px" }}>
                  <label style={label}>Port *</label>
                  <input style={input} type="number" value={form.port ?? 80} onChange={(e) => set("port", Number(e.target.value))} required />
                </div>
              )}
            </div>
          )}
          <div style={{ display: "flex", gap: "12px", marginBottom: "14px" }}>
            <div style={{ flex: 1 }}>
              <label style={label}>Interval (sec)</label>
              <input style={input} type="number" min={10} value={form.interval ?? 60} onChange={(e) => set("interval", Number(e.target.value))} />
            </div>
            <div style={{ flex: 1 }}>
              <label style={label}>Timeout (sec)</label>
              <input style={input} type="number" min={1} value={form.timeout ?? 10} onChange={(e) => set("timeout", Number(e.target.value))} />
            </div>
          </div>
          <div style={row}>
            <label style={label}>Notification Webhook (Discord/Slack URL)</label>
            <input style={input} value={form.notify_webhook ?? ""} onChange={(e) => set("notify_webhook", e.target.value)} placeholder="https://discord.com/api/webhooks/..." />
          </div>
          <div style={{ ...row, display: "flex", alignItems: "center", gap: "8px" }}>
            <input type="checkbox" id="active" checked={form.active ?? true} onChange={(e) => set("active", e.target.checked)} style={{ cursor: "pointer" }} />
            <label htmlFor="active" style={{ ...label, marginBottom: 0, cursor: "pointer" }}>Active (start monitoring immediately)</label>
          </div>
          {error && <p style={{ color: "#ef4444", fontSize: "12px", marginTop: "8px" }}>{error}</p>}
          <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end", marginTop: "20px" }}>
            <button type="button" onClick={onClose} style={{ padding: "7px 16px", borderRadius: "6px", border: "1px solid #2a2e3a", background: "transparent", color: "#a1a1aa", cursor: "pointer", fontSize: "13px" }}>
              Cancel
            </button>
            <button type="submit" disabled={saving} style={{ padding: "7px 16px", borderRadius: "6px", border: "none", background: "#2dd4bf", color: "#0f1117", cursor: saving ? "not-allowed" : "pointer", fontSize: "13px", fontWeight: 600, opacity: saving ? 0.7 : 1 }}>
              {saving ? "Saving…" : "Save"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

export function Monitors() {
  const { monitors, loading, fetch, create, update, remove } = useMonitorStore();
  const [modal, setModal] = useState<Partial<Monitor> | null | "new">(null);
  const navigate = useNavigate();

  useEffect(() => { fetch(); }, []);

  const handleSave = async (data: Partial<Monitor>) => {
    if (modal && typeof modal === "object" && "id" in modal && modal.id) {
      await update(modal.id, data);
    } else {
      await create(data);
    }
  };

  if (loading && monitors.length === 0) {
    return (
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        Loading…
      </div>
    );
  }

  return (
    <div style={{ flex: 1, overflowY: "auto", padding: "24px", backgroundColor: "#0f1117" }}>
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "20px" }}>
        <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
          <h1 style={{ margin: 0, fontSize: "18px", fontWeight: 600, color: "#f4f4f5" }}>Monitors</h1>
          {monitors.length > 0 && (
            <span style={{ fontSize: "12px", color: "#71717a", backgroundColor: "#1e2130", padding: "2px 8px", borderRadius: "12px", border: "1px solid #2a2e3a" }}>
              {monitors.length}
            </span>
          )}
        </div>
        <div style={{ display: "flex", gap: "8px" }}>
          <a
            href="/status"
            target="_blank"
            rel="noreferrer"
            style={{ padding: "6px 12px", borderRadius: "6px", border: "1px solid #2a2e3a", background: "transparent", color: "#71717a", cursor: "pointer", fontSize: "13px", textDecoration: "none", display: "flex", alignItems: "center" }}
          >
            Public Status Page ↗
          </a>
          <button
            onClick={() => setModal("new")}
            style={{ display: "flex", alignItems: "center", gap: "6px", padding: "6px 12px", borderRadius: "6px", border: "none", background: "#2dd4bf", color: "#0f1117", cursor: "pointer", fontSize: "13px", fontWeight: 600 }}
          >
            <Plus size={14} /> Add Monitor
          </button>
        </div>
      </div>

      {monitors.length === 0 ? (
        <div style={{ display: "flex", flexDirection: "column", alignItems: "center", justifyContent: "center", gap: "12px", padding: "80px 32px", color: "#71717a", textAlign: "center" }}>
          <Activity size={36} strokeWidth={1} color="#2a2e3a" />
          <p style={{ margin: 0, fontSize: "14px", lineHeight: "1.6", maxWidth: "320px" }}>
            No monitors yet — add an HTTP, TCP, or ping monitor to track uptime.
          </p>
          <button
            onClick={() => setModal("new")}
            style={{ marginTop: "8px", padding: "8px 16px", borderRadius: "6px", border: "none", background: "#2dd4bf", color: "#0f1117", cursor: "pointer", fontSize: "13px", fontWeight: 600 }}
          >
            Add your first monitor
          </button>
        </div>
      ) : (
        <div style={{ backgroundColor: "#1e2130", borderRadius: "10px", border: "1px solid #2a2e3a", overflow: "hidden" }}>
          {monitors.map((m, i) => (
            <div
              key={m.id}
              onClick={() => navigate(`/monitors/${m.id}`)}
              style={{ display: "flex", alignItems: "center", gap: "12px", padding: "14px 16px", borderBottom: i < monitors.length - 1 ? "1px solid #2a2e3a" : "none", cursor: "pointer" }}
              onMouseEnter={(e) => (e.currentTarget.style.background = "#242838")}
              onMouseLeave={(e) => (e.currentTarget.style.background = "transparent")}
            >
              {/* Status dot */}
              <div style={{ width: 8, height: 8, borderRadius: "50%", backgroundColor: statusColor(m.status), flexShrink: 0 }} />

              {/* Name + type */}
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontSize: "14px", color: "#f4f4f5", fontWeight: 500, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                  {m.name}
                </div>
                <div style={{ fontSize: "12px", color: "#52525b", marginTop: "2px", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
                  {m.type.toUpperCase()} · {m.url || m.host || "—"}
                </div>
              </div>

              {/* Status */}
              <span style={{ fontSize: "11px", fontWeight: 600, color: statusColor(m.status), width: "48px", textAlign: "center" }}>
                {m.status.toUpperCase()}
              </span>

              {/* Uptime */}
              <div style={{ width: "80px", textAlign: "right" }}>
                <div style={{ fontSize: "11px", color: "#52525b" }}>24h</div>
                {uptimeBadge(m.uptime_day)}
              </div>
              <div style={{ width: "80px", textAlign: "right" }}>
                <div style={{ fontSize: "11px", color: "#52525b" }}>7d</div>
                {uptimeBadge(m.uptime_week)}
              </div>

              {/* Actions */}
              <div style={{ display: "flex", gap: "4px" }} onClick={(e) => e.stopPropagation()}>
                <button
                  onClick={() => setModal(m)}
                  style={{ padding: "5px", borderRadius: "5px", border: "none", background: "transparent", color: "#52525b", cursor: "pointer" }}
                  title="Edit"
                  onMouseEnter={(e) => (e.currentTarget.style.color = "#a1a1aa")}
                  onMouseLeave={(e) => (e.currentTarget.style.color = "#52525b")}
                >
                  <Edit2 size={13} />
                </button>
                <button
                  onClick={async () => { if (confirm(`Delete monitor "${m.name}"?`)) await remove(m.id); }}
                  style={{ padding: "5px", borderRadius: "5px", border: "none", background: "transparent", color: "#52525b", cursor: "pointer" }}
                  title="Delete"
                  onMouseEnter={(e) => (e.currentTarget.style.color = "#ef4444")}
                  onMouseLeave={(e) => (e.currentTarget.style.color = "#52525b")}
                >
                  <Trash2 size={13} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {modal !== null && (
        <MonitorModal
          monitor={modal === "new" ? null : modal}
          onClose={() => setModal(null)}
          onSave={handleSave}
        />
      )}
    </div>
  );
}
