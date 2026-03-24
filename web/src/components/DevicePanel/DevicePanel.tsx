import { X, Wifi, WifiOff, HelpCircle, Trash2, Pencil, Check } from "lucide-react";
import { useState } from "react";
import { motion } from "framer-motion";
import { useUIStore } from "../../stores/uiStore";
import { useDeviceStore } from "../../stores/deviceStore";
import { DeviceInfo } from "./DeviceInfo";
import { TagEditor } from "./TagEditor";
import { ConfirmModal } from "../ui/ConfirmModal";
import { api } from "../../lib/api";

const statusConfig = {
  online: { icon: Wifi, color: "#2dd4bf", label: "Online" },
  offline: { icon: WifiOff, color: "#ef4444", label: "Offline" },
  unknown: { icon: HelpCircle, color: "#71717a", label: "Unknown" },
};

function EditableField({ label, value, onSave }: { label: string; value: string; onSave: (v: string) => Promise<void> }) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);

  const commit = async () => {
    await onSave(draft);
    setEditing(false);
  };

  if (editing) {
    return (
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "6px 0" }}>
        <span style={{ color: "#71717a", fontSize: "12px" }}>{label}</span>
        <div style={{ display: "flex", gap: "4px", alignItems: "center" }}>
          <input
            autoFocus
            value={draft}
            onChange={(e) => setDraft(e.target.value)}
            onKeyDown={(e) => { if (e.key === "Enter") commit(); if (e.key === "Escape") { setDraft(value); setEditing(false); } }}
            style={{ background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "4px", color: "#e4e4e7", padding: "2px 6px", fontSize: "12px", width: "140px" }}
          />
          <button onClick={commit} style={{ background: "none", border: "none", cursor: "pointer", color: "#2dd4bf", padding: "2px" }}>
            <Check size={12} />
          </button>
        </div>
      </div>
    );
  }

  return (
    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "6px 0" }}>
      <span style={{ color: "#71717a", fontSize: "12px" }}>{label}</span>
      <div style={{ display: "flex", gap: "6px", alignItems: "center" }}>
        <span style={{ fontSize: "12px", color: "#a1a1aa" }}>{value || <span style={{ color: "#52525b" }}>—</span>}</span>
        <button onClick={() => { setDraft(value); setEditing(true); }} style={{ background: "none", border: "none", cursor: "pointer", color: "#52525b", padding: "2px" }}>
          <Pencil size={10} />
        </button>
      </div>
    </div>
  );
}

export function DevicePanel() {
  const selectedId = useUIStore((s) => s.selectedDeviceId);
  const selectDevice = useUIStore((s) => s.selectDevice);
  const devices = useDeviceStore((s) => s.devices);
  const upsert = useDeviceStore((s) => s.upsert);
  const remove = useDeviceStore((s) => s.remove);
  const [confirmDelete, setConfirmDelete] = useState(false);

  const device = devices.find((d) => d.id === selectedId);
  if (!device) return null;

  const status = statusConfig[device.status] ?? statusConfig.unknown;
  const StatusIcon = status.icon;

  const updateField = async (field: string, value: string) => {
    const updated = await api.devices.update(device.id, { [field]: value });
    upsert(updated);
  };

  const updateTags = async (tags: string[]) => {
    const updated = await api.devices.update(device.id, { tags });
    upsert(updated);
  };

  const handleDelete = async () => {
    await api.devices.delete(device.id);
    remove(device.id);
    selectDevice(null);
  };

  return (
    <motion.div
      initial={{ x: 320, opacity: 0 }}
      animate={{ x: 0, opacity: 1 }}
      exit={{ x: 320, opacity: 0 }}
      transition={{ type: "spring", stiffness: 400, damping: 30 }}
      style={{ height: "100%", backgroundColor: "#1a1d27", overflowY: "auto", padding: "16px" }}
    >
      <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", marginBottom: "16px" }}>
        <h3 style={{ margin: 0, fontWeight: 500, color: "#f4f4f5", fontSize: "14px" }}>
          {device.hostname || "Unknown Device"}
        </h3>
        <div style={{ display: "flex", gap: "4px" }}>
          <button onClick={() => setConfirmDelete(true)} title="Delete device" style={{ background: "none", border: "none", cursor: "pointer", color: "#71717a", padding: "2px" }}>
            <Trash2 size={14} />
          </button>
          <button onClick={() => selectDevice(null)} style={{ background: "none", border: "none", cursor: "pointer", color: "#71717a", padding: "2px" }}>
            <X size={16} />
          </button>
        </div>
      </div>

      <div style={{ display: "flex", alignItems: "center", gap: "8px", marginBottom: "16px" }}>
        <StatusIcon size={16} color={status.color} />
        <span style={{ fontSize: "14px", color: status.color }}>{status.label}</span>
        {device.latency_ms != null && device.latency_ms > 0 && (
          <span style={{ color: "#52525b", fontSize: "12px" }}>{device.latency_ms.toFixed(1)} ms</span>
        )}
        <span style={{ color: "#52525b", fontSize: "12px", marginLeft: "auto" }}>
          {new Date(device.last_seen_at).toLocaleString()}
        </span>
      </div>

      <div style={{ marginBottom: "16px" }}>
        {device.ip_addresses.map((ip) => (
          <DeviceInfo key={ip} label="IP" value={ip} mono />
        ))}
        {device.mac_addresses.map((mac) => (
          <DeviceInfo key={mac} label="MAC" value={mac} mono />
        ))}
        <EditableField label="Hostname" value={device.hostname} onSave={(v) => updateField("hostname", v)} />
        <EditableField label="OS" value={device.os} onSave={(v) => updateField("os", v)} />
        <DeviceInfo label="Discovered" value={device.discovery_method} />
        <DeviceInfo label="First seen" value={new Date(device.first_seen_at).toLocaleString()} />
      </div>

      {device.ports && device.ports.length > 0 && (
        <div style={{ marginBottom: "16px" }}>
          <span style={{ color: "#71717a", fontSize: "12px", display: "block", marginBottom: "6px" }}>
            Open Ports <span style={{ color: "#52525b" }}>({device.ports.length})</span>
          </span>
          <div style={{ display: "flex", flexWrap: "wrap", gap: "4px" }}>
            {device.ports.map((p) => (
              <span
                key={`${p.number}/${p.protocol}`}
                title={p.service || undefined}
                style={{
                  background: "#1e2130", border: "1px solid #2a2e3a", borderRadius: "4px",
                  padding: "2px 6px", fontSize: "11px", fontFamily: "monospace", color: "#a1a1aa",
                }}
              >
                {p.number}/{p.protocol}
                {p.service && <span style={{ color: "#71717a" }}> {p.service}</span>}
              </span>
            ))}
          </div>
        </div>
      )}

      <div>
        <span style={{ color: "#71717a", fontSize: "12px", display: "block", marginBottom: "6px" }}>Tags</span>
        <TagEditor tags={device.tags} onChange={updateTags} />
      </div>

      <ConfirmModal
        open={confirmDelete}
        title="Delete Device"
        message={`Remove "${device.hostname || device.ip_addresses[0]}" from NetMap? This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={handleDelete}
        onCancel={() => setConfirmDelete(false)}
      />
    </motion.div>
  );
}
