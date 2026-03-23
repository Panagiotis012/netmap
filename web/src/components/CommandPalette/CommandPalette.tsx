import { Command } from "cmdk";
import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Map, Monitor, Radar, Settings, Search } from "lucide-react";
import { useUIStore } from "../../stores/uiStore";
import { useDeviceStore } from "../../stores/deviceStore";
import { motion, AnimatePresence } from "framer-motion";

export function CommandPalette() {
  const open = useUIStore((s) => s.commandPaletteOpen);
  const toggle = useUIStore((s) => s.toggleCommandPalette);
  const devices = useDeviceStore((s) => s.devices);
  const selectDevice = useUIStore((s) => s.selectDevice);
  const navigate = useNavigate();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        toggle();
      }
      if (e.key === "Escape" && open) toggle(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, toggle]);

  if (!open) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        style={{ position: "fixed", inset: 0, zIndex: 50, display: "flex", alignItems: "flex-start", justifyContent: "center", paddingTop: "20vh" }}
        onClick={() => toggle(false)}
      >
        <div style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,0.5)" }} />
        <motion.div
          initial={{ scale: 0.95, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.95, opacity: 0 }}
          transition={{ type: "spring", stiffness: 500, damping: 30 }}
          style={{ position: "relative", width: "560px", backgroundColor: "#1a1d27", border: "1px solid #2a2e3a", borderRadius: "12px", boxShadow: "0 25px 50px rgba(0,0,0,0.5)", overflow: "hidden" }}
          onClick={(e) => e.stopPropagation()}
        >
          <Command>
            <div style={{ display: "flex", alignItems: "center", gap: "8px", padding: "0 12px", borderBottom: "1px solid #2a2e3a" }}>
              <Search size={16} color="#71717a" />
              <Command.Input
                placeholder="Search devices, navigate, trigger actions..."
                style={{ width: "100%", padding: "12px 0", background: "transparent", fontSize: "14px", color: "#e4e4e7", border: "none", outline: "none" }}
              />
            </div>
            <Command.List style={{ maxHeight: "320px", overflowY: "auto", padding: "8px" }}>
              <Command.Empty style={{ color: "#71717a", fontSize: "14px", textAlign: "center", padding: "24px" }}>
                No results found.
              </Command.Empty>
              <Command.Group heading="Navigation">
                {[
                  { label: "Map View", icon: Map, path: "/" },
                  { label: "Devices", icon: Monitor, path: "/devices" },
                  { label: "Scans", icon: Radar, path: "/scans" },
                  { label: "Settings", icon: Settings, path: "/settings" },
                ].map(({ label, icon: Icon, path }) => (
                  <Command.Item
                    key={path}
                    onSelect={() => { navigate(path); toggle(false); }}
                    style={{ display: "flex", alignItems: "center", gap: "8px", padding: "8px 12px", borderRadius: "6px", fontSize: "14px", color: "#d4d4d8", cursor: "pointer" }}
                  >
                    <Icon size={14} strokeWidth={1.5} />
                    {label}
                  </Command.Item>
                ))}
              </Command.Group>
              {devices.length > 0 && (
                <Command.Group heading="Devices">
                  {devices.slice(0, 20).map((d) => (
                    <Command.Item
                      key={d.id}
                      value={`${d.hostname} ${d.ip_addresses.join(" ")}`}
                      onSelect={() => { selectDevice(d.id); navigate("/"); toggle(false); }}
                      style={{ display: "flex", alignItems: "center", justifyContent: "space-between", padding: "8px 12px", borderRadius: "6px", fontSize: "14px", color: "#d4d4d8", cursor: "pointer" }}
                    >
                      <span>{d.hostname || d.ip_addresses[0]}</span>
                      <span style={{ fontSize: "12px", color: d.status === "online" ? "#2dd4bf" : "#ef4444" }}>
                        {d.status}
                      </span>
                    </Command.Item>
                  ))}
                </Command.Group>
              )}
            </Command.List>
          </Command>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
