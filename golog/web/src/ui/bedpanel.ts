import { GRAIN_CLASSES, grainClassOf } from "../grain";
import { PATTERNS, PATTERN_BY_ID, lithColour, patternSwatch } from "../render/patterns";
import { SYMBOLS, SYMBOL_BY_ID, symbolIcon } from "../render/symbols";
import type { Bed, ColumnKind, Contact, Log } from "../types";
import { h } from "./dom";
import { openSymbolPicker } from "./symbolpicker";

const CONTACTS: [Contact, string][] = [
  ["sharp", "Sharp"],
  ["erosional", "Erosional"],
  ["gradational", "Gradational"],
  ["unknown", "Unknown"],
];

/** Parses "85, 120 95" into directions; null if any part isn't a direction 0–360. */
export function parseDirections(text: string): number[] | null {
  const parts = text.split(/[\s,;]+/).filter(Boolean);
  const vals = parts.map(Number);
  return vals.every((v) => Number.isFinite(v) && v >= 0 && v <= 360) ? vals : null;
}

/** Spreads 100 % over the lithologies: the first keeps the remainder. */
export function rebalance(bed: Bed): void {
  const n = bed.lithology.length;
  if (n === 0) return;
  const share = n === 1 ? 100 : Math.floor(100 / n);
  bed.lithology.forEach((l, i) => (l.percent = i === 0 ? 100 - share * (n - 1) : share));
}

/** The form for one bed. Every edit changes `bed` and calls `onChange`. */
export function bedFields(log: Log, bed: Bed, onChange: () => void): HTMLElement {
  const toUnit = (m: number) => +(log.units === "cm" ? m * 100 : m).toFixed(4);
  const fromUnit = (v: number) => (log.units === "cm" ? v / 100 : v);
  const visible = (kind: ColumnKind) => log.columns.some((c) => c.kind === kind && c.visible);

  const thickness = h("input", {
    name: "thickness", type: "number", min: 0, step: "any", inputmode: "decimal", required: true,
    value: toUnit(bed.thickness),
    oninput: () => {
      const v = fromUnit(thickness.value.trim() === "" ? NaN : Number(thickness.value));
      if (!thickness.validity.badInput && v > 0 && v <= 10000) {
        bed.thickness = v;
        thickness.setCustomValidity("");
        onChange();
      } else thickness.setCustomValidity("Thickness must be more than 0");
    },
  });

  const contact = h(
    "div",
    { class: "segmented four", role: "radiogroup", "aria-label": "Contact at the base of the bed" },
    ...CONTACTS.map(([value, label]) =>
      h(
        "label",
        {},
        h("input", { type: "radio", name: `contact-${bed.id}`, value, checked: bed.contact === value, onchange: () => ((bed.contact = value), onChange()) }),
        h("span", {}, label),
      ),
    ),
  );

  const text = (label: string, key: "facies" | "formation" | "age" | "notes", hint?: string, multiline = false) => {
    const set = (v: string) => {
      if (v.trim()) bed[key] = v;
      else delete bed[key];
      onChange();
    };
    const input = multiline
      ? h("textarea", { rows: 3, maxlength: 5000, name: key, oninput: (e: Event) => set((e.target as HTMLTextAreaElement).value) }, bed[key] ?? "")
      : h("input", { name: key, maxlength: 200, value: bed[key] ?? "", oninput: (e: Event) => set((e.target as HTMLInputElement).value) });
    return h("label", { class: "field" }, h("span", {}, label), input, hint ? h("small", { class: "hint" }, hint) : null);
  };

  const fields: (HTMLElement | null)[] = [
    h("label", { class: "field" }, h("span", {}, `Thickness (${log.units})`), thickness),
    h("fieldset", { class: "field" }, h("legend", {}, "Lithology"), lithologyRows(log, bed, onChange)),
    h("fieldset", { class: "field" }, h("legend", {}, "Grain size"), h("div", { class: "pair" }, grainSelect(bed, "base", onChange), grainSelect(bed, "top", onChange))),
    h("fieldset", { class: "field" }, h("legend", {}, "Base contact"), contact),
    inBedSelect(bed, onChange),
  ];
  if (visible("symbols") || bed.symbols?.length) fields.push(symbolChips(bed, onChange));
  if (visible("bioturbation") || bed.bioturbation !== undefined) fields.push(bioturbation(bed, onChange));
  if (visible("palaeocurrent") || bed.palaeocurrents?.length) fields.push(palaeocurrents(bed, onChange));
  if (visible("facies") || bed.facies) fields.push(text("Facies", "facies"));
  if (visible("formation") || bed.formation) fields.push(text("Formation starts here", "formation", "Leave empty if the bed belongs to the formation below."));
  if (visible("age") || bed.age) fields.push(text("Age unit starts here", "age", "Leave empty if the bed belongs to the unit below."));
  fields.push(text("Notes", "notes", undefined, true));
  return h("div", { class: "fields" }, ...fields);
}

