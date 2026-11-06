import { describe, expect, it } from "vitest";
import { GRAIN_CLASSES, grainClassOf, grainIndex } from "../grain";
import { parseDirections } from "../ui/bedpanel";
import type { Bed, Column, Log } from "../types";
import { columnBox, drawKey, drawLog, grainWidth, layoutBeds, pageInfo, phiAt, pxPerMetre, tickStep, units } from "./log";
import { PATTERNS, patternDefs } from "./patterns";
import { SYMBOLS } from "./symbols";

const COLUMNS: Column[] = [
  { kind: "age", title: "Age", width: 34, visible: false },
  { kind: "formation", title: "Formation", width: 34, visible: true },
  { kind: "scale", title: "Height", width: 46, visible: true },
  { kind: "lithology", title: "Lithology", width: 64, visible: true },
  { kind: "grain", title: "Grain size", width: 210, visible: true },
  { kind: "symbols", title: "Structures & fossils", width: 90, visible: true },
  { kind: "bioturbation", title: "Bioturbation", width: 44, visible: true },
  { kind: "palaeocurrent", title: "Palaeocurrent", width: 60, visible: true },
  { kind: "facies", title: "Facies", width: 44, visible: false },
  { kind: "notes", title: "Notes", width: 200, visible: true },
];

const bed = (id: string, thickness: number, extra: Partial<Bed> = {}): Bed => ({ id, thickness, contact: "sharp", lithology: [], grain: {}, ...extra });
const log = (beds: Bed[], extra: Partial<Log> = {}): Log => ({
  id: "l", projectId: "p", name: "Test <log>", revision: 1, created: "", updated: "",
  units: "m", scale: 50, paper: "A4", palaeocurrent: "arrows", fill: "colour", columns: COLUMNS.map((c) => ({ ...c })), beds, ...extra,
});

describe("grain size", () => {
  it("maps class phi values to their positions", () => {
    GRAIN_CLASSES.forEach((c, i) => expect(grainIndex(c.phi)).toBeCloseTo(i));
  });
  it("finds the class of a phi value", () => {
    expect(grainClassOf(2.5).key).toBe("f");
    expect(grainClassOf(20).key).toBe("clay");
    expect(grainClassOf(-20).key).toBe("boul");
  });
  it("draws coarser grains wider, boulders filling the column", () => {
    expect(grainWidth(-4)).toBeGreaterThan(grainWidth(2.5));
    expect(grainWidth(undefined)).toBe(grainWidth(9));
    expect(grainWidth(-9, 150)).toBeCloseTo(140); // boulders stop short of the column edge
  });
});

describe("scale and pages", () => {
  it("draws 1 m as 2 cm on paper at 1:50", () => {
    expect(pxPerMetre(50)).toBeCloseTo((20 * 96) / 25.4);
  });
  it("stacks beds from the bottom up", () => {
    const { boxes } = layoutBeds(log([bed("a", 2), bed("b", 1)]), 10, 0);
    expect(boxes[0]!.yBase - boxes[0]!.yTop).toBe(20);
    expect(boxes[1]!.yBase).toBe(boxes[0]!.yTop);
  });
  it("keeps scale labels readable", () => {
    expect(tickStep(60)).toBe(1);
    expect(tickStep(400)).toBe(0.1);
    expect(tickStep(5)).toBe(10);
  });
  it("counts pages and checks the width", () => {
    // 1:50 on A4 leaves about 273 mm − header for the section, i.e. under 13 m per page.
    expect(pageInfo(log([bed("a", 10)])).pages).toBe(1);
    expect(pageInfo(log([bed("a", 30)])).pages).toBe(3);
    expect(pageInfo(log([bed("a", 30)], { scale: 100 })).pages).toBe(2);
    expect(pageInfo(log([])).fitsWidth).toBe(false); // 748 px of columns > 186 mm
    const narrow = log([], { columns: COLUMNS.map((c) => ({ ...c, visible: c.kind === "lithology" || c.kind === "grain" })) });
    expect(pageInfo(narrow).fitsWidth).toBe(true);
  });
});

