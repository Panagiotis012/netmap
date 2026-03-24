import { BrowserRouter, Routes, Route } from "react-router-dom";
import { useEffect } from "react";
import { TopNav } from "./components/TopNav/TopNav";
import { NetworkMap } from "./components/NetworkMap/NetworkMap";
import { DevicePanel } from "./components/DevicePanel/DevicePanel";
import { CommandPalette } from "./components/CommandPalette/CommandPalette";
import { ToastProvider } from "./components/Toast/ToastProvider";
import { useDeviceStore } from "./stores/deviceStore";
import { useUIStore } from "./stores/uiStore";
import { wsClient } from "./lib/ws";
import type { Device } from "./lib/types";
import { Scans } from "./pages/Scans";
import { DeviceList } from "./pages/DeviceList";
import { Settings } from "./pages/Settings";
import { Alerts } from "./pages/Alerts";
import { Networks } from "./pages/settings/Networks";
import { Scanning } from "./pages/settings/Scanning";
import { General } from "./pages/settings/General";

function MapView() {
  const panelOpen = useUIStore((s) => s.panelOpen);
  return (
    <div style={{ display: "flex", height: "100%", flex: 1 }}>
      <div style={{ flex: 1 }}>
        <NetworkMap />
      </div>
      {panelOpen && (
        <div style={{ width: "320px", borderLeft: "1px solid #2a2e3a" }}>
          <DevicePanel />
        </div>
      )}
    </div>
  );
}


export default function App() {
  const fetchDevices = useDeviceStore((s) => s.fetch);
  const upsert = useDeviceStore((s) => s.upsert);

  useEffect(() => {
    fetchDevices();
    wsClient.connect();

    const unsub1 = wsClient.on("device.discovered", (e) => upsert(e.payload as Device));

    return () => {
      unsub1();
      wsClient.disconnect();
    };
  }, []);

  return (
    <ToastProvider>
      <BrowserRouter>
        <div style={{ height: "100vh", display: "flex", flexDirection: "column", backgroundColor: "#0f1117" }}>
          <TopNav />
          <main style={{ flex: 1, overflow: "hidden", display: "flex" }}>
            <Routes>
              <Route path="/" element={<MapView />} />
              <Route path="/devices" element={<DeviceList />} />
              <Route path="/scans" element={<Scans />} />
              <Route path="/alerts" element={<Alerts />} />
              <Route path="/settings" element={<Settings />}>
                <Route index element={<Networks />} />
                <Route path="scanning" element={<Scanning />} />
                <Route path="general" element={<General />} />
              </Route>
            </Routes>
          </main>
          <CommandPalette />
        </div>
      </BrowserRouter>
    </ToastProvider>
  );
}
