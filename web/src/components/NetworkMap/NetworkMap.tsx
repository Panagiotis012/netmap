import { useEffect, useRef } from "react";
import cytoscape from "cytoscape";
import { useDeviceStore } from "../../stores/deviceStore";
import { mapStylesheet } from "./mapStyles";
import { useMapInteractions } from "./useMapInteractions";

export function NetworkMap() {
  const containerRef = useRef<HTMLDivElement>(null);
  const cyRef = useRef<cytoscape.Core | null>(null);
  const devices = useDeviceStore((s) => s.devices);
  const { onNodeTap, onNodeDrag } = useMapInteractions();

  useEffect(() => {
    if (!containerRef.current) return;

    const cy = cytoscape({
      container: containerRef.current,
      style: mapStylesheet,
      layout: { name: "preset" },
      userZoomingEnabled: true,
      userPanningEnabled: true,
      boxSelectionEnabled: false,
      minZoom: 0.3,
      maxZoom: 3,
    });

    cyRef.current = cy;
    onNodeTap(cy);
    onNodeDrag(cy);

    return () => { cy.destroy(); };
  }, []);

  useEffect(() => {
    const cy = cyRef.current;
    if (!cy) return;

    const existingIds = new Set(cy.nodes().map((n) => n.id()));
    const deviceIds = new Set(devices.map((d) => d.id));

    devices.forEach((device) => {
      if (!existingIds.has(device.id)) {
        cy.add({
          group: "nodes",
          data: {
            id: device.id,
            label: device.hostname || device.ip_addresses[0] || device.id.slice(0, 8),
            status: device.status,
          },
          position: {
            x: device.map_x ?? 200 + Math.random() * 600,
            y: device.map_y ?? 200 + Math.random() * 400,
          },
        });
      } else {
        const node = cy.getElementById(device.id);
        node.data("status", device.status);
        node.data("label", device.hostname || device.ip_addresses[0] || device.id.slice(0, 8));
      }
    });

    existingIds.forEach((id) => {
      if (!deviceIds.has(id)) {
        cy.getElementById(id).remove();
      }
    });

    if (devices.length > 1) {
      const center = devices[0];
      devices.slice(1).forEach((d) => {
        const edgeId = `${center.id}-${d.id}`;
        if (!cy.getElementById(edgeId).length) {
          cy.add({
            group: "edges",
            data: { id: edgeId, source: center.id, target: d.id, status: d.status },
          });
        }
      });
    }
  }, [devices]);

  return (
    <div
      ref={containerRef}
      style={{ width: "100%", height: "100%", background: "#0f1117" }}
    />
  );
}
