import { describe, expect, it } from "vitest";
import type { Bed, Log } from "../types";
import { layoutPages, svgSize, wholeLog } from "./pages";

const bed = (id: string, thickness: number, pattern = "sandstone"): Bed => ({ id, thickness, contact: "sharp", lithology: [{ pattern, percent: 100 }], grain: { base: 2.5 } });
const log = (beds: Bed[], extra: Partial<Log> = {}): Log => ({
  id: "l", projectId: "p", name: "Pages", revision: 1, created: "", updated: "",
  units: "m", scale: 50, paper: "A4", palaeocurrent: "arrows", fill: "colour",
  columns: [
    { kind: "scale", title: "Height", width: 46, visible: true },
    { kind: "lithology", title: "Lithology", width: 64, visible: true },
    { kind: "grain", title: "Grain size", width: 210, visible: true },
    { kind: "notes", title: "Notes", width: 200, visible: true },
  ],
  beds, ...extra,
});

describe("pages", () => {
  it("are the paper's size, with the drawing inside the margins", () => {
    const p = layoutPages(log([bed("a", 2)]), false);
    expect(p.svgs).toHaveLength(1);
    expect(p.svgs[0]).toMatch(/^<svg [^>]*width="210mm" height="297mm"/);
    expect(p.shrink).toBe(1);
  });
  it("shrink a log whose columns are wider than the paper", () => {
    const wide = log([bed("a", 2)]);
    wide.columns[3]!.width = 600; // 920 px in all, A4 has 703 px between the margins
    const p = layoutPages(wide, false);
    expect(p.shrink).toBeCloseTo(186 * (96 / 25.4) / 920, 3);
  });
  it("put the key on its own page when the last page is full", () => {
    const full = log([bed("a", 11.8)]); // nearly fills one A4 page at 1:50 (11.9 m)
    expect(layoutPages(full, false).svgs).toHaveLength(1);
    const withKey = layoutPages(full, true).svgs;
    expect(withKey).toHaveLength(2);
    expect(withKey[1]).toContain("Pages – key");
    const short = layoutPages(log([bed("a", 3)]), true).svgs;
    expect(short).toHaveLength(1); // room for the key under the log
    expect(short[0]).toContain(">Key<");
  });
  it("give the whole log its print size in mm", () => {
    const svg = wholeLog(log([bed("a", 1)]), false);
    const [w] = svgSize(svg);
    expect(svg).toContain(`width="${(w / (96 / 25.4)).toFixed(2)}mm"`);
  });
});

describe("CSV", () => {
  it("writes SedLog's columns, names and in-between grain sizes", async () => {
    const { toCSV } = await import("./formats");
    const l = log([
      { ...bed("a", 0.35, "tuff"), contact: "erosional", grain: { base: 1, top: 8 }, notes: 'Dark, "sooty"', symbols: ["bivalve", "columnar-joints"], palaeocurrents: [90, 100], formation: "Lower" },
      bed("b", 1.2, "lava"),
    ]);
    const text = (await toCSV(l).text()).replace(/^﻿/, "");
    const [head, first, second] = text.split("\r\n");
    expect(head!.split(",")).toHaveLength(27);
    expect(first).toBe('35,Erosion,Fine ash,100,<none>,0,<none>,0,m/c,1,clay/silt,8,<none>,"Bivalves, Columnar jointing","Dark, ""sooty""",<none>,0,"90, 100","","","","","","","",Lower,""');
    expect(second!.startsWith("120,Sharp,Lava,100,")).toBe(true);
    const sed = (await toCSV({ ...l, beds: [{ ...l.beds[0]!, facies: "St" }, { ...l.beds[1]!, facies: "4" }] }, true).text()).split("\r\n");
    expect(sed[1]).toContain(',"90, 100",0,'); // palaeocurrents, then facies: SedLog only takes numbers, so "St" becomes 0
    expect(sed[2]).toContain(',<none>,0,"",4,');
    expect(sed[1]).toContain(`"Dark, 'sooty'"`); // SedLog drops double quotes inside cells
    const firstBytes = async (b: Blob) => [...new Uint8Array(await b.arrayBuffer()).slice(0, 3)];
    expect(await firstBytes(toCSV(l))).toEqual([0xef, 0xbb, 0xbf]); // byte-order mark for Excel
    expect(await firstBytes(toCSV(l, true))).toEqual([..."THI"].map((c) => c.charCodeAt(0))); // SedLog refuses the mark
  });
});