function lithologyRows(log: Log, bed: Bed, onChange: () => void): HTMLElement {
  const box = h("div", { class: "litho-rows" });
  const render = () => {
    box.replaceChildren(
      ...bed.lithology.map((l, i) => {
        // In a colour log the swatch is also the colour picker for this part of the bed.
        const colourful = log.fill === "colour";
        const art = h("span", { class: "swatch-art" });
        const showSwatch = () => (art.innerHTML = patternSwatch(l.pattern, 32, lithColour(l.pattern, log.fill, l.colour)));
        showSwatch();
        const picker = colourful
          ? h("input", {
              type: "color", class: "colour-input", name: `colour-${i}`, value: lithColour(l.pattern, "colour", l.colour),
              "aria-label": `Colour of lithology ${i + 1}`,
              oninput: (e: Event) => {
                const v = (e.target as HTMLInputElement).value.toLowerCase();
                if (v === PATTERN_BY_ID.get(l.pattern)?.colour) delete l.colour;
                else l.colour = v;
                showSwatch();
                reset.classList.toggle("unused", !l.colour);
                onChange();
              },
            })
          : null;
        const swatch = h("label", { class: "swatch-holder", title: colourful ? "Change the colour" : undefined }, art, picker);
        // Kept in place (just invisible) when unused, so the rows stay lined up.
        const reset = h("button", {
          type: "button", class: `quiet small${l.colour ? "" : " unused"}`, name: `colour-reset-${i}`, title: "Back to the usual colour", "aria-label": `Usual colour for lithology ${i + 1}`,
          onclick: () => (delete l.colour, render(), onChange()),
        }, "↺");
        const select = h(
          "select",
          {
            "aria-label": `Lithology ${i + 1}`,
            onchange: (e: Event) => {
              l.pattern = (e.target as HTMLSelectElement).value;
              render();
              onChange();
            },
          },
          ...(["Clastic", "Carbonate", "Volcanic", "Other"] as const).map((g) =>
            h("optgroup", { label: g }, ...PATTERNS.filter((p) => p.group === g).map((p) => h("option", { value: p.id, selected: p.id === l.pattern }, p.name))),
          ),
        );
        const many = bed.lithology.length > 1;
        const percent = many
          ? h("input", {
              type: "number", min: 1, max: 100, step: 1, value: l.percent, class: "percent", "aria-label": `Lithology ${i + 1} percent`,
              oninput: (e: Event) => {
                const input = e.target as HTMLInputElement;
                const v = input.value.trim() === "" ? NaN : Number(input.value);
                if (!input.validity.badInput && v > 0 && v <= 100) {
                  l.percent = v;
                  onChange();
                }
              },
            })
          : null;
        const remove = many
          ? h("button", { type: "button", class: "quiet small", "aria-label": `Remove lithology ${i + 1}`, onclick: () => (bed.lithology.splice(i, 1), rebalance(bed), render(), onChange()) }, "×")
          : null;
        return h("div", { class: "litho-row" }, swatch, colourful && reset, select, percent, percent && h("span", { class: "unit" }, "%"), remove);
      }),
    );
    if (bed.lithology.length < 3) {
      box.append(
        h("button", {
          type: "button", class: "quiet small add",
          onclick: () => {
            const used = new Set(bed.lithology.map((l) => l.pattern));
            bed.lithology.push({ pattern: PATTERNS.find((p) => !used.has(p.id))!.id, percent: 0 });
            rebalance(bed);
            render();
            onChange();
          },
        }, bed.lithology.length ? "Add another lithology" : "Add lithology"),
      );
    }
  };
  render();
  return box;
}

