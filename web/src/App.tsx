import { BrowserRouter, Routes, Route } from "react-router-dom";
import { useEffect } from "react";
import { TopNav } from "./components/TopNav/TopNav";
import { DevicePanel } from "./components/DevicePanel/DevicePanel";
import { useDeviceStore } from "./stores/deviceStore";
import { useUIStore } from "./stores/uiStore";
import { wsClient } from "./lib/ws";
import type { Device } from "./lib/types";

function MapView() {
  const panelOpen = useUIStore((s) => s.panelOpen);
  return (
    <div style={{ display: "flex", height: "100%", flex: 1 }}>
      <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>
        Map View (Task 19)
      </div>
      {panelOpen && (
        <div style={{ width: "320px", borderLeft: "1px solid #2a2e3a" }}>
          <DevicePanel />
        </div>
      )}
    </div>
  );
}

function DeviceListPage() {
  return <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>Device List (Task 22)</div>;
}

function ScansPage() {
  return <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>Scans</div>;
}

function SettingsPage() {
  return <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#71717a" }}>Settings</div>;
}

export default function App() {
  const fetchDevices = useDeviceStore((s) => s.fetch);
  const upsert = useDeviceStore((s) => s.upsert);

  useEffect(() => {
    fetchDevices();
    wsClient.connect();

    const unsub1 = wsClient.on("device.discovered", (e) => upsert(e.payload as Device));
    const unsub2 = wsClient.on("device.updated", (e) => upsert(e.payload as Device));

    return () => {
      unsub1();
      unsub2();
      wsClient.disconnect();
    };
  }, []);

  return (
    <BrowserRouter>
      <div style={{ height: "100vh", display: "flex", flexDirection: "column", backgroundColor: "#0f1117" }}>
        <TopNav />
        <main style={{ flex: 1, overflow: "hidden", display: "flex" }}>
          <Routes>
            <Route path="/" element={<MapView />} />
            <Route path="/devices" element={<DeviceListPage />} />
            <Route path="/scans" element={<ScansPage />} />
            <Route path="/alerts" element={<div style={{ padding: "32px", color: "#71717a" }}>Alerts — Phase 2</div>} />
            <Route path="/settings" element={<SettingsPage />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}
