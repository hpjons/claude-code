// GoLog's own lithology patterns, drawn for this app in the usual geological conventions
// (dots for sand, bricks for limestone, circles for conglomerate …). Each is a square tile.

export interface PatternDef {
  id: string;
  name: string;
  group: "Clastic" | "Carbonate" | "Volcanic" | "Other";
  size: number; // tile size in px at 100 %
  colour: string; // background in colour logs: pale, so the black pattern stays readable
  fill?: string; // background in black-and-white logs; white if not given
  body: string; // SVG drawn inside the tile
}

const INK = "#1d2125";
const line = (d: string, w = 0.8) => `<path d="${d}" fill="none" stroke="${INK}" stroke-width="${w}" stroke-linecap="round"/>`;
const dots = (pts: [number, number][], r = 0.75) =>
  pts.map(([x, y]) => `<circle cx="${x}" cy="${y}" r="${r}" fill="${INK}"/>`).join("");

export const PATTERNS: PatternDef[] = [
  {
    id: "sandstone", name: "Sandstone", group: "Clastic", size: 18, colour: "#f6e08a",
    body: dots([[2, 3], [9, 1.5], [15, 4], [5, 8.5], [12, 9], [1.5, 14], [8, 13.5], [14.5, 15.5], [11, 5.5], [5, 17]]),
  },
  {
    id: "siltstone", name: "Siltstone", group: "Clastic", size: 18, colour: "#dcdba4",
    body: line("M1 4.5h5M10 4.5h5M5.5 13.5h5M14 13.5h4M0 13.5h1") + dots([[8, 4.5], [17, 4.5], [3, 13.5], [12.5, 13.5]], 0.6),
  },
  {
    id: "mudstone", name: "Mudstone", group: "Clastic", size: 18, colour: "#c8ccc4",
    body: line("M1 3h5M10 3h5M5.5 9h5M14 9h4M0 9h1.5M1 15h5M10 15h5"),
  },
  {
    id: "claystone", name: "Claystone", group: "Clastic", size: 24, colour: "#b7c6b2",
    body: line("M1 4h10M14 4h9M7 12h10M20 12h4M0 12h4M1 20h10M14 20h9"),
  },
  {
    id: "shale", name: "Shale", group: "Clastic", size: 16, colour: "#a9b1aa",
    body: line("M0 2h16M0 6h16M0 10h16M0 14h16", 0.7),
  },
  {
    id: "conglomerate", name: "Conglomerate", group: "Clastic", size: 28, colour: "#f2bf7c",
    body:
      `<g fill="none" stroke="${INK}" stroke-width="0.8">` +
      `<ellipse cx="7" cy="7" rx="4.5" ry="3.5"/><ellipse cx="20" cy="5" rx="3" ry="2.5"/>` +
      `<ellipse cx="15" cy="16" rx="5" ry="3.8"/><ellipse cx="4" cy="22" rx="3.2" ry="2.6"/>` +
      `<ellipse cx="24" cy="23" rx="3.5" ry="2.8"/></g>` +
      dots([[12, 3], [25, 13], [9, 26], [2, 14]], 0.6),
  },
  {
    id: "matrix-conglomerate", name: "Matrix-supported conglomerate", group: "Clastic", size: 32, colour: "#edc89a",
    body:
      line("M2 4h4M14 3h4M24 8h4M8 15h4M20 19h4M2 25h4M14 29h4M26 28h4", 0.6) +
      `<g fill="none" stroke="${INK}" stroke-width="0.8"><ellipse cx="10" cy="7" rx="3.6" ry="2.6"/>` +
      `<ellipse cx="24" cy="14" rx="3" ry="2.2"/><ellipse cx="9" cy="21" rx="3.2" ry="2.4"/><ellipse cx="25" cy="25" rx="2.4" ry="1.9"/></g>`,
  },
  {
    id: "breccia", name: "Breccia", group: "Clastic", size: 28, colour: "#e6ad86",
    body:
      `<g fill="none" stroke="${INK}" stroke-width="0.8" stroke-linejoin="round">` +
      `<path d="M3 10l5-7 4 6z"/><path d="M16 4l8 2-3 6z"/><path d="M8 20l7-6 3 8z"/>` +
      `<path d="M21 17l6 3-5 6z"/><path d="M2 26l3-6 5 5z"/></g>` +
      dots([[13, 10], [24, 1.5], [17, 26]], 0.6),
  },
  {
    id: "diamictite", name: "Diamictite", group: "Clastic", size: 32, colour: "#d9c6a2",
    body:
      line("M1 5h5M12 5h5M24 5h5M6 13h5M20 13h5M1 21h5M14 21h5M26 21h5M8 29h5M22 29h5", 0.6) +
      `<g fill="none" stroke="${INK}" stroke-width="0.8"><ellipse cx="18" cy="9" rx="3.5" ry="2.6"/>` +
      `<path d="M4 15l5 1-2 5z"/><path d="M25 24l5 2-3 4z"/></g>`,
  },
  {
    id: "limestone", name: "Limestone", group: "Carbonate", size: 24, colour: "#a9d0ea",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.8),
  },
  {
    id: "dolomite", name: "Dolomite", group: "Carbonate", size: 24, colour: "#c6b9e3",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M4 0.5l4 8M16 8.5l4 8M4 16.5l4 8", 0.8),
  },
  {
    id: "chalk", name: "Chalk", group: "Carbonate", size: 24, colour: "#e3f1f6",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.5) + dots([[13, 4.5], [1, 12.5], [12, 20.5]], 0.6),
  },
  {
    id: "lime-mudstone", name: "Lime mudstone", group: "Carbonate", size: 24, colour: "#b7d8ed",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.8) + line("M9 4.5h5M1 12.5h4M10 12.5h5M13 20.5h5", 0.5),
  },
  {
    id: "wackestone", name: "Wackestone", group: "Carbonate", size: 24, colour: "#a6cfe9",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.8) + line("M10 4.5h3M1 12.5h3M14 20.5h3", 0.5) + `<g fill="none" stroke="${INK}" stroke-width="0.6"><circle cx="17" cy="4.5" r="1.3"/><circle cx="10" cy="12.5" r="1.3"/><circle cx="4" cy="20.5" r="1.3"/></g>`,
  },
  {
    id: "packstone", name: "Packstone", group: "Carbonate", size: 24, colour: "#9cc8e5",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.8) + `<g fill="none" stroke="${INK}" stroke-width="0.6"><circle cx="11" cy="4.5" r="1.3"/><circle cx="16" cy="4.5" r="1.3"/><circle cx="2.5" cy="12.5" r="1.3"/><circle cx="9" cy="12.5" r="1.3"/><circle cx="14" cy="12.5" r="1.3"/><circle cx="10" cy="20.5" r="1.3"/><circle cx="15" cy="20.5" r="1.3"/><circle cx="21" cy="20.5" r="1.3"/></g>`,
  },
  {
    id: "grainstone", name: "Grainstone", group: "Carbonate", size: 24, colour: "#90c0e1",
    body: line("M0 0.5h24M0 8.5h24M0 16.5h24M6 0.5v8M18 8.5v8M6 16.5v8", 0.8) + `<g fill="none" stroke="${INK}" stroke-width="0.6">` +
      [[9.5, 4.5], [12.5, 4.5], [15.5, 4.5], [2, 4.5], [21, 4.5], [3, 12.5], [9, 12.5], [12, 12.5], [21, 12.5], [2, 20.5], [9.5, 20.5], [12.5, 20.5], [15.5, 20.5], [21, 20.5]]
        .map(([x, y]) => `<circle cx="${x}" cy="${y}" r="1.2"/>`).join("") + `</g>`,
  },
  {
    id: "marl", name: "Marl", group: "Carbonate", size: 24, colour: "#b4d3cb",
    body: line("M0 0.5h24M0 12.5h24M8 0.5v12M20 12.5v12") + line("M2 6.5h4M14 6.5h5M8 18.5h5M0 18.5h2M22 18.5h2", 0.6),
  },
  {
    id: "tuff", name: "Tuff", group: "Volcanic", size: 20, colour: "#f3c6d2",
    body: line("M2 3l2 3 2-3M12 6l2 3 2-3M6 13l2 3 2-3M15 15l2 3 2-3", 0.7) + dots([[17, 2], [3, 18], [11, 11]], 0.55),
  },
  {
    id: "lava", name: "Lava (basalt)", group: "Volcanic", size: 24, colour: "#c0a2ab",
    body: line("M2 3l3.5 5 3.5-5M14 3l3.5 5 3.5-5M8 15l3.5 5 3.5-5M20 15l3.5 5 3.5-5M-4 15l3.5 5 3.5-5", 0.9),
  },
  {
    id: "hyaloclastite", name: "Hyaloclastite", group: "Volcanic", size: 28, colour: "#d4bb7e",
    body:
      line("M3 4l2 3 2-3M17 9l2 3 2-3M8 20l2 3 2-3", 0.7) +
      `<g fill="none" stroke="${INK}" stroke-width="0.8" stroke-linejoin="round"><path d="M11 5l5 1-2 4z"/>` +
      `<path d="M21 18l5 2-4 4z"/><path d="M2 13l4-1 0 5z"/></g>` +
      dots([[26, 4], [15, 15], [3, 26]], 0.6),
  },
  {
    id: "volcaniclastic", name: "Volcaniclastic sediment", group: "Volcanic", size: 20, colour: "#e9c3c6",
    body: line("M2 4l2 3 2-3M12 13l2 3 2-3", 0.7) + dots([[12, 3], [16, 6], [4, 13], [8, 17], [17, 18], [9, 9]], 0.6),
  },
  {
    id: "coarse-ash", name: "Coarse ash / lapilli tuff", group: "Volcanic", size: 24, colour: "#efb8c5",
    body: line("M2 3l2 3 2-3M14 9l2 3 2-3M5 17l2 3 2-3", 0.7) + `<g fill="none" stroke="${INK}" stroke-width="0.7"><circle cx="14" cy="3.5" r="1.8"/><circle cx="5" cy="11" r="2"/><circle cx="19" cy="18" r="1.8"/></g>`,
  },
  {
    id: "coal", name: "Coal", group: "Other", size: 8, colour: "#1d2125", fill: INK, body: "",
  },
  {
    id: "chert", name: "Chert", group: "Other", size: 24, colour: "#efd3c2",
    body: `<g fill="none" stroke="${INK}" stroke-width="0.8" stroke-linejoin="round"><path d="M2 9l4-7 4 7z"/><path d="M14 21l4-7 4 7z"/></g>`,
  },
  {
    id: "evaporite", name: "Evaporite (gypsum, halite)", group: "Other", size: 20, colour: "#e6daf2",
    body: line("M3 3l4 4M7 3l-4 4M13 13l4 4M17 13l-4 4", 0.8),
  },
  {
    id: "halite", name: "Halite", group: "Other", size: 20, colour: "#f1e8f6",
    body: `<g fill="none" stroke="${INK}" stroke-width="0.7"><rect x="2" y="2" width="6" height="6"/><rect x="12" y="12" width="6" height="6"/></g>`,
  },
  {
    id: "gypsum", name: "Gypsum / anhydrite", group: "Other", size: 20, colour: "#e8dcf1",
    body: line("M2 8l3-5 3 5M12 18l3-5 3 5M12 3h6M2 13h6", 0.7),
  },
  {
    id: "soil", name: "Soil / palaeosol", group: "Other", size: 24, colour: "#caa37e",
    body: line("M2 4c3-2 5 2 8 0M14 12c3-2 5 2 8 0M4 20c3-2 5 2 8 0", 0.8) + line("M18 2v5M7 10v5M20 18v5", 0.6),
  },
];

