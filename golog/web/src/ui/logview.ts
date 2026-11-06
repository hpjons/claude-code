import { api, ApiError } from "../api";
import { grainClassOf } from "../grain";
import { columnBox, drawLog, layoutBeds, pageInfo, phiAt, pxPerMetre } from "../render/log";
import type { Bed, Log } from "../types";
import { bedFields } from "./bedpanel";
import { openExport } from "./exportdialog";
import { openStats } from "./statsdialog";
import { h, plural, toast } from "./dom";
import { History } from "./history";
import { layoutFields } from "./layoutpanel";

const ZOOMS = [0.5, 0.75, 1, 1.25, 1.5, 2, 3]; // screen size relative to print size
const SCALES = [10, 20, 25, 50, 100, 200, 250, 500, 1000];
const MM = 96 / 25.4;

const newId = () => (crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2) + Date.now().toString(36));

/**
 * Saves the log shortly after each change. Only one save is in flight at a time;
 * changes made during a save are sent once it finishes.
 */
class Saver {
  private timer = 0;
  private inflight = false;
  private dirty = false;

  constructor(
    private log: () => Log,
    private onSaved: (saved: Log) => void,
    private onConflict: (current: Log) => void,
    private status: (s: "saved" | "saving" | "error", msg?: string) => void,
  ) {}

  get pending(): boolean {
    return this.dirty || this.inflight;
  }

  changed(): void {
    this.dirty = true;
    this.status("saving");
    clearTimeout(this.timer);
    this.timer = window.setTimeout(() => this.save(), 500);
  }

  async save(): Promise<void> {
    if (this.inflight || !this.dirty) return;
    this.inflight = true;
    this.dirty = false;
    try {
      this.onSaved(await api.saveLog(this.log()));
      this.status(this.dirty ? "saving" : "saved");
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        this.onConflict((err.body as { current: Log }).current);
      } else {
        this.dirty = true;
        this.status("error", (err as Error).message);
        this.timer = window.setTimeout(() => this.save(), 5000);
      }
    } finally {
      this.inflight = false;
    }
    if (this.dirty) this.save();
  }
}

/** Reads and writes a per-browser view setting; works (without memory) if storage is blocked. */
const viewSetting = {
  get(key: string): string | null {
    try {
      return localStorage.getItem(key);
    } catch {
      return null;
    }
  },
  set(key: string, value: string): void {
    try {
      localStorage.setItem(key, value);
    } catch {
      /* not remembered, but still works */
    }
  },
};

/** The parts of a log that undo and redo bring back. */
const CONTENT = ["name", "units", "scale", "paper", "palaeocurrent", "fill", "columns", "beds"] as const;

/** Thickness steps a drag snaps to: the smallest that is at least a few screen px. */
function snapStep(pxPerScreenM: number): number {
  for (const s of [0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10]) if (s * pxPerScreenM >= 4) return s;
  return 10;
}

/** Text fields and colour pickers merge quick edits into one undo step; any other change is a step of its own. */
function editSource(): Element | null {
  const el = document.activeElement;
  return el instanceof HTMLTextAreaElement || (el instanceof HTMLInputElement && !["checkbox", "radio"].includes(el.type)) ? el : null;
}

/** Text fields have their own Ctrl+Z. */
const isTyping = (el: EventTarget | null) =>
  el instanceof HTMLTextAreaElement || (el instanceof HTMLInputElement && !["checkbox", "radio", "color", "button"].includes(el.type));

type Drag =
  | { kind: "top"; bed: Bed; y0: number; t0: number; h0: number; moved: boolean }
  | { kind: "grain"; bed: Bed; end: "base" | "top"; moved: boolean }
  | { kind: "col"; col: number; x0: number; w0: number; moved: boolean }
  | { kind: "move"; bed: Bed; y0: number; to: number; moved: boolean };

