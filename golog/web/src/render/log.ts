// Draws a log as an SVG string at its print scale. The same drawing is used on screen
// (scaled by the zoom) and, later, for PDF/SVG export.
import { GRAIN_CLASSES, grainIndex } from "../grain";
import type { Bed, Column, Lithology, Log, Paper } from "../types";
import { PATTERN_BY_ID, lithColour, patternDefs, rockFill } from "./patterns";
import { SYMBOL_BY_ID, symbolDefs } from "./symbols";

export interface DrawOptions {
  zoom: number; // 1 = print size on a 96 dpi screen
  selectedId?: string;
  pageBreaks?: boolean; // show where printed pages will end
  handles?: boolean; // the editor's invisible click areas and drag handles (never in exports)
  page?: number; // draw only this printed page (0 = top of the section), with the headers repeated
  key?: boolean; // add the key below the log (or below this page)
}

const INK = "#1d2125";
const MUTED = "#6b7177";
export const TITLE = 26; // log name and scale above the columns
export const HEADER = 96;
const FOOT = 10;
const SYMBOL = 16; // symbol size in the symbols column
const GRAIN_EDGE = 10; // px kept free right of the coarsest grain class

const MM = 96 / 25.4; // CSS px per millimetre
export const PAPER_MM: Record<Paper, [number, number]> = { A4: [210, 297], A3: [297, 420], Letter: [215.9, 279.4] };
export const MARGIN_MM = 12;

/** Px per metre of section at print scale 1:scale. */
export const pxPerMetre = (scale: number) => (1000 * MM) / scale;

export const esc = (s: string) =>
  s.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]!);

const r1 = (n: number) => Math.round(n * 10) / 10;

/** Profile width (px from the left of a grain column `w` wide) for a phi value. */
export function grainWidth(phi: number | undefined, w = 210): number {
  const min = Math.max(10, w * 0.085);
  // The coarsest class stops short of the column's right edge, so its tick label stays inside.
  const max = w - GRAIN_EDGE;
  const i = phi === undefined ? 0 : grainIndex(phi);
  return min + (i / (GRAIN_CLASSES.length - 1)) * (max - min);
}

/** The grain class (as phi) nearest to `px` from the left of a grain column `w` wide: the inverse of grainWidth. */
export function phiAt(px: number, w = 210): number {
  const min = Math.max(10, w * 0.085);
  const n = GRAIN_CLASSES.length - 1;
  const i = Math.round(((px - min) / (w - GRAIN_EDGE - min)) * n);
  return GRAIN_CLASSES[Math.max(0, Math.min(n, i))]!.phi;
}

/** Left edge and width of a visible column, in drawing coordinates. */
export function columnBox(log: Log, kind: Column["kind"]): { x: number; width: number } | undefined {
  let x = 0;
  for (const col of visibleColumns(log)) {
    if (col.kind === kind) return { x, width: col.width };
    x += col.width;
  }
  return undefined;
}

/** Bed geometry: y of base and top in drawing coordinates. */
export interface BedBox {
  bed: Bed;
  index: number;
  yBase: number;
  yTop: number;
}

export function layoutBeds(log: Log, pxPerM: number, top = TITLE + HEADER): { boxes: BedBox[]; bodyHeight: number } {
  const total = log.beds.reduce((s, b) => s + b.thickness, 0);
  const bodyHeight = total * pxPerM;
  let y = top + bodyHeight;
  const boxes = log.beds.map((bed, index) => {
    const yBase = y;
    y -= bed.thickness * pxPerM;
    return { bed, index, yBase, yTop: y };
  });
  return { boxes, bodyHeight };
}

/** Tick spacing in metres so that labels are at least ~36 px apart. */
export function tickStep(pxPerM: number): number {
  for (const s of [0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 20, 50, 100, 200, 500, 1000]) if (s * pxPerM >= 36) return s;
  return 1000;
}

/** How the log fits on paper. */
export interface PageInfo {
  pages: number;
  bodyPerPage: number; // px of section per page
  width: number; // px of all visible columns
  pageWidth: number; // px available across the page
  fitsWidth: boolean;
}

