import { useEffect } from "react";
import { NavLink } from "react-router-dom";
import { Map, Monitor, Bell, Radar, Settings, Search, LogOut } from "lucide-react";
import { useDeviceStore } from "../../stores/deviceStore";
import { useUIStore } from "../../stores/uiStore";
import { useNetworkStore } from "../../stores/networkStore";
import { useAlertsStore } from "../../stores/alertsStore";
import { useAuthStore } from "../../stores/authStore";
import { StatusBadge } from "./StatusBadge";
import { ScanButton } from "../ScanButton/ScanButton";
import { ScanPopover } from "../ScanButton/ScanPopover";
import { useScanProgress } from "../ScanButton/useScanProgress";

const navItems = [
  { to: "/", icon: Map, label: "Map" },
  { to: "/devices", icon: Monitor, label: "Devices" },
  { to: "/scans", icon: Radar, label: "Scans" },
  { to: "/alerts", icon: Bell, label: "Alerts" },
  { to: "/settings", icon: Settings, label: "Settings" },
];

export function TopNav() {
  const devices = useDeviceStore((s) => s.devices);
  const toggleCommandPalette = useUIStore((s) => s.toggleCommandPalette);
  const fetchNetworks = useNetworkStore((s) => s.fetch);
  const unread = useAlertsStore((s) => s.unread);
  const { setup, logout } = useAuthStore();

  useScanProgress();

  useEffect(() => {
    fetchNetworks();
  }, []);

  const online = devices.filter((d) => d.status === "online").length;
  const offline = devices.filter((d) => d.status === "offline").length;

  return (
    <div style={{ position: "relative" }}>
      <nav style={{ height: "48px", backgroundColor: "#1a1d27", borderBottom: "1px solid #2a2e3a", display: "flex", alignItems: "center", padding: "0 16px", gap: "4px", flexShrink: 0 }}>
        <span style={{ color: "#2dd4bf", fontWeight: 600, fontSize: "15px", marginRight: "24px", letterSpacing: "-0.02em" }}>
          NetMap
        </span>

        {navItems.map(({ to, icon: Icon, label }) => (
          <NavLink
            key={to}
            to={to}
            end={to === "/"}
            style={({ isActive }) => ({
              display: "flex", alignItems: "center", gap: "6px",
              padding: "6px 12px", borderRadius: "6px", fontSize: "14px",
              textDecoration: "none", transition: "all 0.15s",
              backgroundColor: isActive ? "rgba(45,212,191,0.1)" : "transparent",
              color: isActive ? "#2dd4bf" : "#a1a1aa",
            })}
          >
            {label === "Alerts" ? (
              <span style={{ position: "relative", display: "inline-flex", alignItems: "center" }}>
                <Icon size={16} strokeWidth={1.5} />
                {unread > 0 && (
                  <span style={{
                    position: "absolute",
                    top: "-5px",
                    right: "-6px",
                    minWidth: "14px",
                    height: "14px",
                    backgroundColor: "#ef4444",
                    borderRadius: "7px",
                    fontSize: "9px",
                    fontWeight: 700,
                    color: "#fff",
                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                    padding: "0 3px",
                    lineHeight: 1,
                    pointerEvents: "none",
                  }}>
                    {unread > 99 ? "99+" : unread}
                  </span>
                )}
              </span>
            ) : (
              <Icon size={16} strokeWidth={1.5} />
            )}
            {label}
          </NavLink>
        ))}

        <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: "8px" }}>
          <button
            onClick={() => toggleCommandPalette(true)}
            style={{ display: "flex", alignItems: "center", gap: "6px", padding: "4px 10px", borderRadius: "6px", background: "transparent", border: "none", color: "#71717a", cursor: "pointer", fontSize: "12px" }}
          >
            <Search size={14} strokeWidth={1.5} />
            <kbd style={{ fontSize: "10px", background: "#0f1117", padding: "1px 4px", borderRadius: "3px", border: "1px solid #2a2e3a" }}>
              Cmd+K
            </kbd>
          </button>
          <ScanButton />
          <StatusBadge count={online} label="up" color="teal" />
          <StatusBadge count={offline} label="dn" color="red" />
          {setup && (
            <button
              onClick={logout}
              title="Sign out"
              style={{ display: "flex", alignItems: "center", padding: "4px 6px", borderRadius: "6px", background: "transparent", border: "none", color: "#71717a", cursor: "pointer" }}
            >
              <LogOut size={14} strokeWidth={1.5} />
            </button>
          )}
        </div>
      </nav>
      <ScanPopover />
    </div>
  );
}
