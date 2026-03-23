import type { StylesheetStyle } from "cytoscape";

export const mapStylesheet: StylesheetStyle[] = [
  {
    selector: "node",
    style: {
      label: "data(label)",
      "text-valign": "bottom",
      "text-margin-y": 8,
      "font-size": 11,
      "font-family": "Inter, system-ui, sans-serif",
      color: "#a1a1aa",
      "background-color": "#1a1d27",
      "border-width": 2,
      "border-color": "#2dd4bf",
      width: 40,
      height: 40,
      "overlay-padding": 6,
    },
  },
  {
    selector: "node[status = 'offline']",
    style: {
      "border-color": "#ef4444",
      opacity: 0.4,
    },
  },
  {
    selector: "node[status = 'unknown']",
    style: {
      "border-color": "#71717a",
      opacity: 0.6,
    },
  },
  {
    selector: "node:selected",
    style: {
      "border-color": "#8b5cf6",
      "border-width": 3,
      "background-color": "#242836",
    },
  },
  {
    selector: "node.highlighted",
    style: {
      "border-width": 3,
    },
  },
  {
    selector: "node.dimmed",
    style: {
      opacity: 0.15,
    },
  },
  {
    selector: "edge",
    style: {
      width: 1,
      "line-color": "#2dd4bf33",
      "curve-style": "bezier",
      "target-arrow-shape": "none",
    },
  },
  {
    selector: "edge.dimmed",
    style: {
      opacity: 0.05,
    },
  },
  {
    selector: "edge[status = 'offline']",
    style: {
      "line-color": "#ef444433",
      "line-style": "dashed",
    },
  },
];