export function pageInfo(log: Log): PageInfo {
  const [wmm, hmm] = PAPER_MM[log.paper] ?? PAPER_MM.A4;
  const pageWidth = (wmm - 2 * MARGIN_MM) * MM;
  const bodyPerPage = (hmm - 2 * MARGIN_MM) * MM - TITLE - HEADER - FOOT;
  const body = log.beds.reduce((s, b) => s + b.thickness, 0) * pxPerMetre(log.scale);
  const width = visibleColumns(log).reduce((s, c) => s + c.width, 0);
  return { pages: Math.max(1, Math.ceil(body / bodyPerPage - 1e-9)), bodyPerPage, width, pageWidth, fitsWidth: width <= pageWidth + 0.5 };
}

export const visibleColumns = (log: Log) => log.columns.filter((c) => c.visible);

/** A shape filled with a lithology as this log draws it (its colour, or white, under the pattern). */
const rock = (log: Log, l: Lithology, shape: string) => rockFill(shape, l.pattern, lithColour(l.pattern, log.fill, l.colour));

// ---- per-column drawing ------------------------------------------------------------

interface Ctx {
  log: Log;
  boxes: BedBox[];
  pxPerM: number;
  yHead: number; // top of the column headers
  yBody: number; // top of the section (= bottom of headers)
  yZero: number; // base of the section
}

type Drawer = { head: (c: Ctx, col: Column, x: number) => string; body: (c: Ctx, col: Column, x: number) => string };

function title(c: Ctx, col: Column, x: number, extra = ""): string {
  const t = esc(col.title);
  if (col.width < 70) {
    // Narrow columns get vertical titles, read from the bottom up.
    return `<text transform="translate(${r1(x + col.width / 2 + 4)} ${c.yBody - 8}) rotate(-90)" class="colhead">${t}${extra ? " " + extra : ""}</text>`;
  }
  const lines = wrap(col.title, Math.max(6, Math.floor((col.width - 8) / 6.4))).slice(0, 3);
  return (
    lines
      .map((l, i) => `<text x="${r1(x + col.width / 2)}" y="${c.yBody - 10 - (lines.length - 1 - i) * 13}" class="colhead" text-anchor="middle">${esc(l)}</text>`)
      .join("") +
    (extra ? `<text x="${r1(x + col.width / 2)}" y="${c.yBody - 10 - lines.length * 13}" class="tick" text-anchor="middle">${extra}</text>` : "")
  );
}

function wrap(text: string, maxChars: number): string[] {
  const lines: string[] = [];
  for (const para of text.split("\n")) {
    let cur = "";
    for (const word of para.split(/\s+/).filter(Boolean)) {
      if (cur && (cur + " " + word).length > maxChars) {
        lines.push(cur);
        cur = word;
      } else cur = cur ? cur + " " + word : word;
    }
    lines.push(cur);
  }
  return lines;
}

/** Text in a bed, wrapped to the column and cut off with … if the bed is too short. */
function bedText(box: BedBox, x: number, w: number, text: string, cls = "note", center = false): string {
  const lines = wrap(text.trim(), Math.max(3, Math.floor((w - 8) / 6.1)));
  const room = Math.max(1, Math.floor((box.yBase - box.yTop - 3) / 12.5));
  const shown = lines.slice(0, room);
  if (lines.length > room) shown[room - 1] = shown[room - 1]!.replace(/.?.?$/, "…");
  const tx = center ? x + w / 2 : x + 5;
  const y0 =
    center && shown.length * 12.5 < box.yBase - box.yTop ? (box.yTop + box.yBase) / 2 - ((shown.length - 1) * 12.5) / 2 + 4 : box.yTop + 12;
  return (
    `<text x="${r1(tx)}" y="${r1(y0)}" class="${cls}"${center ? ' text-anchor="middle"' : ""}>` +
    shown.map((l, i) => `<tspan x="${r1(tx)}" dy="${i === 0 ? 0 : 12.5}">${esc(l)}</tspan>`).join("") +
    `</text>`
  );
}

