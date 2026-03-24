import { NavLink, Outlet } from "react-router-dom";

const sidebarItems = [
  { to: "/settings", label: "Networks", end: true },
  { to: "/settings/scanning", label: "Scanning" },
  { to: "/settings/general", label: "General" },
];

export function Settings() {
  return (
    <div style={{ display: "flex", flex: 1, overflow: "hidden" }}>
      {/* Sidebar */}
      <div style={{ width: "160px", background: "#1a1d27", borderRight: "1px solid #2a2e3a", padding: "20px 0", flexShrink: 0 }}>
        <div style={{ fontSize: "10px", fontWeight: 600, color: "#71717a", textTransform: "uppercase", letterSpacing: "0.06em", padding: "0 16px", marginBottom: "8px" }}>
          Settings
        </div>
        {sidebarItems.map(({ to, label, end }) => (
          <NavLink
            key={to}
            to={to}
            end={end}
            style={({ isActive }) => ({
              display: "block",
              padding: "7px 16px",
              fontSize: "13px",
              textDecoration: "none",
              color: isActive ? "#2dd4bf" : "#a1a1aa",
              backgroundColor: isActive ? "rgba(45,212,191,0.08)" : "transparent",
              borderLeft: isActive ? "2px solid #2dd4bf" : "2px solid transparent",
            })}
          >
            {label}
          </NavLink>
        ))}
      </div>
      {/* Content */}
      <div style={{ flex: 1, overflow: "auto" }}>
        <Outlet />
      </div>
    </div>
  );
}
