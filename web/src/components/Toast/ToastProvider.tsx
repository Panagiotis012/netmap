import { createContext, useContext, useCallback, useState } from "react";
import { AnimatePresence } from "framer-motion";
import { Toast } from "./Toast";

interface ToastData {
  id: string;
  message: string;
  type: "success" | "warning" | "info" | "error";
}

interface ToastContextType {
  toast: (message: string, type?: ToastData["type"]) => void;
}

const ToastContext = createContext<ToastContextType>({ toast: () => {} });

export const useToast = () => useContext(ToastContext);

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastData[]>([]);

  const dismiss = useCallback((id: string) => {
    setToasts((t) => t.filter((x) => x.id !== id));
  }, []);

  const toast = useCallback(
    (message: string, type: ToastData["type"] = "info") => {
      const id = Math.random().toString(36).slice(2);
      setToasts((t) => [...t, { id, message, type }]);
      setTimeout(() => dismiss(id), 4000);
    },
    [dismiss]
  );

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div style={{ position: "fixed", top: "60px", right: "16px", zIndex: 50, display: "flex", flexDirection: "column", gap: "8px" }}>
        <AnimatePresence>
          {toasts.map((t) => (
            <Toast key={t.id} {...t} onDismiss={dismiss} />
          ))}
        </AnimatePresence>
      </div>
    </ToastContext.Provider>
  );
}