function wavy(x0: number, x1: number, y: number): string {
  let d = `M${r1(x0)} ${r1(y)}`;
  for (let x = x0; x < x1; x += 8) {
    const xe = Math.min(x + 8, x1);
    d += ` Q${r1(x + 2)} ${r1(y - 2.6)} ${r1((x + xe) / 2)} ${r1(y)} T${r1(xe)} ${r1(y)}`;
  }
  return d;
}

function contact(bed: Bed, x0: number, x1: number, y: number): string {
  const yy = r1(y);
  switch (bed.contact) {
    case "erosional":
      return `<path d="${wavy(x0, x1, y)}" fill="none" stroke="${INK}" stroke-width="1.3"/>`;
    case "gradational":
      return `<path d="M${r1(x0)} ${yy}H${r1(x1)}" stroke="${INK}" stroke-width="1" stroke-dasharray="4 3"/>`;
    case "sharp":
      return `<path d="M${r1(x0)} ${yy}H${r1(x1)}" stroke="${INK}" stroke-width="1.3"/>`;
    default:
      return `<path d="M${r1(x0)} ${yy}H${r1(x1)}" stroke="${INK}" stroke-width="0.6"/>`;
  }
}

/** Spans of beds that share a formation or age name: a new one starts where a bed names it. */
export function units(boxes: BedBox[], field: "formation" | "age"): { name: string; yBase: number; yTop: number }[] {
  const out: { name: string; yBase: number; yTop: number }[] = [];
  for (const box of boxes) {
    const name = box.bed[field]?.trim();
    if (name) out.push({ name, yBase: box.yBase, yTop: box.yTop });
    else if (out.length) out[out.length - 1]!.yTop = box.yTop;
  }
  return out;
}

function unitColumn(field: "formation" | "age"): Drawer {
  return {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) =>
      units(c.boxes, field)
        .map((u) => {
          const h = u.yBase - u.yTop;
          const maxChars = Math.max(1, Math.floor((h - 6) / 6.6));
          const name = u.name.length > maxChars ? u.name.slice(0, Math.max(1, maxChars - 1)) + "…" : u.name;
          return (
            `<path d="M${x} ${r1(u.yBase)}H${x + col.width}" stroke="${INK}" stroke-width="1"/>` +
            `<text transform="translate(${r1(x + col.width / 2 + 4)} ${r1((u.yBase + u.yTop) / 2)}) rotate(-90)" class="unit" text-anchor="middle">${esc(name)}</text>`
          );
        })
        .join(""),
  };
}

function symbolsInBed(ids: string[], x: number, w: number, yTop: number, yBase: number): string {
  const h = yBase - yTop;
  const size = Math.max(8, Math.min(SYMBOL, h - 2));
  const gap = 3;
  const perRow = Math.max(1, Math.floor((w - 4 + gap) / (size + gap)));
  const maxRows = Math.max(1, Math.floor((h + gap) / (size + gap)));
  const cells = Math.min(ids.length, perRow * maxRows);
  const hidden = ids.length - cells;
  const shown = hidden > 0 ? ids.slice(0, cells - 1) : ids; // the last cell says "+n"
  const rows = Math.ceil(cells / perRow);
  const y0 = (yTop + yBase) / 2 - (rows * (size + gap) - gap) / 2;
  let s = "";
  for (let k = 0; k < cells; k++) {
    const row = Math.floor(k / perRow);
    const inRow = Math.min(perRow, cells - row * perRow);
    const x0 = x + w / 2 - (inRow * (size + gap) - gap) / 2 + (k % perRow) * (size + gap);
    const y = y0 + row * (size + gap);
    if (k < shown.length) s += `<use href="#sym-${shown[k]}" x="${r1(x0)}" y="${r1(y)}" width="${r1(size)}" height="${r1(size)}" color="${INK}"/>`;
    else s += `<text x="${r1(x0 + size / 2)}" y="${r1(y + size / 2 + 3.5)}" class="tick" text-anchor="middle">+${ids.length - shown.length}</text>`;
  }
  return s;
}