/** The log editor: drawing on the left, bed and layout panel on the right. */
export async function logView(root: HTMLElement, id: string): Promise<() => boolean> {
  let log = await api.log(id);
  const project = await api.project(log.projectId);
  let selected: string | undefined;
  let tab: "bed" | "layout" = "bed";

  const zoomKey = `golog.zoom2.${id}`;
  let zoom = ZOOMS.indexOf(Number(viewSetting.get(zoomKey)));
  if (zoom < 0) zoom = ZOOMS.indexOf(1);

  const drawing = h("div", {
    class: "drawing", tabindex: 0,
    "aria-label": "Log drawing. Up and down arrows move between beds, Alt with an arrow moves the bed, Delete deletes it.",
  });
  const panelBody = h("div", { class: "panel-body" });
  const statusEl = h("span", { class: "save-status saved", role: "status" }, "Saved");
  const zoomLabel = h("span", { class: "zoom-label" });
  const pagesEl = h("span", { class: "pages" });
  const tip = h("div", { class: "drag-tip", hidden: true });

  const history = new History(
    () => JSON.stringify(CONTENT.map((k) => log[k])),
    (state) => {
      const values = JSON.parse(state) as unknown[];
      CONTENT.forEach((k, i) => ((log as unknown as Record<string, unknown>)[k] = values[i]));
      if (!log.beds.some((b) => b.id === selected)) selected = undefined;
      syncBar();
      redraw();
      showPanel();
      saver.changed();
    },
  );

  const saver = new Saver(
    () => log,
    (saved) => {
      log.revision = saved.revision;
      log.updated = saved.updated;
    },
    (current) => {
      log = current;
      selected = undefined;
      history.reset();
      syncUndo();
      toast("This log was changed in another window. Showing the latest version.", "error");
      syncBar();
      redraw();
      showPanel();
      statusEl.textContent = "Saved";
    },
    (s, msg) => {
      statusEl.textContent = s === "saved" ? "Saved" : s === "saving" ? "Saving…" : "Not saved – retrying";
      statusEl.className = `save-status ${s}`;
      statusEl.title = msg ?? "";
    },
  );

  /** Every edit ends here: redraw, remember for undo, save. */
  const changed = (source: Element | null = editSource()) => {
    redraw();
    history.record(source);
    saver.changed();
    syncUndo();
  };

  function redraw(): void {
    const scroll = { top: drawing.scrollTop, left: drawing.scrollLeft };
    drawing.innerHTML = drawLog(log, { zoom: ZOOMS[zoom]!, selectedId: selected, pageBreaks: true, handles: true });
    drawing.scrollTop = scroll.top;
    drawing.scrollLeft = scroll.left;
    zoomLabel.textContent = `${Math.round(ZOOMS[zoom]! * 100)} %`;
    const info = pageInfo(log);
    pagesEl.textContent = `${plural(info.pages, "page")} ${log.paper}` + (info.fitsWidth ? "" : " · too wide");
    pagesEl.classList.toggle("warn", !info.fitsWidth);
  }

  function select(bedId: string | undefined): void {
    selected = bedId;
    if (bedId) tab = "bed";
    redraw();
    showPanel();
    drawing.querySelector(".hit.selected")?.scrollIntoView({ block: "nearest" });
  }

  // ---- editing on the drawing ------------------------------------------------------------

  let drag: Drag | undefined;
  let swallowClick = false;
  let frame = 0;

  /** Pointer position in drawing coordinates. */
  const at = (e: PointerEvent) => {
    const r = drawing.querySelector("svg")!.getBoundingClientRect();
    const z = ZOOMS[zoom]!;
    return { x: (e.clientX - r.left) / z, y: (e.clientY - r.top) / z };
  };
  const bedNo = (bed: Bed) => log.beds.indexOf(bed) + 1;
  const showTip = (e: PointerEvent, text: string) => {
    tip.textContent = text;
    tip.hidden = false;
    tip.style.left = `${e.clientX + 14}px`;
    tip.style.top = `${e.clientY + 12}px`;
  };
  const fmt = (m: number) => (log.units === "cm" ? `${+(m * 100).toFixed(1)} cm` : `${+m.toFixed(3)} m`);
  const later = (f: () => void) => {
    cancelAnimationFrame(frame);
    frame = requestAnimationFrame(f);
  };

  drawing.addEventListener("pointerdown", (e) => {
    if (e.button !== 0 || !(e.target instanceof Element)) return;
    const t = e.target;
    const bed = log.beds.find((b) => b.id === t.getAttribute("data-bed"));
    const p = at(e);
    if (t.classList.contains("h-top") && bed) {
      drag = { kind: "top", bed, y0: e.clientY, t0: bed.thickness, h0: drawing.scrollHeight, moved: false };
    } else if (t.classList.contains("h-grain") && bed) {
      const box = layoutBeds(log, pxPerMetre(log.scale)).boxes.find((b) => b.bed === bed)!;
      drag = { kind: "grain", bed, end: p.y > (box.yTop + box.yBase) / 2 ? "base" : "top", moved: false };
    } else if (t.classList.contains("h-col")) {
      const col = Number(t.getAttribute("data-col"));
      drag = { kind: "col", col, x0: e.clientX, w0: log.columns[col]!.width, moved: false };
    } else if (t.classList.contains("hit") && bed) {
      drag = { kind: "move", bed, y0: e.clientY, to: log.beds.indexOf(bed), moved: false };
      return; // a plain press may still be a click or a scroll, so don't capture yet
    } else return;
    e.preventDefault();
    drawing.setPointerCapture(e.pointerId);
  });

  drawing.addEventListener("pointermove", (e) => {
    if (!drag) return;
    const z = ZOOMS[zoom]!;
    const ppm = pxPerMetre(log.scale);
    if (drag.kind === "top") {
      const d = drag;
      const step = snapStep(ppm * z);
      const raw = d.t0 + (d.y0 - e.clientY) / (ppm * z);
      const t = Math.max(step, +(Math.round(raw / step) * step).toFixed(4));
      d.moved ||= Math.abs(e.clientY - d.y0) > 2;
      if (d.moved && t !== d.bed.thickness) {
        d.bed.thickness = t;
        later(() => {
          // The drawing grows or shrinks at the bottom; scroll with it so the bed's base stays put.
          const before = drawing.scrollHeight;
          redraw();
          drawing.scrollTop += drawing.scrollHeight - before;
        });
      }
      if (d.moved) showTip(e, `Bed ${bedNo(d.bed)}: ${fmt(d.bed.thickness)}`);
    } else if (drag.kind === "grain") {
      const d = drag;
      const g = columnBox(log, "grain");
      if (!g) return;
      const phi = phiAt(at(e).x - g.x, g.width);
      const grain = d.bed.grain;
      // With no separate top grain size, the base end moves the whole edge.
      const whole = d.end === "base" && grain.top === undefined;
      if (d.end === "base") grain.base = phi;
      else grain.top = phi;
      if (grain.top !== undefined && grain.top === grain.base) delete grain.top;
      d.moved = true;
      later(redraw);
      showTip(e, `Bed ${bedNo(d.bed)}${whole ? "" : d.end === "base" ? " at base" : " at top"}: ${grainClassOf(phi).name.toLowerCase()}`);
    } else if (drag.kind === "col") {
      const d = drag;
      const col = log.columns[d.col]!;
      const mm = Math.round((d.w0 + (e.clientX - d.x0) / z) / MM);
      col.width = Math.round(Math.max(5, Math.min(200, mm)) * MM * 10) / 10;
      d.moved = true;
      later(redraw);
      showTip(e, `${col.title || "Column"}: ${Math.round(col.width / MM)} mm`);
    } else {
      const d = drag;
      if (!d.moved) {
        if (Math.abs(e.clientY - d.y0) < 6 || e.pointerType === "touch") return; // touch drags scroll
        d.moved = true;
        drawing.setPointerCapture(e.pointerId);
      }
      const { boxes, bodyHeight } = layoutBeds(log, pxPerMetre(log.scale));
      const y = at(e).y;
      // Insert where the pointer is: before the first bed whose middle is above it (beds are listed from the base).
      const k = boxes.filter((b) => (b.yTop + b.yBase) / 2 > y).length;
      d.to = k;
      const lineY = k === 0 ? (boxes[0]?.yBase ?? bodyHeight) : boxes[k - 1]!.yTop;
      drawing.querySelector(".drop-line")?.remove();
      const svg = drawing.querySelector("svg")!;
      svg.insertAdjacentHTML("beforeend", `<path class="drop-line" d="M0 ${lineY}H${svg.viewBox.baseVal.width}"/>`);
      showTip(e, `Move bed ${bedNo(d.bed)} here`);
    }
  });

  const endDrag = () => {
    const d = drag;
    drag = undefined;
    tip.hidden = true;
    if (!d?.moved) return;
    swallowClick = true;
    setTimeout(() => (swallowClick = false), 0);
    cancelAnimationFrame(frame);
    if (d.kind === "move") {
      drawing.querySelector(".drop-line")?.remove();
      const from = log.beds.indexOf(d.bed);
      const to = d.to > from ? d.to - 1 : d.to;
      if (to === from) return;
      log.beds.splice(from, 1);
      log.beds.splice(to, 0, d.bed);
    }
    if ("bed" in d) selected = d.bed.id;
    changed(null);
    showPanel();
  };
  drawing.addEventListener("pointerup", endDrag);
  drawing.addEventListener("pointercancel", () => {
    if (drag?.kind === "move") {
      // The browser took over (e.g. to scroll): leave the beds where they were.
      drag = undefined;
      tip.hidden = true;
      drawing.querySelector(".drop-line")?.remove();
    } else endDrag();
  });

  drawing.addEventListener("click", (e) => {
    if (swallowClick) return;
    const hit = (e.target as Element).closest("[data-bed]");
    select(hit ? (hit.getAttribute("data-bed") ?? undefined) : undefined);
  });

  drawing.addEventListener("keydown", (e) => {
    const i = log.beds.findIndex((b) => b.id === selected);
    if ((e.key === "ArrowUp" || e.key === "ArrowDown") && e.altKey) {
      e.preventDefault();
      if (i >= 0) moveBed(i, e.key === "ArrowUp" ? 1 : -1);
    } else if (e.key === "ArrowUp" || e.key === "ArrowDown") {
      e.preventDefault();
      const next = i < 0 ? (e.key === "ArrowUp" ? 0 : log.beds.length - 1) : i + (e.key === "ArrowUp" ? 1 : -1);
      const bed = log.beds[Math.max(0, Math.min(log.beds.length - 1, next))];
      if (bed) select(bed.id);
    } else if ((e.key === "Delete" || e.key === "Backspace") && i >= 0) {
      e.preventDefault();
      deleteBed(log.beds[i]!);
      drawing.focus();
    } else if (e.key === "Escape") select(undefined);
  });

  // Undo and redo work anywhere in the editor, except inside a text field, which has its own undo.
  const onKey = (e: KeyboardEvent) => {
    if (!editor.isConnected) return document.removeEventListener("keydown", onKey);
    if (!(e.ctrlKey || e.metaKey) || e.altKey || isTyping(e.target)) return;
    const k = e.key.toLowerCase();
    if (k === "z" && !e.shiftKey) undo();
    else if ((k === "z" && e.shiftKey) || k === "y") redo();
    else return;
    e.preventDefault();
  };
  document.addEventListener("keydown", onKey);

  // ---- side panel -------------------------------------------------------------------

  const tabs = h(
    "div",
    { class: "tabs", role: "tablist" },
    h("button", { role: "tab", id: "tab-bed", onclick: () => ((tab = "bed"), showPanel()) }, "Bed"),
    h("button", { role: "tab", id: "tab-layout", onclick: () => ((tab = "layout"), showPanel()) }, "Layout"),
  );
  const panel = h("aside", { class: "panel" }, tabs, panelBody);

  function showPanel(): void {
    for (const b of tabs.querySelectorAll("button")) b.setAttribute("aria-selected", String(b.id === `tab-${tab}`));
    if (tab === "layout") {
      panelBody.replaceChildren(h("h2", {}, "Layout and printing"), layoutFields(log, () => (syncBar(), changed())));
      return;
    }
    const bed = log.beds.find((b) => b.id === selected);
    if (bed) {
      const n = log.beds.indexOf(bed) + 1;
      panelBody.replaceChildren(
        h("h2", {}, `Bed ${n} `, h("span", { class: "meta" }, `of ${log.beds.length}, counted from the base`)),
        bedFields(log, bed, () => changed()),
        h(
          "div",
          { class: "panel-actions" },
          h("button", { onclick: () => addBed(log.beds.indexOf(bed) + 1) }, "Add bed above"),
          h("button", { onclick: () => addBed(log.beds.indexOf(bed)) }, "Add bed below"),
          h("button", { onclick: () => duplicateBed(bed) }, "Duplicate"),
          h("button", { class: "quiet danger", onclick: () => deleteBed(bed) }, "Delete bed"),
        ),
        h(
          "p",
          { class: "hint" },
          "Changes are saved automatically. On the drawing, drag a bed's top edge to change its thickness, the grain-size edge to change grain size, or the bed itself to move it. ↑ ↓ moves between beds, Alt+↑ ↓ moves the bed, Esc closes.",
        ),
      );
    } else {
      const total = log.beds.reduce((s, b) => s + b.thickness, 0);
      panelBody.replaceChildren(
        h("h2", {}, log.beds.length ? "Beds" : "First bed"),
        h(
          "p",
          { class: "hint" },
          log.beds.length
            ? `${plural(log.beds.length, "bed")}, ${+(log.units === "cm" ? total * 100 : total).toFixed(3)} ${log.units} in total. Click a bed in the log to edit it.`
            : "Start at the bottom of the section: the first bed you add is the lowest.",
        ),
        h("button", { class: "primary", onclick: () => addBed(log.beds.length) }, log.beds.length ? "Add bed on top" : "Add the first bed"),
      );
    }
  }

  function addBed(at: number): void {
    // A new bed starts as a copy of its neighbour's rock, which is usually closest to what comes next.
    const like = log.beds[at - 1] ?? log.beds[at];
    const bed: Bed = like
      ? { id: newId(), thickness: like.thickness, contact: "sharp", lithology: like.lithology.map((l) => ({ ...l })), grain: { ...like.grain } }
      : { id: newId(), thickness: 1, contact: "sharp", lithology: [{ pattern: "sandstone", percent: 100 }], grain: { base: 2.5 } };
    log.beds.splice(at, 0, bed);
    selected = bed.id;
    changed(null);
    select(bed.id);
    panelBody.querySelector<HTMLInputElement>("input[name=thickness]")?.select();
  }

  function duplicateBed(bed: Bed): void {
    const copy: Bed = { ...structuredClone(bed), id: newId() };
    delete copy.formation; // a copy continues the unit it sits in rather than starting it again
    delete copy.age;
    log.beds.splice(log.beds.indexOf(bed) + 1, 0, copy);
    selected = copy.id;
    changed(null);
    select(copy.id);
    toast(`Bed ${log.beds.indexOf(copy) + 1} is a copy of bed ${log.beds.indexOf(bed) + 1}.`);
  }

  function moveBed(i: number, by: number): void {
    const j = i + by;
    if (j < 0 || j >= log.beds.length) return;
    [log.beds[i], log.beds[j]] = [log.beds[j]!, log.beds[i]!];
    changed(null);
    select(log.beds[j]!.id);
  }

  function deleteBed(bed: Bed): void {
    const i = log.beds.indexOf(bed);
    log.beds.splice(i, 1);
    selected = log.beds[Math.min(i, log.beds.length - 1)]?.id;
    changed(null);
    select(selected);
    toast(`Bed ${i + 1} deleted.`, "info", { label: "Undo", run: undo });
  }

  function undo(): void {
    if (history.undo()) syncUndo();
  }
  function redo(): void {
    if (history.redo()) syncUndo();
  }

  // ---- top bar ----------------------------------------------------------------------

  const nameInput = h("input", {
    class: "log-name", value: log.name, maxlength: 200, "aria-label": "Log name",
    onchange: () => {
      if (!nameInput.value.trim()) {
        nameInput.value = log.name;
        return;
      }
      log.name = nameInput.value.trim();
      document.title = `${log.name} – GoLog`;
      changed(null);
    },
  });
  const units = h(
    "select",
    { name: "units", "aria-label": "Units", onchange: () => ((log.units = units.value as Log["units"]), showPanel(), changed(null)) },
    h("option", { value: "m" }, "metres"),
    h("option", { value: "cm" }, "centimetres"),
  );
  const scale = h(
    "select",
    { name: "scale", "aria-label": "Scale", onchange: () => ((log.scale = Number(scale.value)), showPanel(), changed(null)) },
    ...SCALES.map((s) => h("option", { value: s }, `1:${s}`)),
  );
  const undoBtn = h("button", { class: "quiet", name: "undo", "aria-label": "Undo", title: "Undo (Ctrl+Z)", onclick: undo }, "↶");
  const redoBtn = h("button", { class: "quiet", name: "redo", "aria-label": "Redo", title: "Redo (Ctrl+Shift+Z)", onclick: redo }, "↷");

  /** Puts the log's settings into the top bar, e.g. after undo. */
  function syncBar(): void {
    nameInput.value = log.name;
    document.title = `${log.name} – GoLog`;
    units.value = log.units;
    scale.value = String(log.scale);
  }
  function syncUndo(): void {
    undoBtn.disabled = !history.canUndo;
    redoBtn.disabled = !history.canRedo;
  }

  const setZoom = (z: number) => {
    zoom = Math.max(0, Math.min(ZOOMS.length - 1, z));
    viewSetting.set(zoomKey, String(ZOOMS[zoom]));
    redraw();
  };

  const editor = h(
    "div",
    { class: "editor" },
    h(
      "header",
      { class: "bar" },
      h("nav", { class: "crumbs" }, h("a", { href: "#/" }, "Projects"), h("a", { href: `#/p/${project.id}` }, project.name)),
      nameInput,
      h(
        "div",
        { class: "bar-tools" },
        h("div", { class: "undo", role: "group", "aria-label": "Undo and redo" }, undoBtn, redoBtn),
        h("label", { class: "inline" }, "Scale ", scale),
        h("label", { class: "inline" }, "Units ", units),
        h(
          "div",
          { class: "zoom", role: "group", "aria-label": "Zoom" },
          h("button", { class: "quiet", onclick: () => setZoom(zoom - 1), "aria-label": "Zoom out" }, "−"),
          zoomLabel,
          h("button", { class: "quiet", onclick: () => setZoom(zoom + 1), "aria-label": "Zoom in" }, "+"),
        ),
        pagesEl,
        h("button", { name: "stats", onclick: () => void openStats(log) }, "Statistics"),
        h("button", { name: "export", onclick: () => openExport(log) }, "Export & print…"),
        statusEl,
      ),
    ),
    h("div", { class: "workspace" }, drawing, panel),
    tip,
  );
  root.replaceChildren(editor);
  syncBar();
  syncUndo();
  redraw();
  showPanel();
  drawing.scrollTop = drawing.scrollHeight;

  // Returned to the router: true means "unsaved changes, ask before leaving".
  return () => {
    if (saver.pending) saver.save();
    return saver.pending;
  };
}
