import { api, type CSVOptions, type ImportResult } from "../api";
import { drawLog } from "../render/log";
import type { Log } from "../types";
import { h, plural, toast } from "./dom";

/** Lets the person pick a file and returns its bytes as base64 (what the import API takes). */
export function pickFile(accept: string): Promise<{ name: string; data: string } | null> {
  return new Promise((resolve) => {
    const input = h("input", { type: "file", accept, hidden: true });
    input.addEventListener("change", async () => {
      const file = input.files?.[0];
      input.remove();
      if (!file) return resolve(null);
      if (file.size > 6_000_000) {
        toast("That file is too large to be a SedLog file.", "error");
        return resolve(null);
      }
      const bytes = new Uint8Array(await file.arrayBuffer());
      let bin = "";
      for (let i = 0; i < bytes.length; i += 0x8000) bin += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
      resolve({ name: file.name, data: btoa(bin) });
    });
    input.addEventListener("cancel", () => (input.remove(), resolve(null)));
    document.body.append(input);
    input.click();
  });
}

const FORMATS: Record<ImportResult["format"], string> = { slg: "SedLog log", tem: "SedLog template", csv: "CSV file", golog: "GoLog file" };

/** Imports a SedLog log, template or CSV file into a project, after showing what it will look like. */
export async function importIntoProject(projectId: string): Promise<void> {
  const file = await pickFile(".slg,.tem,.csv,.txt,.golog,.json");
  if (!file) return;
  let res: ImportResult;
  try {
    res = await api.importFile(file.data);
  } catch (err) {
    toast(`Could not import ${file.name}: ${(err as Error).message}`, "error");
    return;
  }

  const name = h("input", { name: "import-name", required: true, maxlength: 200, value: res.name || file.name.replace(/\.[^.]+$/, ""), "aria-label": "Log name" });
  const summary = h("p", { class: "import-summary" });
  const warnings = h("div", { class: "import-warnings" });
  const preview = h("div", { class: "import-preview", "aria-label": "Preview" });
  const mapping = h("div", { class: "import-mapping" });
  const ok = h("button", { class: "primary", name: "import-ok" }, "Import");

  const show = () => {
    const c = res.content;
    const total = c.beds.reduce((s, b) => s + b.thickness, 0);
    summary.textContent =
      `${FORMATS[res.format]}: ` +
      (res.format === "tem" ? "a layout without beds; the new log starts empty with these columns and settings." : `${plural(c.beds.length, "bed")}, ${+(c.units === "cm" ? total * 100 : total).toFixed(2)} ${c.units} at 1:${c.scale}.`);
    warnings.replaceChildren(
      ...(res.warnings.length
        ? [h("h3", {}, "Not brought over exactly"), h("ul", {}, ...res.warnings.map((w) => h("li", {}, w)))]
        : [h("p", { class: "hint" }, "Everything in the file has a place in GoLog.")]),
    );
    const log: Log = { ...c, id: "preview", projectId, name: name.value || "Preview", revision: 0, created: "", updated: "" };
    preview.innerHTML = drawLog(log, { zoom: 0.6 });
    ok.disabled = res.format === "csv" && c.beds.length === 0;
  };

  if (res.csv) {
    const info = res.csv;
    const opts: CSVOptions = { mapping: { ...info.mapping }, units: info.units, topFirst: info.topFirst };
    const reparse = async () => {
      try {
        res = await api.importFile(file.data, opts);
        show();
      } catch (err) {
        toast((err as Error).message, "error");
      }
    };
    const example = (col: number) => info.sample.map((r) => r[col]).find((v) => v && v.trim()) ?? "";
    const columnSelect = (key: string, label: string) =>
      h(
        "label",
        { class: "map-row" },
        h("span", {}, label),
        h(
          "select",
          {
            name: `map-${key}`,
            onchange: (e: Event) => {
              const v = Number((e.target as HTMLSelectElement).value);
              if (v < 0) delete opts.mapping[key];
              else opts.mapping[key] = v;
              reparse();
            },
          },
          h("option", { value: -1, selected: opts.mapping[key] === undefined }, "—"),
          ...info.headers.map((hd, i) => {
            const ex = example(i);
            return h("option", { value: i, selected: opts.mapping[key] === i }, `${hd || `Column ${i + 1}`}${ex ? ` (e.g. ${ex.slice(0, 24)})` : ""}`);
          }),
        ),
      );
    mapping.append(
      h("h3", {}, "Columns"),
      h("p", { class: "hint" }, `${plural(info.rows, "row")} in the file. GoLog matched the columns by their headings; change any that are wrong.`),
      h(
        "div",
        { class: "map-options" },
        h(
          "label",
          { class: "inline" },
          "Thickness in ",
          h(
            "select",
            { name: "csv-units", onchange: (e: Event) => ((opts.units = (e.target as HTMLSelectElement).value as "m" | "cm"), reparse()) },
            h("option", { value: "cm", selected: opts.units === "cm" }, "centimetres"),
            h("option", { value: "m", selected: opts.units === "m" }, "metres"),
          ),
        ),
        h(
          "label",
          { class: "inline" },
          h("input", { type: "checkbox", name: "csv-top-first", checked: opts.topFirst, onchange: (e: Event) => ((opts.topFirst = (e.target as HTMLInputElement).checked), reparse()) }),
          " The first row is the top bed",
        ),
      ),
      h("div", { class: "map-grid" }, ...info.fields.map((f) => columnSelect(f.key, f.label))),
    );
  }

  const dialog = h(
    "dialog",
    { class: "picker import", "aria-label": `Import ${file.name}` },
    h("header", {}, h("h2", {}, `Import ${file.name}`)),
    h(
      "div",
      { class: "import-body" },
      h("label", { class: "field" }, h("span", {}, "Name of the new log"), name),
      summary,
      mapping,
      warnings,
      preview,
    ),
    h(
      "footer",
      {},
      h("button", { type: "button", onclick: () => dialog.close() }, "Cancel"),
      ok,
    ),
  );
  name.addEventListener("input", show);
  ok.addEventListener("click", async () => {
    if (!name.value.trim()) return name.focus();
    ok.disabled = true;
    try {
      const log = await api.createLog(projectId, name.value, res.content);
      dialog.close();
      location.hash = `#/l/${log.id}`;
    } catch (err) {
      ok.disabled = false;
      toast((err as Error).message, "error");
    }
  });
  dialog.addEventListener("close", () => dialog.remove());
  document.body.append(dialog);
  show();
  dialog.showModal();
}