describe("drawLog", () => {
  const svg = drawLog(
    log([
      bed("x", 1, {
        contact: "erosional", lithology: [{ pattern: "sandstone", percent: 100 }], grain: { base: 1.5 }, notes: "Cross-bedded <sets>",
        symbols: ["bivalve", "roots"], inBed: "trough-cross-bedding", bioturbation: 3, palaeocurrents: [90], formation: "Grindavík Fm", facies: "St",
      }),
      bed("y", 0.5, { lithology: [{ pattern: "shale", percent: 100 }] }),
    ]),
    { zoom: 1, selectedId: "x", handles: true },
  );
  it("includes only the patterns and symbols used", () => {
    expect(svg).toContain('id="pat-sandstone"');
    expect(svg).toContain('fill="#f6e08a"'); // sandstone's colour under the pattern
    expect(svg).not.toContain('id="pat-coal');
    expect(svg).toContain('id="sym-bivalve"');
    expect(svg).toContain('id="inbed-trough-cross-bedding"');
    expect(svg).not.toContain('id="sym-ammonite"');
  });
  it("escapes text", () => {
    expect(svg).toContain("Cross-bedded &lt;sets&gt;");
    expect(svg).toContain("Test &lt;log&gt;");
  });
  it("draws the new columns", () => {
    expect(svg).toContain('href="#sym-roots"');
    expect(svg).toContain("rotate(90)"); // palaeocurrent arrow pointing east
    expect(svg).toContain("Grindavík Fm");
    expect(svg).toContain("Scale 1:50");
  });
  it("can leave out either part of the grain-size scale", () => {
    const l = log([bed("a", 1)]);
    const both = drawLog(l, { zoom: 1 });
    expect(both).toContain(">vf<");
    expect(both).toContain(">gravel<");
    const noClasses = drawLog({ ...l, hideGrainClasses: true }, { zoom: 1 });
    expect(noClasses).not.toContain(">vf<");
    expect(noClasses).toContain(">gravel<");
    const neither = drawLog({ ...l, hideGrainClasses: true, hideGrainGroups: true }, { zoom: 1 });
    expect(neither).not.toContain(">vf<");
    expect(neither).not.toContain(">gravel<");
    expect(neither).toContain(">Grain size<"); // the column title stays
    const bare = drawLog({ ...l, hideGrainClasses: true, hideGrainGroups: true, hideGrainTitle: true }, { zoom: 1 });
    expect(bare).not.toContain(">Grain size<");
    expect(drawLog({ ...l, hideGrainTitle: true }, { zoom: 1 })).not.toContain(">Grain size<");
  });
  it("leaves hidden columns out", () => {
    expect(svg).not.toContain(">St<"); // facies column is hidden
  });
  it("marks the selected bed", () => {
    expect(svg).toMatch(/class="hit selected" data-bed="x"/);
  });
  it("scales with the zoom but keeps the viewBox", () => {
    const a = drawLog(log([bed("a", 1)]), { zoom: 1 });
    const b = drawLog(log([bed("a", 1)]), { zoom: 2 });
    const w = (s: string) => Number(/ width="([\d.]+)"/.exec(s)![1]);
    expect(w(b)).toBeCloseTo(2 * w(a));
    expect(/viewBox="([^"]+)"/.exec(a)![1]).toBe(/viewBox="([^"]+)"/.exec(b)![1]);
  });
  it("shows page breaks when asked", () => {
    expect(drawLog(log([bed("a", 30)]), { zoom: 1, pageBreaks: true })).toContain("page 3");
  });
  it("fills lithologies with their colour, a bed's own colour, or white", () => {
    const beds = [bed("a", 1, { lithology: [{ pattern: "limestone", percent: 70 }, { pattern: "shale", percent: 30, colour: "#c0392b" }] })];
    const colour = drawLog(log(beds), { zoom: 1 });
    expect(colour).toContain('fill="#a9d0ea"/><rect');
    expect(colour).toContain('fill="#c0392b"');
    expect(colour).toContain('fill="url(#pat-shale)"');
    const bw = drawLog(log(beds, { fill: "white" }), { zoom: 1 });
    expect(bw).not.toContain("a9d0ea");
    expect(bw).not.toContain("c0392b");
  });
  it("keeps the rock colour behind an in-bed structure", () => {
    expect(svg).toMatch(/<path d="M[^"]+Z" fill="#f6e08a"\/>/);
  });
  it("prompts for the first bed when empty", () => {
    expect(drawLog(log([]), { zoom: 1 })).toContain("No beds yet");
  });
});

describe("editing on the drawing", () => {
  it("finds the grain class back from a profile width", () => {
    for (const w of [150, 210]) for (const c of GRAIN_CLASSES) expect(phiAt(grainWidth(c.phi, w) + 3, w)).toBe(c.phi);
    expect(phiAt(-50)).toBe(GRAIN_CLASSES[0]!.phi);
    expect(phiAt(9999)).toBe(GRAIN_CLASSES[GRAIN_CLASSES.length - 1]!.phi);
  });
  it("places columns left to right", () => {
    const l = log([]);
    expect(columnBox(l, "lithology")!.x).toBe(80); // after formation (34) and scale (46)
    expect(columnBox(l, "facies")).toBeUndefined(); // hidden
  });
  it("adds drag handles only when asked", () => {
    const l = log([bed("a", 1), bed("b", 1)]);
    const plain = drawLog(l, { zoom: 1 });
    const edit = drawLog(l, { zoom: 2, handles: true, selectedId: "b" });
    expect(plain).not.toContain("h-top");
    expect(plain).not.toContain('class="hit'); // exports must not carry the editor's click areas
    expect(edit.match(/class="h-top"/g)).toHaveLength(2);
    expect(edit.match(/class="h-grain"/g)).toHaveLength(2);
    expect(edit.match(/class="h-col"/g)).toHaveLength(COLUMNS.filter((c) => c.visible).length);
    expect(edit.match(/class="knob"/g)).toHaveLength(2); // grain corners of the selected bed
  });
});