const DRAWERS: Record<Column["kind"], Drawer> = {
  age: unitColumn("age"),
  formation: unitColumn("formation"),

  scale: {
    head: (c, col, x) => title(c, col, x, `(${c.log.units})`),
    body: (c, col, x) => {
      const total = c.log.beds.reduce((s, b) => s + b.thickness, 0);
      const step = tickStep(c.pxPerM);
      const xr = x + col.width;
      let s = "";
      for (let k = 0; k * step <= total + 1e-9; k++) {
        const m = k * step;
        const y = r1(c.yZero - m * c.pxPerM);
        const label = c.log.units === "cm" ? Math.round(m * 100) : +m.toFixed(2);
        s += `<path d="M${r1(xr - 7)} ${y}H${r1(xr)}" stroke="${INK}" stroke-width="0.9"/>`;
        // Labels sit beside their tick, but never on top of the column headers or below the frame.
        s += `<text x="${r1(xr - 10)}" y="${r1(Math.min(Math.max(y + 3.5, c.yBody + 9), c.yZero - 1.5))}" class="tick" text-anchor="end">${label}</text>`;
        if (m + step / 2 <= total) s += `<path d="M${r1(xr - 4)} ${r1(c.yZero - (m + step / 2) * c.pxPerM)}H${r1(xr)}" stroke="${INK}" stroke-width="0.6"/>`;
      }
      return s;
    },
  },

  lithology: {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) =>
      c.boxes
        .map(({ bed, yBase, yTop }) => {
          const h = r1(yBase - yTop);
          let s = "";
          if (!bed.lithology.length) s += `<rect x="${r1(x)}" y="${r1(yTop)}" width="${col.width}" height="${h}" fill="#fff"/>`;
          const total = bed.lithology.reduce((t, l) => t + l.percent, 0) || 1;
          let lx = x;
          for (const l of bed.lithology) {
            const w = (l.percent / total) * col.width;
            s += rock(c.log, l, `rect x="${r1(lx)}" y="${r1(yTop)}" width="${r1(w)}" height="${h}"`);
            lx += w;
          }
          return s + contact(bed, x, x + col.width, yBase);
        })
        .join(""),
  },

  grain: {
    head: (c, col, x) => {
      // Without its scale, the column's title sits at the foot of the header like the others.
      if (c.log.hideGrainClasses && c.log.hideGrainGroups) return c.log.hideGrainTitle ? "" : title(c, col, x);
      const xOf = (i: number) => x + grainWidth(GRAIN_CLASSES[i]!.phi, col.width);
      let s = "";
      const groups: Record<string, number[]> = {};
      GRAIN_CLASSES.forEach((g, i) => (groups[g.group] ??= []).push(i));
      for (const [g, idx] of c.log.hideGrainGroups ? [] : Object.entries(groups)) {
        const a = xOf(idx[0]!) - 4;
        const b = xOf(idx[idx.length - 1]!) + 4;
        s += `<text x="${r1((a + b) / 2)}" y="${c.yBody - 50}" class="colhead" text-anchor="middle">${g}</text>`;
        s += `<path d="M${r1(a)} ${c.yBody - 40}V${c.yBody - 44}H${r1(b)}V${c.yBody - 40}" fill="none" stroke="${INK}" stroke-width="0.6"/>`;
      }
      if (!c.log.hideGrainClasses) GRAIN_CLASSES.forEach((g, i) => {
        const xi = r1(xOf(i));
        s += `<path d="M${xi} ${c.yBody - 6}V${c.yBody}" stroke="${INK}" stroke-width="0.8"/>`;
        s += `<text transform="translate(${r1(xi + 3.5)} ${c.yBody - 9}) rotate(-90)" class="tick">${g.short}</text>`;
      });
      if (col.title.trim() && !c.log.hideGrainTitle) s += `<text x="${r1(x + col.width / 2)}" y="${c.yHead + 15}" class="colhead" text-anchor="middle">${esc(col.title)}</text>`;
      return s;
    },
    body: (c, col, x) => {
      let s = "";
      c.boxes.forEach((box, i) => {
        const { bed, yBase, yTop } = box;
        const wb = grainWidth(bed.grain.base, col.width);
        const wt = grainWidth(bed.grain.top ?? bed.grain.base, col.width);
        const shape = `M${r1(x)} ${r1(yBase)}H${r1(x + wb)}L${r1(x + wt)} ${r1(yTop)}H${r1(x)}Z`;
        const sym = bed.inBed ? SYMBOL_BY_ID.get(bed.inBed) : undefined;
        const first = bed.lithology[0];
        // An in-bed structure replaces the pattern but keeps the rock's colour.
        if (!first) s += `<path d="${shape}" fill="#fff"/>`;
        else if (sym) s += `<path d="${shape}" fill="${lithColour(first.pattern, c.log.fill, first.colour)}"/>`;
        else s += rock(c.log, first, `path d="${shape}"`);
        if (sym) s += `<path d="${shape}" fill="url(#inbed-${sym.id})"/>`;
        s += `<path d="M${r1(x + wb)} ${r1(yBase)}L${r1(x + wt)} ${r1(yTop)}" stroke="${INK}" stroke-width="1"/>`;
        const below = c.boxes[i - 1];
        const wBelow = below ? grainWidth(below.bed.grain.top ?? below.bed.grain.base, col.width) : 0;
        s += contact(bed, x, x + Math.max(wb, wBelow), yBase);
      });
      const top = c.boxes[c.boxes.length - 1];
      if (top) s += `<path d="M${r1(x)} ${r1(top.yTop)}H${r1(x + grainWidth(top.bed.grain.top ?? top.bed.grain.base, col.width))}" stroke="${INK}" stroke-width="1"/>`;
      return s;
    },
  },

  symbols: {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) =>
      c.boxes
        .map(({ bed, yBase, yTop }) => {
          const ids = (bed.symbols ?? []).filter((id) => SYMBOL_BY_ID.has(id));
          return ids.length ? symbolsInBed(ids, x, col.width, yTop, yBase) : "";
        })
        .join(""),
  },

  bioturbation: {
    head: (c, col, x) =>
      title(c, col, x) +
      `<text x="${r1(x + 4)}" y="${c.yBody - 3}" class="tiny">0</text><text x="${r1(x + col.width - 4)}" y="${c.yBody - 3}" class="tiny" text-anchor="end">6</text>`,
    body: (c, col, x) =>
      c.boxes
        .map(({ bed, yBase, yTop }) => {
          if (bed.bioturbation === undefined) return "";
          const w = (bed.bioturbation / 6) * (col.width - 6);
          const bar =
            w > 0
              ? `<rect x="${r1(x + 3)}" y="${r1(yTop + 0.5)}" width="${r1(w)}" height="${r1(Math.max(1, yBase - yTop - 1))}" fill="#c9cec9" stroke="${INK}" stroke-width="0.6"/>`
              : "";
          const label =
            yBase - yTop >= 11 ? `<text x="${r1(x + col.width - 4)}" y="${r1((yTop + yBase) / 2 + 3.5)}" class="tick" text-anchor="end">${bed.bioturbation}</text>` : "";
          return bar + label;
        })
        .join(""),
  },

  palaeocurrent: {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) =>
      c.boxes
        .map((box) => {
          const vals = box.bed.palaeocurrents ?? [];
          if (!vals.length) return "";
          if (c.log.palaeocurrent === "numbers") return bedText(box, x, col.width, vals.map((v) => `${+v.toFixed(1)}°`).join(", "));
          const len = Math.max(8, Math.min(16, box.yBase - box.yTop - 2));
          const per = Math.max(1, Math.floor((col.width - 4) / (len + 2)));
          const shown = vals.slice(0, per);
          const cy = (box.yTop + box.yBase) / 2;
          const h = len / 2;
          return (
            shown
              .map((deg, k) => {
                const cx = x + col.width / 2 - ((shown.length - 1) * (len + 2)) / 2 + k * (len + 2);
                return (
                  `<g transform="translate(${r1(cx)} ${r1(cy)}) rotate(${+deg.toFixed(1)})">` +
                  `<path d="M0 ${r1(h)}V${r1(-h + 1)}M-2.6 ${r1(-h + 4)}L0 ${r1(-h)}L2.6 ${r1(-h + 4)}" fill="none" stroke="${INK}" stroke-width="1.1" stroke-linecap="round" stroke-linejoin="round"/></g>`
                );
              })
              .join("") +
            (vals.length > shown.length ? `<text x="${r1(x + col.width - 3)}" y="${r1(box.yBase - 3)}" class="tiny" text-anchor="end">+${vals.length - shown.length}</text>` : "")
          );
        })
        .join(""),
  },

  facies: {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) => c.boxes.map((box) => (box.bed.facies?.trim() ? bedText(box, x, col.width, box.bed.facies, "note", true) : "")).join(""),
  },

  notes: {
    head: (c, col, x) => title(c, col, x),
    body: (c, col, x) => c.boxes.map((box) => (box.bed.notes?.trim() ? bedText(box, x, col.width, box.bed.notes) : "")).join(""),
  },
};

