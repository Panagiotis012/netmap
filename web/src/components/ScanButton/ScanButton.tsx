import { useNetworkStore } from "../../stores/networkStore";
import { useScanStore } from "../../stores/scanStore";
import { useToast } from "../Toast/ToastProvider";

export function ScanButton() {
  const networks = useNetworkStore((s) => s.networks);
  const scanning = useScanStore((s) => s.scanning);
  const activeScan = useScanStore((s) => s.activeScan);
  const startScan = useScanStore((s) => s.startScan);
  const { toast } = useToast();

  const firstNetwork = networks[0];
  const disabled = scanning || activeScan !== null || !firstNetwork;

  const handleClick = async () => {
    if (!firstNetwork) return;
    try {
      await startScan(firstNetwork.subnet);
    } catch {
      toast("Σφάλμα κατά την εκκίνηση scan", "error");
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={disabled}
      title={!firstNetwork ? "Πρόσθεσε δίκτυο στις Settings" : undefined}
      style={{
        display: "flex",
        alignItems: "center",
        gap: "6px",
        padding: "5px 12px",
        borderRadius: "6px",
        border: "none",
        cursor: disabled ? "not-allowed" : "pointer",
        fontSize: "13px",
        fontWeight: 600,
        backgroundColor: disabled ? "#2a2e3a" : "#2dd4bf",
        color: disabled ? "#71717a" : "#0f1117",
        transition: "all 0.15s",
      }}
    >
      {scanning ? "Scanning..." : "⚡ Scan Now"}
    </button>
  );
}
