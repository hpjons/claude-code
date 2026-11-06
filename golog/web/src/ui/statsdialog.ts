import { api } from "../api";
import { cell, download, fileName } from "../export/formats";
import { esc } from "../render/log";
import { PATTERN_BY_ID, lithColour } from "../render/patterns";
import type { Log, LogContent, OrderTest, Share, Statistics } from "../types";
import { h, plural, toast } from "./dom";

const CONTACT_NAMES: Record<string, string> = { erosional: "Erosional", sharp: "Sharp", gradational: "Gradational", unknown: "Not recorded" };
const GRAIN_NAMES: Record<string, string> = { gravel: "Gravel", sand: "Sand", mud: "Mud", "": "Not recorded" };
const TREND_NAMES: Record<string, string> = { fining: "Fining upward", coarsening: "Coarsening upward", constant: "No change", "": "Not recorded" };

function content(log: Log): LogContent {
  const { units, scale, paper, palaeocurrent, fill, columns, beds } = log;
  return { units, scale, paper, palaeocurrent, fill, columns, beds };
}

const num = (v: number, digits = 2) => v.toFixed(digits);

/** The statistics window: what the section is made of, and whether its beds are ordered. */
export async function openStats(log: Log): Promise<void> {
  const body = h("div", { class: "stats-body" }, h("p", { class: "hint", role: "status" }, "Calculating…"));
  const saveBtn = h("button", { type: "button", name: "stats-csv", disabled: true }, "Save as CSV");
  const dialog = h(
    "dialog",
    { class: "picker stats", "aria-label": "Statistics" },
    h("header", {}, h("h2", {}, `Statistics for ${log.name}`)),
    body,
    h("footer", {}, saveBtn, h("button", { type: "button", onclick: () => dialog.close() }, "Close")),
  );
  dialog.addEventListener("close", () => dialog.remove());
  document.body.append(dialog);
  dialog.showModal();

  let st: Statistics;
  try {
    st = await api.stats(content(log));
  } catch (err) {
    body.replaceChildren(h("p", { class: "warn" }, `Could not calculate the statistics: ${(err as Error).message}`));
    return;
  }
  const len = (m: number) => (log.units === "cm" ? `${+(m * 100).toFixed(1)} cm` : `${+m.toFixed(2)} m`);
  saveBtn.disabled = false;
  saveBtn.addEventListener("click", () => {
    download(fileName(log, "statistics.csv"), statsCSV(log, st));
    toast(`Statistics saved as “${fileName(log, "statistics.csv")}”.`);
  });

  const s = st.summary;
  if (s.beds === 0) {
    body.replaceChildren(h("p", {}, "This log has no beds yet. Add beds, then open the statistics again."));
    return;
  }
  const parts: Node[] = [];
  if (st.notes?.length) parts.push(h("div", { class: "stats-notes" }, ...st.notes.map((n) => h("p", { class: "hint" }, n))));

  // The section
  parts.push(
    h(
      "section",
      { class: "stats-section", "aria-labelledby": "st-sec" },
      h("h3", { id: "st-sec" }, "The section"),
      h(
        "dl",
        { class: "stats-figures" },
        ...figure("Beds", String(s.beds)),
        ...figure("Total thickness", len(s.thickness)),
        ...figure("Mean bed", len(s.meanBed)),
        ...figure("Median bed", len(s.medianBed)),
        ...figure("Thinnest – thickest", `${len(s.minBed)} – ${len(s.maxBed)}`),
        ...figure("Standard deviation", len(s.sdBed)),
      ),
      h(
        "div",
        { class: "stats-tables" },
        shareTable("Lithology", s.lithologies, (k) => (k ? (PATTERN_BY_ID.get(k)?.name ?? k) : "Not recorded"), len, (k) => (k ? lithColour(k, log.fill) : undefined)),
        shareTable("Grain size", s.grain, (k) => GRAIN_NAMES[k] ?? k, len),
        shareTable("Within beds", s.trends, (k) => TREND_NAMES[k] ?? k, len),
        shareTable("Contact at base", s.contacts, (k) => CONTACT_NAMES[k] ?? k, len),
        shareTable("Facies", s.facies, (k) => k || "Not recorded", len),
      ),
    ),
  );

  // Facies order
  if (st.facies) {
    const f = st.facies;
    const grid = h("div", { class: "stats-matrix" });
    grid.innerHTML = matrixSVG(f.codes, f.prob, f.counts);
    parts.push(
      h(
        "section",
        { class: "stats-section", "aria-labelledby": "st-fac" },
        h("h3", { id: "st-fac" }, "Are the facies ordered?"),
        h(
          "p",
          { class: "stats-lead" },
          `How often each facies is followed by each other one, from ${plural(f.transitions, "change")} of facies up the section. `,
          "In an ordered section each facies is mostly followed by one other.",
        ),
        h("div", { class: "stats-pair" }, grid, histogram(f.m, "m", "Markov order metric m", 1)),
        verdict(f.m, "m"),
      ),
    );
  }

  // Thickness order
  if (st.runs) {
    const r = st.runs;
    const up = r.thickening.filter((x) => x > 0).length;
    const down = r.thinning.filter((x) => x > 0).length;
    const longest = (xs: number[]) => Math.max(0, ...xs) + (Math.max(0, ...xs) > 0 ? 1 : 0);
    parts.push(
      h(
        "section",
        { class: "stats-section", "aria-labelledby": "st-runs" },
        h("h3", { id: "st-runs" }, "Do beds thicken or thin upward?"),
        h(
          "p",
          { class: "stats-lead" },
          `${up} beds are thicker than the bed below and ${down} are thinner. `,
          `The longest thickening run is ${plural(longest(r.thickening), "bed")}, the longest thinning run ${plural(longest(r.thinning), "bed")}. `,
          "Long runs give a high r.",
        ),
        histogram(r.r, "r", "Runs order metric r", 3),
        verdict(r.r, "r"),
      ),
    );
  }

  if (st.facies || st.runs) {
    parts.push(
      h(
        "p",
        { class: "hint stats-method" },
        "Each test compares this log with the same beds shuffled into random order many times (a swap is skipped when it would put two beds of the same facies together). ",
        "p is the share of shuffled copies that score as high as this log or higher: a small p means the order is unlikely to be chance. ",
        "Method: Burgess, P.M. (2016), Identifying ordered strata: evidence, methods, and meaning. Journal of Sedimentary Research 86, 148–167.",
      ),
    );
  }
  body.replaceChildren(...parts);
}