/** The whole log as an SVG document string. */
export function drawLog(log: Log, opt: DrawOptions): string {
  const pxPerM = pxPerMetre(log.scale);
  const yHead = TITLE;
  const yBody = TITLE + HEADER;
  const { boxes, bodyHeight } = layoutBeds(log, pxPerM, yBody);
  const yZero = yBody + bodyHeight;
  const ctx: Ctx = { log, boxes, pxPerM, yHead, yBody, yZero };
  const cols = visibleColumns(log);
  const colsWidth = cols.reduce((s, c) => s + c.width, 0);
  const width = Math.max(colsWidth, 240);
  const paged = opt.page !== undefined;
  const info = pageInfo(log);

  // On a page, the section is shown from `top` to `bottom` (drawing coordinates).
  const top = paged ? yBody + opt.page! * info.bodyPerPage : yBody;
  const bottom = paged ? Math.min(yZero, top + info.bodyPerPage) : yZero;
  const clipId = `body-clip-${opt.page ?? "all"}`;

  const key = opt.key ? drawKey(log, width) : undefined;
  const logBottom = Math.max(yBody + (bottom - top) + FOOT, yBody + (paged ? FOOT : 70));
  const height = logBottom + (key ? KEY_GAP + key.height : 0);

  let heads = "";
  let bodies = "";
  let headLines = "";
  let bodyLines = "";
  let x = 0;
  for (const col of cols) {
    heads += DRAWERS[col.kind].head(ctx, col, x);
    bodies += DRAWERS[col.kind].body(ctx, col, x);
    if (x > 0) {
      headLines += `<path d="M${r1(x)} ${yHead}V${yBody}" stroke="${INK}" stroke-width="0.9"/>`;
      bodyLines += `<path d="M${r1(x)} ${r1(top)}V${r1(bottom)}" stroke="${INK}" stroke-width="0.9"/>`;
    }
    x += col.width;
  }
  const shift = top - yBody; // how far the section is moved up on this page
  const yEnd = yBody + (bottom - top); // bottom of the frame
  let frame = "";
  if (cols.length) {
    frame += `<rect x="0.5" y="${yHead + 0.5}" width="${r1(colsWidth - 1)}" height="${r1(yEnd - yHead - 1)}" fill="none" stroke="${INK}" stroke-width="1"/>`;
    frame += `<path d="M0 ${yBody}H${r1(colsWidth)}" stroke="${INK}" stroke-width="1"/>`;
  }

  const pageNote = paged && info.pages > 1 ? ` – page ${opt.page! + 1} of ${info.pages}` : "";
  const heading =
    `<text x="2" y="17" class="logtitle">${esc(log.name + pageNote)}</text>` +
    `<text x="${r1(width - 2)}" y="17" class="tick" text-anchor="end">Scale 1:${log.scale}</text>`;

  let breaks = "";
  if (opt.pageBreaks && !paged) {
    for (let p = 1; p < info.pages; p++) {
      // Pages are filled from the top of the section down.
      const y = r1(yBody + p * info.bodyPerPage);
      breaks += `<path class="pagebreak" d="M0 ${y}H${r1(width)}" stroke="#3b7ea1" stroke-width="1" stroke-dasharray="6 4"/>`;
      breaks += `<text class="pagebreak tiny" x="${r1(width - 3)}" y="${r1(y - 3)}" text-anchor="end" style="fill:#3b7ea1">page ${p + 1} ↓</text>`;
    }
  }

  // Click areas for selecting beds; without the editor's styles they would show as black boxes.
  const hits = paged || !opt.handles
    ? ""
    : boxes
        .map(
          (b) =>
            `<rect class="hit${b.bed.id === opt.selectedId ? " selected" : ""}" data-bed="${esc(b.bed.id)}" x="0" y="${r1(b.yTop)}"` +
            ` width="${r1(colsWidth)}" height="${r1(Math.max(b.yBase - b.yTop, 2))}"/>`,
        )
        .join("");

  let handles = "";
  if (opt.handles && !paged) {
    // Kept about 8 screen px wide whatever the zoom. Later handles win where they overlap.
    const hw = 4 / opt.zoom;
    const grain = columnBox(log, "grain");
    if (grain) {
      for (const b of boxes) {
        const wb = grainWidth(b.bed.grain.base, grain.width);
        const wt = grainWidth(b.bed.grain.top ?? b.bed.grain.base, grain.width);
        handles +=
          `<path class="h-grain" data-bed="${esc(b.bed.id)}" d="M${r1(grain.x + wb)} ${r1(b.yBase - hw)}L${r1(grain.x + wt)} ${r1(b.yTop + hw)}"` +
          ` stroke-width="${r1(2 * hw)}"/>`;
        if (b.bed.id === opt.selectedId) {
          for (const [cx, cy] of [[grain.x + wb, b.yBase], [grain.x + wt, b.yTop]] as const)
            handles += `<circle class="knob" cx="${r1(cx)}" cy="${r1(cy)}" r="${r1(3.5 / opt.zoom)}"/>`;
        }
      }
    }
    for (const b of boxes) {
      handles += `<rect class="h-top" data-bed="${esc(b.bed.id)}" x="0" y="${r1(b.yTop - hw)}" width="${r1(colsWidth)}" height="${r1(2 * hw)}"/>`;
    }
    let cx = 0;
    for (const col of cols) {
      cx += col.width;
      handles += `<rect class="h-col" data-col="${log.columns.indexOf(col)}" x="${r1(cx - hw)}" y="${yHead}" width="${r1(2 * hw)}" height="${HEADER}"/>`;
    }
  }

  const empty = log.beds.length ? "" : `<text x="${r1(width / 2)}" y="${yBody + 40}" class="empty" text-anchor="middle">No beds yet. Add the lowest bed first.</text>`;

  // The section (and anything tied to it) is clipped to this page's slice when drawing pages.
  const section = paged
    ? `<clipPath id="${clipId}"><rect x="0" y="${yBody}" width="${r1(width)}" height="${r1(bottom - top + (bottom >= yZero ? FOOT : 0))}"/></clipPath>` +
      `<g clip-path="url(#${clipId})"><g transform="translate(0 ${r1(-shift)})">${bodies}${bodyLines}</g></g>`
    : bodies + bodyLines;

  return (
    `<svg xmlns="http://www.w3.org/2000/svg" class="logsvg" width="${r1(width * opt.zoom)}" height="${r1(height * opt.zoom)}"` +
    ` viewBox="0 0 ${r1(width)} ${r1(height)}" role="img" aria-label="Graphic log ${esc(log.name)}">` +
    defsAndStyle(log) +
    `<rect width="100%" height="100%" fill="#fff"/>` +
    heading +
    section +
    heads +
    headLines +
    frame +
    empty +
    breaks +
    hits +
    handles +
    (key ? `<g transform="translate(0 ${r1(logBottom + KEY_GAP)})">${key.svg}</g>` : "") +
    `</svg>`
  );
}