function grainSelect(bed: Bed, end: "base" | "top", onChange: () => void): HTMLElement {
  const current = bed.grain[end];
  return h(
    "label",
    {},
    h("span", { class: "sub" }, end === "base" ? "At base" : "At top"),
    h(
      "select",
      {
        name: `grain-${end}`,
        onchange: (e: Event) => {
          const v = (e.target as HTMLSelectElement).value;
          if (v === "") delete bed.grain[end];
          else bed.grain[end] = GRAIN_CLASSES.find((c) => c.key === v)!.phi;
          onChange();
        },
      },
      h("option", { value: "", selected: current === undefined }, end === "base" ? "Not recorded" : "Same as base"),
      ...GRAIN_CLASSES.map((c) => h("option", { value: c.key, selected: current !== undefined && grainClassOf(current).key === c.key }, c.name)),
    ),
  );
}

function inBedSelect(bed: Bed, onChange: () => void): HTMLElement {
  return h(
    "label",
    { class: "field" },
    h("span", {}, "Structure inside the bed"),
    h(
      "select",
      {
        name: "inBed",
        onchange: (e: Event) => {
          const v = (e.target as HTMLSelectElement).value;
          if (v) bed.inBed = v;
          else delete bed.inBed;
          onChange();
        },
      },
      h("option", { value: "", selected: !bed.inBed }, "None (show the lithology pattern)"),
      ...SYMBOLS.filter((s) => s.inBed).map((s) => h("option", { value: s.id, selected: bed.inBed === s.id }, s.name)),
    ),
    h("small", { class: "hint" }, "Drawn repeated inside the grain-size profile, in place of the lithology pattern."),
  );
}

function symbolChips(bed: Bed, onChange: () => void): HTMLElement {
  const list = h("div", { class: "chips" });
  const render = () => {
    const ids = bed.symbols ?? [];
    list.replaceChildren(
      ...ids.map((id, i) => {
        const chip = h("span", { class: "chip" });
        chip.innerHTML = symbolIcon(id, 18);
        chip.append(
          SYMBOL_BY_ID.get(id)?.name ?? id,
          h("button", {
            type: "button", class: "chip-x", "aria-label": `Remove ${SYMBOL_BY_ID.get(id)?.name ?? id}`,
            onclick: () => {
              ids.splice(i, 1);
              if (!ids.length) delete bed.symbols;
              render();
              onChange();
            },
          }, "×"),
        );
        return chip;
      }),
      h("button", {
        type: "button", class: "quiet small", name: "add-symbol",
        onclick: () =>
          openSymbolPicker(new Set(bed.symbols ?? []), (id, on) => {
            const cur = bed.symbols ?? [];
            bed.symbols = on ? [...cur, id] : cur.filter((s) => s !== id);
            if (!bed.symbols.length) delete bed.symbols;
            render();
            onChange();
          }),
      }, ids.length ? "Add or remove…" : "Add symbols…"),
    );
  };
  render();
  return h("fieldset", { class: "field" }, h("legend", {}, "Structures & fossils"), list);
}

function bioturbation(bed: Bed, onChange: () => void): HTMLElement {
  const options: [string, string][] = [["", "–"], ...[0, 1, 2, 3, 4, 5, 6].map((n): [string, string] => [String(n), String(n)])];
  return h(
    "fieldset",
    { class: "field" },
    h("legend", {}, "Bioturbation index"),
    h(
      "div",
      { class: "segmented", role: "radiogroup", "aria-label": "Bioturbation index, 0 none to 6 complete" },
      ...options.map(([value, label]) =>
        h(
          "label",
          {},
          h("input", {
            type: "radio", name: `bi-${bed.id}`, value, checked: (bed.bioturbation === undefined ? "" : String(bed.bioturbation)) === value,
            onchange: () => {
              if (value === "") delete bed.bioturbation;
              else bed.bioturbation = Number(value);
              onChange();
            },
          }),
          h("span", {}, label),
        ),
      ),
    ),
    h("small", { class: "hint" }, "0 = no bioturbation, 6 = completely reworked; – = not recorded."),
  );
}

function palaeocurrents(bed: Bed, onChange: () => void): HTMLElement {
  const input = h("input", {
    name: "palaeocurrents", inputmode: "decimal", placeholder: "e.g. 85, 120, 95",
    value: (bed.palaeocurrents ?? []).join(", "),
    oninput: () => {
      const vals = parseDirections(input.value);
      if (vals === null) {
        input.setCustomValidity("Directions are numbers from 0 to 360, separated by commas");
        return;
      }
      input.setCustomValidity("");
      if (vals.length) bed.palaeocurrents = vals;
      else delete bed.palaeocurrents;
      onChange();
    },
  });
  return h("label", { class: "field" }, h("span", {}, "Palaeocurrents (° from north)"), input);
}
