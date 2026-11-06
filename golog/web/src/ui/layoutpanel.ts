import { MARGIN_MM, PAPER_MM, pageInfo } from "../render/log";
import type { Log, Paper } from "../types";
import { api } from "../api";
import { h, plural, toast } from "./dom";
import { pickFile } from "./importdialog";

const MM = 96 / 25.4;

/** Paper, palaeocurrent style and the columns: which are shown, their order, titles and widths. */
export function layoutFields(log: Log, onChange: () => void): HTMLElement {
  const fit = h("p", { class: "hint", role: "status" });
  const updateFit = () => {
    const info = pageInfo(log);
    const over = (info.width - info.pageWidth) / MM;
    fit.textContent =
      `${plural(info.pages, "page")} on ${log.paper} at 1:${log.scale}. ` +
      (info.fitsWidth
        ? `The columns fit the page width (${Math.round(info.width / MM)} of ${Math.round(info.pageWidth / MM)} mm).`
        : `The columns are ${Math.ceil(over)} mm too wide for the page: hide or narrow a column, or use A3.`);
    fit.classList.toggle("warn", !info.fitsWidth);
  };
  const changed = () => {
    updateFit();
    onChange();
  };

  const paper = h(
    "select",
    { name: "paper", onchange: () => ((log.paper = paper.value as Paper), changed()) },
    ...(Object.keys(PAPER_MM) as Paper[]).map((p) => h("option", { value: p, selected: log.paper === p }, `${p} portrait (${PAPER_MM[p][0]} × ${PAPER_MM[p][1]} mm)`)),
  );
  const fill = h(
    "select",
    { name: "fill", onchange: () => ((log.fill = fill.value as Log["fill"]), changed()) },
    h("option", { value: "colour", selected: log.fill === "colour" }, "Colour"),
    h("option", { value: "white", selected: log.fill === "white" }, "Black and white"),
  );
  const arrows = h(
    "select",
    { name: "palaeocurrent", onchange: () => ((log.palaeocurrent = arrows.value as Log["palaeocurrent"]), changed()) },
    h("option", { value: "arrows", selected: log.palaeocurrent === "arrows" }, "Arrows"),
    h("option", { value: "numbers", selected: log.palaeocurrent === "numbers" }, "Numbers"),
  );

  const scaleBox = (key: "hideGrainClasses" | "hideGrainGroups" | "hideGrainTitle", name: string, label: string) =>
    h(
      "label",
      { class: "inline" },
      h("input", {
        type: "checkbox", name, checked: !log[key],
        onchange: (e: Event) => {
          if ((e.target as HTMLInputElement).checked) delete log[key];
          else log[key] = true;
          changed();
        },
      }),
      ` ${label}`,
    );

  const list = h("ol", { class: "column-list" });
  const render = () => {
    list.replaceChildren(
      ...log.columns.map((col, i) => {
        const width = h("input", {
          type: "number", min: 5, max: 200, step: 1, value: Math.round(col.width / MM), class: "mm", "aria-label": `${col.title} width in mm`,
          oninput: () => {
            const mm = Number(width.value);
            if (!width.validity.badInput && mm >= 5 && mm <= 200) {
              col.width = Math.round(mm * MM * 10) / 10;
              changed();
            }
          },
        });
        const move = (d: number) => {
          const j = i + d;
          if (j < 0 || j >= log.columns.length) return;
          [log.columns[i], log.columns[j]] = [log.columns[j]!, log.columns[i]!];
          render();
          changed();
        };
        return h(
          "li",
          { class: col.visible ? "" : "hidden-col" },
          h("input", { type: "checkbox", checked: col.visible, "aria-label": `Show ${col.title}`, onchange: (e: Event) => ((col.visible = (e.target as HTMLInputElement).checked), render(), changed()) }),
          h("input", { class: "col-title", value: col.title, maxlength: 60, "aria-label": "Column title", oninput: (e: Event) => ((col.title = (e.target as HTMLInputElement).value), changed()) }),
          width,
          h("span", { class: "unit" }, "mm"),
          h("button", { type: "button", class: "quiet small", "aria-label": `Move ${col.title} left`, disabled: i === 0, onclick: () => move(-1) }, "↑"),
          h("button", { type: "button", class: "quiet small", "aria-label": `Move ${col.title} right`, disabled: i === log.columns.length - 1, onclick: () => move(1) }, "↓"),
        );
      }),
    );
  };
  render();
  updateFit();

  return h(
    "div",
    { class: "fields" },
    h("label", { class: "field" }, h("span", {}, "Paper"), paper),
    fit,
    h(
      "label",
      { class: "field" },
      h("span", {}, "Lithologies"),
      fill,
      h("small", { class: "hint" }, "Each rock has a usual colour. Click the swatch next to a lithology to give a bed its own colour."),
    ),
    h("label", { class: "field" }, h("span", {}, "Palaeocurrents shown as"), arrows),
    h(
      "fieldset",
      { class: "field" },
      h("legend", {}, "Columns"),
      h("p", { class: "hint" }, "Ticked columns are drawn, top of the list on the left. Widths are on paper."),
      list,
    ),
    h(
      "fieldset",
      { class: "field" },
      h("legend", {}, "Grain-size scale"),
      h("p", { class: "hint" }, "Shown above the grain-size column."),
      scaleBox("hideGrainClasses", "grain-classes", "Grain sizes (clay, silt, vf … boul)"),
      scaleBox("hideGrainGroups", "grain-groups", "Groups (mud, sand, gravel)"),
      scaleBox("hideGrainTitle", "grain-title", "Column title"),
    ),
    h("p", { class: "hint" }, `Margins are ${MARGIN_MM} mm. Long logs continue on the next page; the dashed blue lines show where.`),
    h(
      "div",
      { class: "field" },
      h("button", { type: "button", name: "load-layout", onclick: loadLayout }, "Use a SedLog layout…"),
      h("small", { class: "hint" }, "Takes the columns, scale and units from a SedLog template (.tem) or log (.slg); the beds stay as they are."),
    ),
  );

  async function loadLayout(): Promise<void> {
    const file = await pickFile(".tem,.slg,.golog");
    if (!file) return;
    try {
      const res = await api.importFile(file.data);
      if (res.format === "csv") throw new Error("a CSV file has no layout; choose a .tem, .slg or GoLog file");
      const c = res.content;
      Object.assign(log, { columns: c.columns, scale: c.scale, units: c.units, palaeocurrent: c.palaeocurrent });
      arrows.value = log.palaeocurrent;
      render();
      changed();
      toast(`Layout from ${file.name} applied. Undo puts the old one back.`);
    } catch (err) {
      toast(`Could not use ${file.name}: ${(err as Error).message}`, "error");
    }
  }
}