/** Patterns, symbols and text styles a drawing of this log needs. */
function defsAndStyle(log: Log): string {
  const fills = log.beds.flatMap((b) => b.lithology.map((l) => l.pattern));
  const inBed = [...new Set(log.beds.map((b) => b.inBed).filter((s): s is string => !!s && SYMBOL_BY_ID.has(s)))];
  const symbols = [...log.beds.flatMap((b) => b.symbols ?? []), ...inBed];
  const inBedDefs = inBed
    .map((id) => `<pattern id="inbed-${id}" patternUnits="userSpaceOnUse" width="22" height="18"><use href="#sym-${id}" x="2" y="0" width="18" height="18" color="${INK}"/></pattern>`)
    .join("");
  return (
    `<defs>${patternDefs(fills)}${symbolDefs(symbols)}${inBedDefs}</defs>` +
    `<style>text{font-family:system-ui,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;fill:${INK}}.colhead{font-size:10.5px;font-weight:600}.tick{font-size:9.5px}.tiny{font-size:8px;fill:${MUTED}}` +
    `.note{font-size:10.5px}.unit{font-size:10.5px;font-weight:600}.logtitle{font-size:13px;font-weight:700}.empty{font-size:12px;font-style:italic;fill:${MUTED}}.keyhead{font-size:11px;font-weight:700}</style>`
  );
}

