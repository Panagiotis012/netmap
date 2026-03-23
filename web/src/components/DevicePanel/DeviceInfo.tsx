import { Copy, Check } from "lucide-react";
import { useState } from "react";

interface Props {
  label: string;
  value: string;
  mono?: boolean;
}

export function DeviceInfo({ label, value, mono }: Props) {
  const [copied, setCopied] = useState(false);

  const copy = () => {
    navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 1500);
  };

  return (
    <div style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "6px 0" }}>
      <span style={{ color: "#71717a", fontSize: "12px" }}>{label}</span>
      <button
        onClick={copy}
        style={{ display: "flex", alignItems: "center", gap: "4px", fontSize: "12px", fontFamily: mono ? "monospace" : "inherit", background: "none", border: "none", color: "#a1a1aa", cursor: "pointer" }}
      >
        {value}
        {copied ? <Check size={12} color="#2dd4bf" /> : <Copy size={12} style={{ opacity: 0.4 }} />}
      </button>
    </div>
  );
}
