import { useCallback, useRef } from "react";
import type cytoscape from "cytoscape";
import { useUIStore } from "../../stores/uiStore";
import { useDeviceStore } from "../../stores/deviceStore";

export function useMapInteractions() {
  const selectDevice = useUIStore((s) => s.selectDevice);
  const updatePosition = useDeviceStore((s) => s.updatePosition);
  const saveTimeout = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  const onNodeTap = useCallback(
    (cy: cytoscape.Core) => {
      cy.on("tap", "node", (e) => {
        const node = e.target;
        selectDevice(node.id());
        cy.elements().removeClass("highlighted dimmed");
        const connected = node.connectedEdges().connectedNodes();
        cy.elements().not(connected).not(node).addClass("dimmed");
        node.connectedEdges().connectedNodes().addClass("highlighted");
      });

      cy.on("tap", (e) => {
        if (e.target === cy) {
          selectDevice(null);
          cy.elements().removeClass("highlighted dimmed");
        }
      });
    },
    [selectDevice]
  );

  const onNodeDrag = useCallback(
    (cy: cytoscape.Core) => {
      cy.on("dragfree", "node", (e) => {
        const node = e.target;
        const pos = node.position();
        if (saveTimeout.current) clearTimeout(saveTimeout.current);
        saveTimeout.current = setTimeout(() => {
          updatePosition(node.id(), pos.x, pos.y);
        }, 500);
      });
    },
    [updatePosition]
  );

  return { onNodeTap, onNodeDrag };
}