function figure(label: string, value: string): Node[] {
  return [h("div", {}, h("dt", {}, label), h("dd", {}, value))];
}

function shareTable(title: string, shares: Share[], name: (k: string) => string, len: (m: number) => string, colour?: (k: string) => string | undefined): HTMLElement {
  if (shares.length === 0 || (shares.length === 1 && shares[0]!.key === "")) {
    return h("div", { class: "stats-share" }, h("h4", {}, title), h("p", { class: "hint" }, "Not recorded"));
  }
  return h(
    "div",
    { class: "stats-share" },
    h("h4", {}, title),
    h(
      "table",
      {},
      h("thead", {}, h("tr", {}, h("th", { scope: "col" }, "", h("span", { class: "sr" }, title)), h("th", { scope: "col" }, "Beds"), h("th", { scope: "col" }, "Thickness"), h("th", { scope: "col" }, "%"))),
      h(
        "tbody",
        {},
        ...shares.map((sh) => {
          const c = colour?.(sh.key);
          return h(
            "tr",
            { class: sh.key === "" ? "none" : undefined },
            h("th", { scope: "row" }, c ? h("span", { class: "swatch", style: `background:${c}` }) : null, name(sh.key)),
            h("td", {}, String(sh.beds)),
            h("td", {}, len(sh.thickness)),
            h("td", {}, h("span", { class: "pct-bar", style: `--w:${Math.min(100, sh.percent).toFixed(1)}` }), num(sh.percent, 0)),
          );
        }),
      ),
    ),
  );
}

