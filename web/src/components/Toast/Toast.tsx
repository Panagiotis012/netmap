import { motion } from "framer-motion";
import { X, CheckCircle, AlertTriangle, Info } from "lucide-react";

interface ToastProps {
  id: string;
  message: string;
  type: "success" | "warning" | "info" | "error";
  onDismiss: (id: string) => void;
}

const icons = { success: CheckCircle, warning: AlertTriangle, info: Info, error: AlertTriangle };
const colors = {
  success: { border: "rgba(45,212,191,0.3)", text: "#2dd4bf" },
  warning: { border: "rgba(245,158,11,0.3)", text: "#f59e0b" },
  info: { border: "rgba(59,130,246,0.3)", text: "#3b82f6" },
  error: { border: "rgba(239,68,68,0.3)", text: "#ef4444" },
};

export function Toast({ id, message, type, onDismiss }: ToastProps) {
  const Icon = icons[type];
  const color = colors[type];
  return (
    <motion.div
      layout
      initial={{ x: 400, opacity: 0 }}
      animate={{ x: 0, opacity: 1 }}
      exit={{ x: 400, opacity: 0 }}
      transition={{ type: "spring", stiffness: 500, damping: 30 }}
      style={{ display: "flex", alignItems: "center", gap: "8px", backgroundColor: "#1a1d27", border: `1px solid ${color.border}`, borderRadius: "8px", padding: "8px 12px", boxShadow: "0 10px 15px rgba(0,0,0,0.3)", minWidth: "280px" }}
    >
      <Icon size={16} color={color.text} />
      <span style={{ fontSize: "14px", color: "#e4e4e7", flex: 1 }}>{message}</span>
      <button onClick={() => onDismiss(id)} style={{ background: "none", border: "none", cursor: "pointer", color: "#71717a" }}>
        <X size={14} />
      </button>
    </motion.div>
  );
}