describe("pages and key", () => {
  // 30 m at 1:50 on A4: three pages.
  const long = log([bed("a", 10, { contact: "sharp", lithology: [{ pattern: "sandstone", percent: 100 }] }), bed("b", 10, { contact: "erosional", symbols: ["roots"] }), bed("c", 10, { lithology: [{ pattern: "limestone", percent: 100, colour: "#c0392b" }] })]);
  it("draws one page at a time with the headers repeated", () => {
    const info = pageInfo(long);
    expect(info.pages).toBe(3);
    const p2 = drawLog(long, { zoom: 1, page: 1 });
    expect(p2).toContain("page 2 of 3");
    expect(p2).toContain(">Lithology<"); // column headers on every page
    expect(p2).toContain('clip-path="url(#body-clip-1)"');
    expect(p2).not.toContain('class="hit');
    const h = (s: string) => Number(/viewBox="0 0 [\d.]+ ([\d.]+)"/.exec(s)![1]);
    expect(h(p2)).toBeLessThanOrEqual((297 - 24) * (96 / 25.4) + 1); // fits the page height
  });
  it("makes a key of what the log uses", () => {
    const k = drawKey(long, 600);
    expect(k.svg).toContain(">Sandstone<");
    expect(k.svg).toContain(">Limestone<");
    expect(k.svg).toContain(">Roots<");
    expect(k.svg).toContain(">Erosional contact<");
    expect(k.svg).not.toContain("Gradational");
    expect(k.svg).toContain('fill="#a9d0ea"'); // the usual colour, not the bed's own
    const svg = drawLog(long, { zoom: 1, key: true });
    expect(svg).toContain('id="pat-limestone"');
    expect(drawKey(log([]), 600).height).toBe(0);
  });
});

describe("formations", () => {
  it("run from the bed that names them up to the next named bed", () => {
    const l = log([bed("a", 1, { formation: "Lower" }), bed("b", 1), bed("c", 1, { formation: "Upper" }), bed("d", 1)]);
    const { boxes } = layoutBeds(l, 10, 0);
    const u = units(boxes, "formation");
    expect(u.map((x) => x.name)).toEqual(["Lower", "Upper"]);
    expect(u[0]!.yBase - u[0]!.yTop).toBe(20);
    expect(u[1]!.yTop).toBe(0);
  });
});

describe("input parsing", () => {
  it("reads palaeocurrent lists", () => {
    expect(parseDirections("85, 120 95;360")).toEqual([85, 120, 95, 360]);
    expect(parseDirections("")).toEqual([]);
    expect(parseDirections("85, north")).toBeNull();
    expect(parseDirections("400")).toBeNull();
  });
});

/** Elements in an SVG string that repeat an attribute (which XML, and so PDF and PNG export, refuses). */
function repeatedAttributes(svg: string): string[] {
  const bad: string[] = [];
  for (const tag of svg.match(/<[a-zA-Z][^>]*>/g) ?? []) {
    const names = [...tag.matchAll(/\s([\w:-]+)=/g)].map((m) => m[1]);
    if (new Set(names).size !== names.length) bad.push(tag.slice(0, 80));
  }
  return bad;
}

describe("libraries", () => {
  it("every pattern and symbol is valid XML (no repeated attributes)", () => {
    const all = log([bed("a", 1, { lithology: PATTERNS.slice(0, 3).map((p) => ({ pattern: p.id, percent: 33 })), symbols: SYMBOLS.map((s) => s.id) })]);
    expect(repeatedAttributes(drawLog(all, { zoom: 1, key: true }))).toEqual([]);
    expect(repeatedAttributes(drawLog(all, { zoom: 1, page: 0, handles: true, selectedId: "a" }))).toEqual([]);
    expect(repeatedAttributes(patternDefs(PATTERNS.map((p) => p.id)))).toEqual([]);
  });
  it("patterns have unique ids, colours and valid tiles", () => {
    for (const p of PATTERNS) expect(p.colour).toMatch(/^#[0-9a-f]{6}$/);
    expect(new Set(PATTERNS.map((p) => p.id)).size).toBe(PATTERNS.length);
    for (const p of PATTERNS) expect(patternDefs([p.id])).toContain(`width="${p.size}"`);
  });
  it("symbols have unique ids", () => {
    expect(new Set(SYMBOLS.map((s) => s.id)).size).toBe(SYMBOLS.length);
    expect(SYMBOLS.filter((s) => s.inBed).length).toBeGreaterThan(5);
  });
});

describe("printing", () => {
  it("gives each printed page its own pattern and symbol names", async () => {
    const { ownIds } = await import("../export/formats");
    const svg = drawLog(log([bed("a", 1, { lithology: [{ pattern: "shale", percent: 100 }], symbols: ["roots"] })]), { zoom: 1, page: 0 });
    const out = ownIds(svg, "print0-");
    expect(out).toContain('id="print0-pat-shale"');
    expect(out).toContain('fill="url(#print0-pat-shale)"');
    expect(out).toContain('href="#print0-sym-roots"');
    expect(out).toContain('clip-path="url(#print0-body-clip-0)"');
    expect(out).not.toMatch(/url\(#(?!print0-)/);
  });
});