/** The transition probabilities as a grid: rows are the facies below, columns the facies above. */
export function matrixSVG(codes: string[], prob: number[][], counts: number[][]): string {
  const n = codes.length;
  const cellSize = n > 10 ? 30 : 40;
  const left = 44;
  const top = 34;
  const w = left + n * cellSize + 4;
  const ht = top + n * cellSize + 40;
  const out: string[] = [];
  out.push(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${w} ${ht}" width="${w}" height="${ht}" role="img" aria-label="Facies transition probabilities">`);
  out.push(`<text x="${left + (n * cellSize) / 2}" y="12" class="axis-title" text-anchor="middle">to the facies above</text>`);
  // rows from the top in reading order (first facies at the top)
  for (let i = 0; i < n; i++) {
    const y = top + i * cellSize;
    out.push(`<text x="${left - 6}" y="${y + cellSize / 2 + 4}" text-anchor="end" class="tick">${esc(codes[i]!)}</text>`);
    for (let j = 0; j < n; j++) {
      const x = left + j * cellSize;
      if (i === j) {
        out.push(`<rect x="${x}" y="${y}" width="${cellSize}" height="${cellSize}" class="cell-self"><title>${esc(codes[i]!)} → ${esc(codes[i]!)}: not counted (same facies)</title></rect>`);
        out.push(`<text x="${x + cellSize / 2}" y="${y + cellSize / 2 + 4}" text-anchor="middle" class="tick">–</text>`);
        continue;
      }
      const p = prob[i]![j]!;
      out.push(
        `<rect x="${x}" y="${y}" width="${cellSize}" height="${cellSize}" class="cell" fill-opacity="${(0.06 + 0.94 * p).toFixed(3)}">` +
          `<title>${esc(codes[i]!)} → ${esc(codes[j]!)}: ${p.toFixed(2)} (${counts[i]![j]} of ${counts[i]!.reduce((a, b) => a + b, 0)})</title></rect>`,
      );
      if (p > 0) out.push(`<text x="${x + cellSize / 2}" y="${y + cellSize / 2 + 4}" text-anchor="middle" class="${p > 0.5 ? "val strong" : "val"}">${p.toFixed(2).replace(/^0/, "")}</text>`);
    }
  }
  for (let j = 0; j < n; j++) out.push(`<text x="${left + j * cellSize + cellSize / 2}" y="${top - 6}" text-anchor="middle" class="tick">${esc(codes[j]!)}</text>`);
  out.push(`<text x="${left + (n * cellSize) / 2}" y="${ht - 8}" text-anchor="middle" class="axis-title">rows: the facies below</text>`);
  out.push(`</svg>`);
  return out.join("");
}

/** Shuffled values as bars, and this log's value as a line. */
export function histogram(t: OrderTest, sym: string, title: string, minMax: number): HTMLElement {
  const W = 440;
  const H = 200;
  const L = 44;
  const R = 12;
  const T = 22;
  const B = 40;
  const counts = t.bins.counts;
  const xMax = Math.max(minMax, Math.ceil((t.bins.start + counts.length * t.bins.width) * 2) / 2, Math.ceil(t.observed * 2 + 0.01) / 2);
  const total = counts.reduce((a, b) => a + b, 0) || 1;
  const yMaxRaw = Math.max(...counts, 0) / total;
  const yStep = yMaxRaw > 0.5 ? 0.2 : yMaxRaw > 0.2 ? 0.1 : 0.05;
  const yMax = Math.max(yStep, Math.ceil(yMaxRaw / yStep - 1e-9) * yStep);
  const px = (v: number) => L + (v / xMax) * (W - L - R);
  const py = (v: number) => H - B - (v / yMax) * (H - B - T);
  const out: string[] = [];
  out.push(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${W} ${H}" width="100%" role="img" aria-label="${esc(title)}: this log ${t.observed.toFixed(2)} against ${t.shuffles} shuffled copies">`);
  for (let v = 0; v <= yMax + 1e-9; v += yStep) {
    out.push(`<line x1="${L}" x2="${W - R}" y1="${py(v)}" y2="${py(v)}" class="grid"/>`);
    out.push(`<text x="${L - 6}" y="${py(v) + 4}" text-anchor="end" class="tick">${+v.toFixed(2)}</text>`);
  }
  const xStep = xMax <= 1 ? 0.1 : xMax <= 3 ? 0.5 : 1;
  for (let v = 0; v <= xMax + 1e-9; v += xStep) out.push(`<text x="${px(v)}" y="${H - B + 16}" text-anchor="middle" class="tick">${+v.toFixed(1)}</text>`);
  counts.forEach((c, i) => {
    if (!c) return;
    const x0 = px(t.bins.start + i * t.bins.width);
    const x1 = px(t.bins.start + (i + 1) * t.bins.width);
    out.push(`<rect x="${x0 + 0.5}" y="${py(c / total)}" width="${Math.max(1, x1 - x0 - 1)}" height="${H - B - py(c / total)}" class="hbar"/>`);
  });
  out.push(`<line x1="${L}" x2="${W - R}" y1="${H - B}" y2="${H - B}" class="axis"/>`);
  const ox = px(t.observed);
  out.push(`<line x1="${ox}" x2="${ox}" y1="${T - 8}" y2="${H - B}" class="observed"/>`);
  const right = ox > W - 130;
  out.push(`<text x="${ox + (right ? -5 : 5)}" y="${T}" text-anchor="${right ? "end" : "start"}" class="observed-label">this log: ${sym} = ${t.observed.toFixed(2)}</text>`);
  out.push(`<text x="${L + (W - L - R) / 2}" y="${H - 6}" text-anchor="middle" class="axis-title">${esc(title)}; bars: ${t.shuffles} shuffled copies</text>`);
  out.push(`</svg>`);
  const fig = h("figure", { class: "stats-hist" });
  fig.innerHTML = out.join("");
  return fig;
}

function verdict(t: OrderTest, sym: string): HTMLElement {
  const n = Math.round(t.p * t.shuffles);
  const p = t.p < 0.001 ? "p < 0.001" : `p = ${t.p < 0.01 ? t.p.toFixed(3) : t.p.toFixed(2)}`;
  const strong = t.p < 0.05;
  return h(
    "p",
    { class: `stats-verdict${strong ? " yes" : ""}` },
    h("strong", {}, `${sym} = ${t.observed.toFixed(2)}`),
    ` (shuffled copies average ${t.mean.toFixed(2)}). ${n} of ${t.shuffles} shuffled copies score as high or higher, ${p}: `,
    h("strong", {}, t.evidence),
    ".",
  );
}

/** All the numbers in one CSV file, for a spreadsheet. */
export function statsCSV(log: Log, st: Statistics): Blob {
  const rows: (string | number)[][] = [];
  const s = st.summary;
  const r4 = (v: number) => +v.toFixed(4);
  rows.push(["GoLog statistics", log.name]);
  rows.push([]);
  rows.push(["Beds", s.beds]);
  rows.push(["Total thickness (m)", r4(s.thickness)]);
  rows.push(["Mean bed (m)", r4(s.meanBed)], ["Median bed (m)", r4(s.medianBed)], ["Thinnest bed (m)", r4(s.minBed)], ["Thickest bed (m)", r4(s.maxBed)], ["Standard deviation (m)", r4(s.sdBed)]);
  const table = (title: string, shares: Share[], name: (k: string) => string) => {
    rows.push([], [title, "Beds", "Thickness (m)", "Percent"]);
    for (const sh of shares) rows.push([name(sh.key), sh.beds, r4(sh.thickness), +sh.percent.toFixed(2)]);
  };
  table("Lithology", s.lithologies, (k) => (k ? (PATTERN_BY_ID.get(k)?.name ?? k) : "Not recorded"));
  table("Grain size", s.grain, (k) => GRAIN_NAMES[k] ?? k);
  table("Within beds", s.trends, (k) => TREND_NAMES[k] ?? k);
  table("Contact at base", s.contacts, (k) => CONTACT_NAMES[k] ?? k);
  table("Facies", s.facies, (k) => k || "Not recorded");
  const test = (name: string, t: OrderTest) => {
    rows.push([], [name, r4(t.observed)], ["Mean of shuffled copies", r4(t.mean)], ["Shuffled copies", t.shuffles], ["p", r4(t.p)], ["Evidence", t.evidence]);
  };
  if (st.facies) {
    const f = st.facies;
    rows.push([], ["Facies transition probabilities (row: facies below, column: facies above)"], ["", ...f.codes]);
    f.prob.forEach((row, i) => rows.push([f.codes[i]!, ...row.map((v) => +v.toFixed(4))]));
    rows.push([], ["Facies transition counts"], ["", ...f.codes]);
    f.counts.forEach((row, i) => rows.push([f.codes[i]!, ...row]));
    test("Markov order metric m", f.m);
  }
  if (st.runs) test("Runs order metric r", st.runs.r);
  if (st.notes?.length) rows.push([], ...st.notes.map((n) => ["Note", n]));
  return new Blob(["﻿" + rows.map((r) => r.map(cell).join(",")).join("\r\n") + "\r\n"], { type: "text/csv;charset=utf-8" });
}
