// Turns a log into files to save: PDF, SVG, PNG, CSV and GoLog's own file, and prints it.
import { grainClassOf } from "../grain";
import { PATTERN_BY_ID } from "../render/patterns";
import { SYMBOL_BY_ID } from "../render/symbols";
import type { Bed, Log } from "../types";
import { layoutPages, svgSize, wholeLog } from "./pages";

/** A safe file name from the log's name. */
export function fileName(log: Log, ext: string): string {
  const base = log.name.replace(/[\\/:*?"<>|\u0000-\u001f]+/g, " ").replace(/\s+/g, " ").trim().slice(0, 100) || "log";
  return `${base}.${ext}`;
}

/** Hands a file to the browser to save. Also announced as an event, which the browser test listens for. */
export function download(name: string, blob: Blob): void {
  window.dispatchEvent(new CustomEvent("golog-download", { detail: { name, blob } }));
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = name;
  document.body.append(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}

function parse(svg: string): SVGSVGElement {
  return new DOMParser().parseFromString(svg, "image/svg+xml").documentElement as unknown as SVGSVGElement;
}

export async function toPDF(log: Log, withKey: boolean): Promise<Blob> {
  const [{ jsPDF }, { svg2pdf }] = await Promise.all([import("jspdf"), import("svg2pdf.js")]);
  const { svgs, paper } = layoutPages(log, withKey);
  const [w, h] = paper;
  const doc = new jsPDF({ unit: "mm", format: [w, h], orientation: "portrait", compress: true });
  doc.setProperties({ title: log.name, creator: "GoLog" });
  // svg2pdf reads styles from the live document, so each page is drawn from a hidden copy.
  const holder = document.createElement("div");
  holder.style.cssText = "position:fixed;left:-10000px;top:0;visibility:hidden";
  document.body.append(holder);
  try {
    for (const [i, svg] of svgs.entries()) {
      if (i > 0) doc.addPage([w, h], "portrait");
      // PDF's built-in Helvetica has normal and bold only.
      const el = parse(svg.replace(/font-weight:600/g, "font-weight:bold"));
      holder.replaceChildren(el);
      await svg2pdf(el, doc, { x: 0, y: 0, width: w, height: h });
    }
  } finally {
    holder.remove();
  }
  return doc.output("blob");
}

export function toSVG(log: Log, withKey: boolean): Blob {
  return new Blob([`<?xml version="1.0" encoding="UTF-8"?>\n`, wholeLog(log, withKey)], { type: "image/svg+xml" });
}

/** The whole log as a PNG, at 300 dpi or less if the log is too long for a picture that size. */
export async function toPNG(log: Log, withKey: boolean): Promise<Blob> {
  const svg = wholeLog(log, withKey);
  const [w, h] = svgSize(svg);
  let k = 300 / 96;
  k = Math.min(k, 16000 / w, 16000 / h, Math.sqrt(120e6 / (w * h)));
  const img = new Image();
  img.src = "data:image/svg+xml;charset=utf-8," + encodeURIComponent(svg);
  await img.decode();
  const canvas = document.createElement("canvas");
  canvas.width = Math.round(w * k);
  canvas.height = Math.round(h * k);
  const g = canvas.getContext("2d")!;
  g.fillStyle = "#fff";
  g.fillRect(0, 0, canvas.width, canvas.height);
  g.drawImage(img, 0, 0, canvas.width, canvas.height);
  return new Promise((resolve, reject) => canvas.toBlob((b) => (b ? resolve(b) : reject(new Error("the picture could not be made"))), "image/png"));
}

// SedLog's names for GoLog's patterns and symbols, so the CSV reads back into SedLog and GoLog.
const SEDLOG_LITHOLOGY: Record<string, string> = {
  mudstone: "Mudstone", claystone: "Claystone", shale: "Shale", siltstone: "Siltstone", sandstone: "Sandstone",
  conglomerate: "Conglomerate", coal: "Coal", limestone: "Limestone", chert: "Chert", volcaniclastic: "Volcaniclastic",
  "lime-mudstone": "Lime mudstone", wackestone: "Wackestone", packstone: "Packstone", grainstone: "Grainstone",
  halite: "Halite", gypsum: "Gypsum/Anhydrite", dolomite: "Dolomite", breccia: "Breccia",
  "matrix-conglomerate": "Matrix-supported conglomerate", lava: "Lava", tuff: "Fine ash", "coarse-ash": "Coarse ash",
};
const SEDLOG_SYMBOL: Record<string, string> = {
  "shell-debris": "Shells", bivalve: "Bivalves", gastropod: "Gastropods", ammonite: "Cephalopods", brachiopod: "Brachiopods",
  echinoid: "Echinoids", crinoid: "Crinoids", coral: "Solitary corals", foraminifera: "Foraminifera", algae: "Algae",
  bryozoa: "Bryzoa", stromatolite: "Stromatolites", bone: "Vertebrates", plant: "Plant material", roots: "Roots", wood: "Logs",
  "tree-stump": "Tree stumps", ostracod: "Ostracods", radiolaria: "Radiolaria", sponge: "Sponges",
  nodules: "Nodules and concretions", intraclasts: "Intraclasts", "rip-up-clasts": "Mudclasts", "flute-casts": "Flute marks",
  "groove-casts": "Groove marks", scours: "Scours", "current-ripples": "Current ripple cross-lamination",
  "wave-ripples": "Wave ripple cross-lamination", "planar-cross-bedding": "Planar cross bedding", "trough-cross-bedding": "Trough cross bedding",
  "planar-lamination": "Horizontal planar lamination", hummocky: "Hummocky cross stratification", swaley: "Swaley cross stratification",
  mudcracks: "Mudcracks", "synaeresis-cracks": "Synaeresis cracks", convolute: "Convolute lamination", "load-casts": "Load casts",
  "water-escape": "Water structures", herringbone: "Herring-bone cross bedding", bioturbation: "Moderate bioturbation",
  footprints: "Tracks", trails: "Trails", "vertical-burrows": "Vertical burrows", "horizontal-burrows": "Horizontal burrrows",
};
const SEDLOG_GRAIN: Record<string, string> = { gran: "granule", pebb: "pebble", cobb: "cobble", boul: "boulder" };
// The values SedLog itself writes, including its in-between classes, so they keep their names.
const SEDLOG_GRAIN_VALUES: [string, number][] = [
  ["clay", 10], ["clay/silt", 8], ["silt", 6], ["vf", 3.5], ["vf/f", 3], ["f", 2.5], ["f/m", 2], ["m", 1.5], ["m/c", 1],
  ["c", 0.5], ["c/vc", 0], ["vc", -0.5], ["vc/granule", -1], ["granule", -1.5], ["pebble", -3],
];
const SEDLOG_CONTACT: Record<Bed["contact"], string> = { unknown: "<none>", sharp: "Sharp", erosional: "Erosion", gradational: "Gradational" };

const lithName = (id: string) => SEDLOG_LITHOLOGY[id] ?? PATTERN_BY_ID.get(id)?.name ?? id;
const symName = (id: string) => SEDLOG_SYMBOL[id] ?? SYMBOL_BY_ID.get(id)?.name ?? id;
const grainName = (phi: number | undefined) => {
  if (phi === undefined) return "<none>";
  const exact = SEDLOG_GRAIN_VALUES.find(([, v]) => v === phi);
  if (exact) return exact[0];
  const key = grainClassOf(phi).key;
  return SEDLOG_GRAIN[key] ?? key;
};

export function cell(v: string | number): string {
  const s = String(v);
  return /[",\n\r]/.test(s) || s === "" ? `"${s.replace(/"/g, '""')}"` : s;
}

/**
 * The beds as a table in SedLog's CSV columns (plus formation and age), from the base up.
 * For spreadsheets the file starts with a byte-order mark, which Excel needs to read Icelandic
 * letters right; SedLog refuses files that have one, so its copy goes without.
 */
export function toCSV(log: Log, forSedLog = false): Blob {
  const head = [
    "THICKNESS (CM)", "BASE BOUNDARY", "LITHOLOGY", "LITHOLOGY %", "LITHOLOGY2", "LITHOLOGY2 %", "LITHOLOGY3", "LITHOLOGY3 %",
    "GRAIN SIZE BASE", "PHI VALUES BASE", "GRAIN SIZE TOP", "PHI VALUES TOP", "SYMBOLS IN BED", "SYMBOLS/STRUCTURES", "NOTES COLUMN",
    "BIOTURBATION TYPE", "INTENSITY", "PALAEOCURRENT VALUES", "FACIES", "OTHER1 TEXT", "OTHER1 SYMBOL", "OTHER2 TEXT", "OTHER2 SYMBOL",
    "OTHER3 TEXT", "OTHER3 SYMBOL", "FORMATION", "AGE",
  ];
  // SedLog's CSV reader drops double quotes inside a cell, so its copy gets single ones.
  const text = (s: string | undefined) => (forSedLog ? (s ?? "").replace(/"/g, "'") : (s ?? ""));
  const rows = log.beds.map((b) => {
    const lith = [0, 1, 2].flatMap((i) => {
      const l = b.lithology[i];
      return l ? [lithName(l.pattern), +l.percent.toFixed(1)] : ["<none>", 0];
    });
    const top = b.grain.top ?? b.grain.base;
    return [
      +(b.thickness * 100).toFixed(4), SEDLOG_CONTACT[b.contact] ?? "<none>", ...lith,
      grainName(b.grain.base), b.grain.base ?? 0, grainName(top), top ?? 0,
      b.inBed ? symName(b.inBed) : "<none>", (b.symbols ?? []).map(symName).join(", "), text(b.notes),
      // SedLog's facies are whole numbers and it refuses anything else, so its copy gets 0 for names like "St".
      "<none>", b.bioturbation ?? 0, (b.palaeocurrents ?? []).join(", "), forSedLog ? (/^\d+$/.test(b.facies?.trim() ?? "") ? b.facies!.trim() : 0) : (b.facies ?? ""),
      "", "", "", "", "", "", text(b.formation), text(b.age),
    ].map(cell);
  });
  return new Blob([(forSedLog ? "" : "\uFEFF") + [head.map(cell), ...rows].map((r) => r.join(",")).join("\r\n") + "\r\n"], { type: "text/csv;charset=utf-8" });
}

/** GoLog's own file: everything about the log, to back up or give to someone with GoLog. */
export function toGoLog(log: Log): Blob {
  const { id: _id, projectId: _p, revision: _r, created: _c, updated: _u, name, ...content } = log;
  const file = { golog: 1, kind: "log", name, saved: new Date().toISOString(), content };
  return new Blob([JSON.stringify(file, null, 1)], { type: "application/json" });
}

/**
 * Gives a drawing's patterns and symbols names of its own. Pages printed from the app's
 * document would otherwise use the hidden on-screen drawing's patterns, which do not print.
 */
export function ownIds(svg: string, prefix: string): string {
  return svg.replace(/\bid="([^"]+)"/g, `id="${prefix}$1"`).replace(/url\(#([^)]+)\)/g, `url(#${prefix}$1)`).replace(/href="#([^"]+)"/g, `href="#${prefix}$1"`);
}

/** Prints the log on its paper, one page per sheet, through the browser's print window. */
export function printLog(log: Log, withKey: boolean): void {
  const { svgs, paper } = layoutPages(log, withKey);
  document.getElementById("print-root")?.remove();
  const root = document.createElement("div");
  root.id = "print-root";
  const style = document.createElement("style");
  style.textContent = `@page { size: ${paper[0]}mm ${paper[1]}mm; margin: 0 }`;
  root.append(style);
  for (const [i, svg] of svgs.entries()) {
    const sheet = document.createElement("div");
    sheet.className = "print-sheet";
    sheet.innerHTML = ownIds(svg, `print${i}-`);
    root.append(sheet);
  }
  document.body.append(root);
  const done = () => {
    root.remove();
    window.removeEventListener("afterprint", done);
  };
  window.addEventListener("afterprint", done);
  window.print();
}