export const PATTERN_BY_ID = new Map(PATTERNS.map((p) => [p.id, p]));

/** How a log's lithologies are filled: conventional colours, or white as in a black-and-white log. */
export type FillMode = "colour" | "white";

/** The background a lithology is drawn on: its own colour if set, else the pattern's. */
export function lithColour(pattern: string, mode: FillMode, own?: string): string {
  const p = PATTERN_BY_ID.get(pattern);
  if (mode === "white") return p?.fill ?? "#fff";
  return own ?? p?.colour ?? "#fff";
}

/**
 * SVG <pattern> elements for the given pattern ids, scaled by `scale` (1 = 100 %). Their
 * background is transparent: a rock is drawn as its colour with the pattern on top, which
 * leaves no seams between tiles in PDF viewers.
 */
export function patternDefs(ids: Iterable<string>, scale = 1, prefix = "pat-"): string {
  let out = "";
  for (const id of new Set(ids)) {
    const p = PATTERN_BY_ID.get(id);
    if (!p) continue;
    const s = p.size;
    out +=
      `<pattern id="${prefix}${p.id}" patternUnits="userSpaceOnUse" width="${s * scale}" height="${s * scale}"` +
      ` viewBox="0 0 ${s} ${s}">${p.body}</pattern>`;
  }
  return out;
}

/** A shape filled with a rock: its colour, then its pattern. `shape` is e.g. `rect x=".." …`. */
export function rockFill(shape: string, pattern: string, colour: string, extra = ""): string {
  const [tag, ...attrs] = shape.split(" ");
  const a = attrs.join(" ");
  return `<${tag} ${a} fill="${colour}"/>` + (PATTERN_BY_ID.has(pattern) ? `<${tag} ${a} fill="url(#pat-${pattern})"${extra ? " " + extra : ""}/>` : "");
}

/** A small stand-alone SVG swatch of a pattern, for pickers. */
export function patternSwatch(id: string, size = 36, colour = lithColour(id, "colour")): string {
  const p = PATTERN_BY_ID.get(id);
  if (!p) return "";
  const box = `x="0.5" y="0.5" width="${size - 1}" height="${size - 1}"`;
  return (
    `<svg class="swatch" width="${size}" height="${size}" viewBox="0 0 ${size} ${size}" aria-hidden="true">` +
    `<defs>${patternDefs([id], 1, `sw-${size}-`)}</defs>` +
    `<rect ${box} fill="${colour}"/><rect ${box} fill="url(#sw-${size}-${id})" stroke="${INK}"/></svg>`
  );
}
