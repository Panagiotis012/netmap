import { useEffect } from "react";
import { NavLink } from "react-router-dom";
import { Map, Monitor, Bell, Radar, Settings, Search } from "lucide-react";
import { useDeviceStore } from "../../stores/deviceStore";
import { useUIStore } from "../../stores/uiStore";
import { useNetworkStore } from "../../stores/networkStore";
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
            <Icon size={16} strokeWidth={1.5} />
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
        </div>
      </nav>
      <ScanPopover />
    </div>
  );
}
