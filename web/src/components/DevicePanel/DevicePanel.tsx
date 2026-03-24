import { X, Wifi, WifiOff, HelpCircle, Trash2 } from "lucide-react";
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
        <span style={{ color: "#52525b", fontSize: "12px" }}>
          Last seen {new Date(device.last_seen_at).toLocaleString()}
        </span>
      </div>

      <div style={{ marginBottom: "16px" }}>
        {device.ip_addresses.map((ip) => (
          <DeviceInfo key={ip} label="IP" value={ip} mono />
        ))}
        {device.mac_addresses.map((mac) => (
          <DeviceInfo key={mac} label="MAC" value={mac} mono />
        ))}
        {device.os && <DeviceInfo label="OS" value={device.os} />}
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