/** A page with only the key, for when it does not fit under the log. */
export function drawKeyPage(log: Log, width: number): string {
  const key = drawKey(log, width);
  const height = TITLE + key.height + FOOT;
  return (
    `<svg xmlns="http://www.w3.org/2000/svg" class="logsvg" width="${r1(width)}" height="${r1(height)}" viewBox="0 0 ${r1(width)} ${r1(height)}">` +
    defsAndStyle(log) +
    `<rect width="100%" height="100%" fill="#fff"/>` +
    `<text x="2" y="17" class="logtitle">${esc(log.name)} – key</text>` +
    `<g transform="translate(0 ${TITLE + 4})">${key.svg}</g></svg>`
  );
}

// ---- key ------------------------------------------------------------------------------

const KEY_GAP = 14;
const KEY_ITEM_W = 160; // px per key entry, about 42 mm
const KEY_ROW = 19;

/** The key to the patterns, symbols and contacts a log uses, laid out `width` px wide. */
export function drawKey(log: Log, width: number): { svg: string; height: number } {
  const items: { art: string; label: string }[] = [];
  const seenFill = new Set<string>();
  for (const b of log.beds) {
    for (const l of b.lithology) {
      if (seenFill.has(l.pattern) || !PATTERN_BY_ID.has(l.pattern)) continue;
      seenFill.add(l.pattern);
      items.push({
        art: rockFill(`rect x="0" y="-9" width="28" height="13"`, l.pattern, lithColour(l.pattern, log.fill), `stroke="${INK}" stroke-width="0.7"`),
        label: PATTERN_BY_ID.get(l.pattern)!.name,
      });
    }
  }
  const symbolIds: string[] = [];
  for (const b of log.beds) for (const s of [...(b.inBed ? [b.inBed] : []), ...(b.symbols ?? [])]) if (SYMBOL_BY_ID.has(s) && !symbolIds.includes(s)) symbolIds.push(s);
  for (const s of symbolIds) {
    items.push({ art: `<use href="#sym-${s}" x="6" y="-11" width="16" height="16" color="${INK}"/>`, label: SYMBOL_BY_ID.get(s)!.name });
  }
  const contacts: [Bed["contact"], string][] = [["sharp", "Sharp contact"], ["erosional", "Erosional contact"], ["gradational", "Gradational contact"]];
  for (const [c, label] of contacts) {
    if (log.beds.some((b, i) => i > 0 && b.contact === c)) {
      items.push({ art: contact({ contact: c } as Bed, 0, 28, -3), label });
    }
  }
  if (!items.length) return { svg: "", height: 0 };

  const perRow = Math.max(1, Math.floor(width / KEY_ITEM_W));
  const rows = Math.ceil(items.length / perRow);
  let svg = `<text x="2" y="10" class="keyhead">Key</text>`;
  items.forEach((it, k) => {
    const x = (k % perRow) * KEY_ITEM_W + 2;
    const y = 30 + Math.floor(k / perRow) * KEY_ROW;
    svg += `<g transform="translate(${x} ${y})">${it.art}<text x="36" y="0" class="tick">${esc(it.label)}</text></g>`;
  });
  return { svg, height: 30 + rows * KEY_ROW - 8 };
}
