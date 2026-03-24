interface ConfirmModalProps {
  open: boolean;
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  onConfirm: () => void;
  onCancel: () => void;
}

export function ConfirmModal({
  open, title, message,
  confirmLabel = "Confirm", cancelLabel = "Cancel",
  onConfirm, onCancel,
}: ConfirmModalProps) {
  if (!open) return null;

  return (
    <div
      onClick={onCancel}
      style={{
        position: "fixed", inset: 0, zIndex: 200,
        background: "rgba(0,0,0,0.6)",
        display: "flex", alignItems: "center", justifyContent: "center",
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: "#1e2130", border: "1px solid #2a2e3a", borderRadius: "10px",
          padding: "24px", minWidth: "320px", maxWidth: "440px",
        }}
      >
        <h3 style={{ margin: "0 0 8px", fontSize: "15px", color: "#f4f4f5" }}>{title}</h3>
        <p style={{ margin: "0 0 20px", fontSize: "13px", color: "#a1a1aa" }}>{message}</p>
        <div style={{ display: "flex", gap: "8px", justifyContent: "flex-end" }}>
          <button
            onClick={onCancel}
            style={{ padding: "6px 14px", borderRadius: "6px", border: "1px solid #2a2e3a", background: "transparent", color: "#a1a1aa", cursor: "pointer", fontSize: "13px" }}
          >
            {cancelLabel}
          </button>
          <button
            onClick={onConfirm}
            style={{ padding: "6px 14px", borderRadius: "6px", border: "none", background: "#ef4444", color: "#fff", cursor: "pointer", fontSize: "13px", fontWeight: 600 }}
          >
            {confirmLabel}
          </button>
        </div>
      </div>
    </div>
  );
}
