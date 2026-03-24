import { useEffect, useState } from "react";
import { useNetworkStore } from "../../stores/networkStore";
import { ConfirmModal } from "../../components/ui/ConfirmModal";
import { api } from "../../lib/api";
import type { Network } from "../../lib/types";
import { useToast } from "../../components/Toast/ToastProvider";

interface NetworkFormData {
  name: string;
  subnet: string;
  gateway: string;
}

function isValidCIDR(value: string): boolean {
  return /^\d{1,3}(\.\d{1,3}){3}\/\d{1,2}$/.test(value);
}

export function Networks() {
  const networks = useNetworkStore((s) => s.networks);
  const fetchNetworks = useNetworkStore((s) => s.fetch);
  const { toast } = useToast();

  const [form, setForm] = useState<NetworkFormData>({ name: "", subnet: "", gateway: "" });
  const [cidrError, setCidrError] = useState("");
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editForm, setEditForm] = useState<NetworkFormData>({ name: "", subnet: "", gateway: "" });
  const [confirmDelete, setConfirmDelete] = useState<Network | null>(null);

  useEffect(() => { fetchNetworks(); }, [fetchNetworks]);

  const handleAdd = async () => {
    if (!isValidCIDR(form.subnet)) {
      setCidrError("Invalid CIDR (e.g. 192.168.1.0/24)");
      return;
    }
    setCidrError("");
    try {
      await api.networks.create(form);
      setForm({ name: "", subnet: "", gateway: "" });
      fetchNetworks();
    } catch {
      toast("Failed to add network", "error");
    }
  };

  const handleSaveEdit = async (id: string) => {
    try {
      await api.networks.update(id, editForm);
      setEditingId(null);
      fetchNetworks();
    } catch {
      toast("Failed to update network", "error");
    }
  };

  const handleDelete = async (network: Network) => {
    try {
      await api.networks.delete(network.id);
      setConfirmDelete(null);
      fetchNetworks();
    } catch {
      toast("Failed to delete network", "error");
    }
  };

  const inputStyle: React.CSSProperties = {
    background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px",
    color: "#e4e4e7", padding: "6px 10px", fontSize: "13px",
  };

  return (
    <div style={{ padding: "24px" }}>
      <h2 style={{ color: "#f4f4f5", margin: "0 0 20px", fontSize: "16px" }}>Networks</h2>

      {/* Network list */}
      <div style={{ marginBottom: "24px" }}>
        {networks.map((n) => (
          <div key={n.id} style={{ background: "#0f1117", borderRadius: "6px", padding: "10px 14px", marginBottom: "6px", display: "flex", alignItems: "center", gap: "10px" }}>
            {editingId === n.id ? (
              <>
                <input value={editForm.name} onChange={(e) => setEditForm({ ...editForm, name: e.target.value })} style={{ ...inputStyle, width: "120px" }} placeholder="Name" />
                <input value={editForm.subnet} onChange={(e) => setEditForm({ ...editForm, subnet: e.target.value })} style={{ ...inputStyle, width: "140px" }} placeholder="Subnet" />
                <input value={editForm.gateway} onChange={(e) => setEditForm({ ...editForm, gateway: e.target.value })} style={{ ...inputStyle, width: "120px" }} placeholder="Gateway" />
                <button onClick={() => handleSaveEdit(n.id)} style={{ padding: "5px 10px", borderRadius: "5px", border: "none", background: "#2dd4bf", color: "#0f1117", cursor: "pointer", fontSize: "12px", fontWeight: 600 }}>Save</button>
                <button onClick={() => setEditingId(null)} style={{ padding: "5px 10px", borderRadius: "5px", border: "1px solid #2a2e3a", background: "transparent", color: "#a1a1aa", cursor: "pointer", fontSize: "12px" }}>Cancel</button>
              </>
            ) : (
              <>
                <div style={{ flex: 1 }}>
                  <div style={{ fontSize: "13px", color: "#e4e4e7" }}>{n.name}</div>
                  <div style={{ fontSize: "11px", color: "#71717a" }}>{n.subnet}{n.gateway ? ` · gw: ${n.gateway}` : ""}</div>
                </div>
                <button onClick={() => { setEditingId(n.id); setEditForm({ name: n.name, subnet: n.subnet, gateway: n.gateway }); }} style={{ background: "none", border: "none", color: "#a1a1aa", cursor: "pointer", fontSize: "12px" }}>✏</button>
                <button onClick={() => setConfirmDelete(n)} style={{ background: "none", border: "none", color: "#ef4444", cursor: "pointer", fontSize: "12px" }}>🗑</button>
              </>
            )}
          </div>
        ))}
      </div>

      {/* Add form */}
      <div style={{ background: "#1a1d27", borderRadius: "8px", padding: "16px" }}>
        <div style={{ fontSize: "12px", color: "#a1a1aa", marginBottom: "10px", fontWeight: 500 }}>Add Network</div>
        <div style={{ display: "flex", gap: "8px", flexWrap: "wrap", alignItems: "flex-start" }}>
          <div>
            <input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} style={{ ...inputStyle, width: "140px" }} placeholder="Name (e.g. Home LAN)" />
          </div>
          <div>
            <input
              value={form.subnet}
              onChange={(e) => { setForm({ ...form, subnet: e.target.value }); setCidrError(""); }}
              onBlur={() => { if (form.subnet && !isValidCIDR(form.subnet)) setCidrError("Invalid CIDR"); }}
              style={{ ...inputStyle, width: "150px", borderColor: cidrError ? "#ef4444" : undefined }}
              placeholder="192.168.1.0/24"
            />
            {cidrError && <div style={{ fontSize: "11px", color: "#ef4444", marginTop: "2px" }}>{cidrError}</div>}
          </div>
          <input value={form.gateway} onChange={(e) => setForm({ ...form, gateway: e.target.value })} style={{ ...inputStyle, width: "130px" }} placeholder="Gateway (optional)" />
          <button
            onClick={handleAdd}
            disabled={!form.name || !form.subnet}
            style={{ padding: "7px 14px", borderRadius: "6px", border: "none", background: (!form.name || !form.subnet) ? "#2a2e3a" : "#2dd4bf", color: (!form.name || !form.subnet) ? "#71717a" : "#0f1117", cursor: (!form.name || !form.subnet) ? "not-allowed" : "pointer", fontSize: "13px", fontWeight: 600 }}
          >
            Add
          </button>
        </div>
      </div>

      <ConfirmModal
        open={confirmDelete !== null}
        title="Delete Network"
        message={`Delete "${confirmDelete?.name}" (${confirmDelete?.subnet})? This cannot be undone.`}
        confirmLabel="Delete"
        onConfirm={() => confirmDelete && handleDelete(confirmDelete)}
        onCancel={() => setConfirmDelete(null)}
      />
    </div>
  );
}
