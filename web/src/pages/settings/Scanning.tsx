import { useEffect, useState } from "react";
import { useConfigStore } from "../../stores/configStore";
import { useToast } from "../../components/Toast/ToastProvider";

const INTERVALS = ["1m", "5m", "15m", "1h", "off"] as const;

export function Scanning() {
  const config = useConfigStore();
  const { toast } = useToast();
  const [interval, setIntervalVal] = useState(config.scan_interval);
  const [workers, setWorkers] = useState(config.scan_workers);
  const [portRanges, setPortRanges] = useState(config.port_ranges);
  const [portError, setPortError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    config.fetch()
      .then(() => {
        const c = useConfigStore.getState();
        setIntervalVal(c.scan_interval);
        setWorkers(c.scan_workers);
        setPortRanges(c.port_ranges);
      })
      .catch(() => toast("Failed to load settings", "error"));
  }, []);

  const isValidPorts = (v: string) => /^\d+(,\d+)*$/.test(v);

  const handleSave = async () => {
    if (!isValidPorts(portRanges)) {
      setPortError("Format: comma-separated numbers (e.g. 22,80,443)");
      return;
    }
    setPortError("");
    setSaving(true);
    try {
      await config.save({ scan_interval: interval, scan_workers: workers, port_ranges: portRanges });
      toast("Settings saved", "success");
    } catch {
      toast("Failed to save settings", "error");
      // Restore from store
      const c = useConfigStore.getState();
      setIntervalVal(c.scan_interval);
      setWorkers(c.scan_workers);
      setPortRanges(c.port_ranges);
    } finally {
      setSaving(false);
    }
  };

  const inputStyle: React.CSSProperties = {
    background: "#0f1117", border: "1px solid #2a2e3a", borderRadius: "6px",
    color: "#e4e4e7", padding: "6px 10px", fontSize: "13px",
  };

  return (
    <div style={{ padding: "24px", maxWidth: "480px" }}>
      <h2 style={{ color: "#f4f4f5", margin: "0 0 24px", fontSize: "16px" }}>Scanning</h2>

      <div style={{ marginBottom: "20px" }}>
        <label style={{ display: "block", fontSize: "12px", color: "#71717a", marginBottom: "6px" }}>Scan Interval</label>
        <div style={{ display: "flex", gap: "4px" }}>
          {INTERVALS.map((v) => (
            <button
              key={v}
              onClick={() => setIntervalVal(v)}
              style={{
                padding: "5px 12px", borderRadius: "5px", fontSize: "12px", fontWeight: 500,
                border: interval === v ? "none" : "1px solid #2a2e3a",
                background: interval === v ? "#2dd4bf" : "transparent",
                color: interval === v ? "#0f1117" : "#a1a1aa",
                cursor: "pointer",
              }}
            >
              {v}
            </button>
          ))}
        </div>
      </div>

      <div style={{ marginBottom: "20px" }}>
        <label style={{ display: "block", fontSize: "12px", color: "#71717a", marginBottom: "6px" }}>Workers (1–200)</label>
        <input
          type="number"
          min={1}
          max={200}
          value={workers}
          onChange={(e) => setWorkers(Math.min(200, Math.max(1, parseInt(e.target.value) || 1)))}
          style={{ ...inputStyle, width: "80px" }}
        />
      </div>

      <div style={{ marginBottom: "24px" }}>
        <label style={{ display: "block", fontSize: "12px", color: "#71717a", marginBottom: "6px" }}>Ports (comma-separated)</label>
        <input
          type="text"
          value={portRanges}
          onChange={(e) => { setPortRanges(e.target.value); setPortError(""); }}
          onBlur={() => { if (!isValidPorts(portRanges)) setPortError("Format: comma-separated numbers"); }}
          style={{ ...inputStyle, width: "280px", borderColor: portError ? "#ef4444" : undefined }}
          placeholder="22,80,443,8080,8443"
        />
        {portError && <div style={{ fontSize: "11px", color: "#ef4444", marginTop: "3px" }}>{portError}</div>}
      </div>

      <button
        onClick={handleSave}
        disabled={saving}
        style={{ padding: "7px 18px", borderRadius: "6px", border: "none", background: saving ? "#2a2e3a" : "#2dd4bf", color: saving ? "#71717a" : "#0f1117", cursor: saving ? "not-allowed" : "pointer", fontWeight: 600, fontSize: "13px" }}
      >
        {saving ? "Saving..." : "Save"}
      </button>
    </div>
  );
}
