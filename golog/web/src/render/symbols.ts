// GoLog's own symbols for sedimentary structures, fossils and trace fossils, drawn for this
// app as simple line art in a 20 × 20 box. They follow the general look of field-log symbols
// but are not copied from any program or chart.

export type SymbolGroup = "Structures" | "Fossils" | "Trace fossils" | "Other";

export interface SymbolDef {
  id: string;
  name: string;
  group: SymbolGroup;
  body: string; // SVG inside a 0 0 20 20 box
  inBed?: boolean; // may be drawn repeated inside the grain-size profile
}

const S = `fill="none" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"`;
/** A path in the symbol style; `extra` attributes replace the style's own (e.g. a thinner line). */
const p = (d: string, extra = "") => {
  const own = new Set([...extra.matchAll(/([\w-]+)=/g)].map((m) => m[1]));
  const base = [...S.matchAll(/([\w-]+)="[^"]*"/g)].filter((m) => !own.has(m[1])).map((m) => m[0]).join(" ");
  return `<path d="${d}" ${base}${extra ? " " + extra : ""}/>`;
};

export const SYMBOLS: SymbolDef[] = [
  // ---- sedimentary structures ----
  { id: "planar-lamination", name: "Planar lamination", group: "Structures", inBed: true, body: p("M2 5h16M2 10h16M2 15h16") },
  { id: "low-angle-lamination", name: "Low-angle lamination", group: "Structures", inBed: true, body: p("M2 7l16-2M2 12l16-2M2 17l16-2") },
  { id: "planar-cross-bedding", name: "Planar cross-bedding", group: "Structures", inBed: true, body: p("M2 4h16M2 17h16M5 17l5-13M10 17l5-13M15 17l3-8") },
  { id: "trough-cross-bedding", name: "Trough cross-bedding", group: "Structures", inBed: true, body: p("M1 8c3 6 7 6 10 0M9 8c3 6 7 6 10 0M3 14c3 5 6 5 8 0M11 14c3 5 6 5 8 0") },
  { id: "current-ripples", name: "Current ripple lamination", group: "Structures", inBed: true, body: p("M1 14c3-5 5-6 6-6l-1 6M7 14c3-5 5-6 6-6l-1 6M13 14c3-5 5-6 6-6l-1 6") },
  { id: "wave-ripples", name: "Wave ripple lamination", group: "Structures", inBed: true, body: p("M1 12c2-4 3-4 4.5 0s3 4 4.5 0 3-4 4.5 0 3 4 4.5 0") },
  { id: "climbing-ripples", name: "Climbing ripples", group: "Structures", inBed: true, body: p("M1 17c2-3 3-4 5-4M6 17c2-3 3-4 5-4M4 11c2-3 3-4 5-4M9 11c2-3 3-4 5-4M8 5c2-3 3-4 5-4") },
  { id: "hummocky", name: "Hummocky cross-stratification", group: "Structures", inBed: true, body: p("M1 10c4-6 8-6 10 0s6 6 8 0M1 15c4-5 8-5 10 0s6 5 8 0") },
  { id: "swaley", name: "Swaley cross-stratification", group: "Structures", inBed: true, body: p("M1 7c3 6 7 6 9 0s6 6 9 0M1 13c3 5 7 5 9 0s6 5 9 0M1 17h18", 'stroke-width="1"') },
  { id: "herringbone", name: "Herringbone cross-bedding", group: "Structures", inBed: true, body: p("M2 9h16M2 17h16M4 9l3-6M9 9l3-6M14 9l3-6M7 17l-3-6M12 17l-3-6M17 17l-3-6") },
  { id: "flaser", name: "Flaser bedding", group: "Structures", inBed: true, body: p("M2 7c2-2 4-2 5 0M11 6c2-2 4-2 5 0M5 13c2-2 4-2 5 0M13 14c2-2 4-2 5 0") + p("M1 17h18", 'stroke-width="0.8"') },
  { id: "graded", name: "Normal grading", group: "Structures", body: p("M4 3l6 14 6-14z") },
  { id: "reverse-graded", name: "Reverse grading", group: "Structures", body: p("M4 17l6-14 6 14z") },
  { id: "massive", name: "Massive (structureless)", group: "Structures", body: p("M3 3h14v14H3z") },
  { id: "convolute", name: "Convolute lamination", group: "Structures", inBed: true, body: p("M1 15c3 0 3-9 6-9s2 6 4 6 2-8 5-8 3 6 3 6") },
  { id: "load-casts", name: "Load casts", group: "Structures", body: p("M1 8h18M2 8c0 6 5 6 5 0M8 8c0 7 5 7 5 0M14 8c0 5 4 5 4 0") },
  { id: "flute-casts", name: "Flute casts", group: "Structures", body: p("M1 7h18M3 7c0 5 4 6 6 6-1-3-2-5-2-6M11 7c0 5 4 6 6 6-1-3-2-5-2-6") },
  { id: "groove-casts", name: "Groove casts", group: "Structures", body: p("M1 7h18M3 7v6M3 13h14M7 7v4M11 7v6M15 7v4M17 7v6") },
  { id: "scours", name: "Scours", group: "Structures", body: p("M1 7h4c1 7 9 7 10 0h4M3 13h14", 'stroke-width="1"') },
  { id: "water-escape", name: "Water-escape structures", group: "Structures", body: p("M2 17c3-3 2-6 4-8s1-5 2-7M9 17c2-3 2-6 3-8s1-5 2-7M15 17c1-3 1-6 2-8") },
  { id: "mudcracks", name: "Desiccation cracks", group: "Structures", body: p("M2 5h16M5 5l1 6 -1 6M10 5l-1 4 1 4 0 4M15 5l1 7-1 5") },
  { id: "synaeresis-cracks", name: "Synaeresis cracks", group: "Structures", body: p("M2 6h16M5 6c1 3-1 5 0 8M10 6c-1 2 1 4 0 6M15 6c1 2-1 5 0 9") },
  { id: "rip-up-clasts", name: "Mud rip-up clasts", group: "Structures", body: p("M3 8c2-2 5-2 6 0-1 2-4 2-6 0zM11 13c2-2 5-2 6 0-1 2-4 2-6 0zM4 15c1-1 3-1 4 0-1 1-3 1-4 0z") },
  { id: "intraclasts", name: "Intraclasts", group: "Structures", body: p("M3 6l4-2 3 2-2 3-4 0zM11 12l4-2 3 3-3 3-3-1zM4 14l3-1 1 3-3 1z") },
  { id: "imbrication", name: "Clast imbrication", group: "Structures", body: p("M2 14l4-6 3 1-4 6zM8 14l4-6 3 1-4 6zM14 14l3-5 2 1-3 5z") },
  { id: "nodules", name: "Concretions / nodules", group: "Structures", body: `<ellipse cx="6" cy="8" rx="4" ry="2.6" ${S}/><ellipse cx="14" cy="13" rx="4" ry="2.6" ${S}/>` },
  { id: "columnar-joints", name: "Columnar jointing", group: "Structures", body: p("M3 18V4M8 18V2M13 18V4M17 18V3M3 4l5-2 5 2 4-1") },
  { id: "pillows", name: "Pillow structures", group: "Structures", body: p("M2 10c0-5 7-5 7 0 0 3-7 3-7 0zM10 15c0-5 8-5 8 0 0 3-8 3-8 0zM9 6c1-3 7-3 8 0") },
  { id: "vesicles", name: "Vesicles", group: "Structures", body: `<g ${S}><circle cx="5" cy="6" r="1.6"/><circle cx="12" cy="5" r="1.2"/><circle cx="15" cy="12" r="1.8"/><circle cx="7" cy="14" r="1.3"/></g>` },

  // ---- fossils ----
  { id: "bivalve", name: "Bivalves", group: "Fossils", body: p("M3 13c0-6 4-9 7-9s7 3 7 9c-3 2-11 2-14 0zM10 4v10M6 6l2 7M14 6l-2 7") },
  { id: "gastropod", name: "Gastropods", group: "Fossils", body: p("M10 2c-4 4-6 8-6 11a6 5 0 0 0 12 0c0-3-2-7-6-11zM6 9c3 1 5 1 8 0M5 13c3 1 7 1 10 0") },
  { id: "brachiopod", name: "Brachiopods", group: "Fossils", body: p("M10 3c-5 0-7 6-7 9 0 4 3 5 7 5s7-1 7-5c0-3-2-9-7-9zM10 3v4M8 17c0-2 4-2 4 0") },
  { id: "ammonite", name: "Ammonites", group: "Fossils", body: p("M10 10m-7 0a7 7 0 1 0 14 0 7 7 0 1 0-14 0M10 10m-3.5 0a3.5 3.5 0 1 0 7 0 3.5 3.5 0 1 0-7 0M10 3v3.5M17 10h-3.5M10 17v-3.5M3 10h3.5") },
  { id: "coral", name: "Corals", group: "Fossils", body: p("M6 18V9c0-3-2-4-2-6M6 9c0-2 2-3 2-5M13 18v-6c0-3 2-4 2-7M13 12c0-2-2-3-2-5") },
  { id: "bryozoa", name: "Bryozoa", group: "Fossils", body: p("M10 18V10M10 10l-5-5M10 10l5-5M5 5V2M5 5H2M15 5V2M15 5h3M10 14l-4-2M10 14l4-2") },
  { id: "sponge", name: "Sponges", group: "Fossils", body: p("M5 18c-1-5-1-10 1-15h8c2 5 2 10 1 15zM8 6v9M12 6v9M6 10h8") },
  { id: "stromatolite", name: "Stromatolites", group: "Fossils", body: p("M2 17h16M3 17c0-5 3-8 7-8s7 3 7 8M5 17c0-3 2-5 5-5s5 2 5 5M6 9c1-3 2-5 4-5s3 2 4 5") },
  { id: "algae", name: "Algae", group: "Fossils", body: p("M6 18c-2-4 2-6 0-10s2-6 0-8M12 18c2-4-2-6 0-10s-2-6 0-8M16 18c1-3-1-5 0-8") },
  { id: "crinoid", name: "Crinoid ossicles", group: "Fossils", body: `<g ${S}><circle cx="6" cy="7" r="3"/><circle cx="6" cy="7" r="0.6"/><circle cx="14" cy="13" r="3"/><circle cx="14" cy="13" r="0.6"/></g>` },
  { id: "echinoid", name: "Echinoids", group: "Fossils", body: p("M3 13c0-5 3-8 7-8s7 3 7 8zM10 5v8M6 7l2 6M14 7l-2 6") },
  { id: "foraminifera", name: "Foraminifera", group: "Fossils", body: `<g ${S}><circle cx="6" cy="12" r="2"/><circle cx="9.5" cy="9" r="2.6"/><circle cx="14" cy="7" r="3.2"/></g>` },
  { id: "ostracod", name: "Ostracods", group: "Fossils", body: p("M3 11c0-4 3-6 7-6s7 2 7 6-3 5-7 5-7-1-7-5zM5 9c3-2 7-2 10 0") },
  { id: "radiolaria", name: "Radiolaria", group: "Fossils", body: `<g ${S}><circle cx="10" cy="10" r="4.5"/></g>` + p("M10 2v3.5M10 14.5V18M2 10h3.5M14.5 10H18M4.3 4.3l2.5 2.5M13.2 13.2l2.5 2.5M15.7 4.3l-2.5 2.5M6.8 13.2l-2.5 2.5") },
  { id: "fish", name: "Fish remains", group: "Fossils", body: p("M2 10c4-5 10-5 13 0-3 5-9 5-13 0zM15 10l4-4v8zM6 9.5h0.1") },
  { id: "bone", name: "Bone fragments", group: "Fossils", body: p("M5 6l10 8M3 5c0-2 3-2 3 0 2 0 2 3 0 3M17 15c0 2-3 2-3 0-2 0-2-3 0-3") },
  { id: "plant", name: "Plant fragments", group: "Fossils", body: p("M3 16C7 13 12 8 17 3M7 13l-3-3M10 10l-1-4M10 10l4 1M13 7l1-3M13 7l3 1") },
  { id: "leaf", name: "Leaves", group: "Fossils", body: p("M4 16C4 9 9 4 16 4c0 7-5 12-12 12zM4 16l9-9M7 13h3M10 10V7") },
  { id: "wood", name: "Wood / logs", group: "Fossils", body: p("M2 8h16v5H2zM6 8v5M2 10.5c4 1 8-1 12 0") },
  { id: "tree-stump", name: "Tree stumps", group: "Fossils", body: p("M6 17V7h8v10M6 7c1-2 7-2 8 0M6 17l-3 2M14 17l3 2M10 17v2M8 10v5M12 9v6") },
  { id: "roots", name: "Roots", group: "Fossils", body: p("M10 2v6M10 8c-1 3-4 4-5 9M10 8c1 3 4 4 5 9M10 8v10M7 12l-3 1M13 12l3 1") },
  { id: "shell-debris", name: "Shell fragments", group: "Fossils", body: p("M3 6c2-2 4-2 5 0M11 9c2-2 4-2 5 0M4 13c2-2 4-2 5 0M12 16c2-2 4-2 5 0") },

  // ---- trace fossils ----
  { id: "vertical-burrows", name: "Vertical burrows", group: "Trace fossils", body: p("M5 3v14M10 3v14M15 3v14") + p("M4 3h2M9 3h2M14 3h2", 'stroke-width="1.6"') },
  { id: "u-burrows", name: "U-shaped burrows", group: "Trace fossils", body: p("M5 3v9a5 5 0 0 0 10 0V3") },
  { id: "horizontal-burrows", name: "Horizontal burrows", group: "Trace fossils", body: p("M2 7c4-2 6 2 10 0s5-1 6 0M2 14c3 1 6-2 9 0s5 1 7 0") },
  { id: "branching-burrows", name: "Branching burrow systems", group: "Trace fossils", body: p("M2 12h7M9 12l5-5M9 12l5 5M14 7h4M14 17h4M9 12v-6") },
  { id: "trails", name: "Trails and tracks", group: "Trace fossils", body: p("M2 10c3-3 5 3 8 0s5 3 8 0") + `<g fill="currentColor"><circle cx="4" cy="15" r="0.9"/><circle cx="8" cy="15" r="0.9"/><circle cx="12" cy="15" r="0.9"/><circle cx="16" cy="15" r="0.9"/></g>` },
  { id: "bioturbation", name: "Bioturbation (general)", group: "Trace fossils", body: p("M4 3c3 3-3 5 0 8s-3 5 0 7M10 3c3 3-3 5 0 8s-3 5 0 7M16 3c3 3-3 5 0 8s-3 5 0 7") },
  { id: "footprints", name: "Footprints", group: "Trace fossils", body: p("M5 15l-2-4M5 15l0-5M5 15l2-4M13 9l-2-4M13 9l0-5M13 9l2-4") },

  // ---- other ----
  { id: "pumice", name: "Pumice", group: "Other", body: `<g ${S}><path d="M3 8l4-3 5 1 1 5-4 3-5-2z"/><circle cx="7" cy="9" r="0.7"/><circle cx="10" cy="8" r="0.7"/><circle cx="8" cy="11.5" r="0.7"/></g>` },
  { id: "scoria", name: "Scoria", group: "Other", body: `<g ${S}><path d="M4 14l2-6 6-2 4 4-2 6-6 1z"/></g><g fill="currentColor"><circle cx="8" cy="10" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="9" cy="14" r="0.8"/></g>` },
  { id: "glass", name: "Volcanic glass", group: "Other", body: p("M3 15l4-11 3 7 3-5 4 9z") },
  { id: "dropstone", name: "Dropstone", group: "Other", body: p("M1 15h6M13 15h6M1 18h18") + `<ellipse cx="10" cy="13" rx="3" ry="2.4" ${S}/>` },
  { id: "sample", name: "Sample taken", group: "Other", body: p("M4 4h12v12H4z") + `<circle cx="10" cy="10" r="2" fill="currentColor"/>` },
  { id: "photo", name: "Photograph", group: "Other", body: p("M3 6h3l1.5-2h5L14 6h3v10H3z") + `<circle cx="10" cy="11" r="3" ${S}/>` },
];

export const SYMBOL_BY_ID = new Map(SYMBOLS.map((s) => [s.id, s]));
export const SYMBOL_GROUPS: SymbolGroup[] = ["Structures", "Fossils", "Trace fossils", "Other"];

/** <symbol> definitions for use with <use href="#sym-id">. */
export function symbolDefs(ids: Iterable<string>, prefix = "sym-"): string {
  let out = "";
  for (const id of new Set(ids)) {
    const s = SYMBOL_BY_ID.get(id);
    if (s) out += `<symbol id="${prefix}${s.id}" viewBox="0 0 20 20">${s.body}</symbol>`;
  }
  return out;
}

/** A stand-alone icon for pickers and chips. */
export function symbolIcon(id: string, size = 24): string {
  const s = SYMBOL_BY_ID.get(id);
  if (!s) return "";
  return `<svg class="symicon" width="${size}" height="${size}" viewBox="0 0 20 20" aria-hidden="true">${s.body}</svg>`;
}
