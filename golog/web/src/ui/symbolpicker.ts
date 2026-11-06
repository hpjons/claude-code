import { SYMBOLS, SYMBOL_GROUPS, symbolIcon } from "../render/symbols";
import { h } from "./dom";

/**
 * A dialog for choosing symbols: a searchable grid, grouped. Clicking a symbol toggles it;
 * `onToggle` is called each time. The dialog removes itself when closed.
 */
export function openSymbolPicker(chosen: Set<string>, onToggle: (id: string, on: boolean) => void): void {
  const grid = h("div", { class: "picker-groups" });
  const search = h("input", { type: "search", placeholder: "Search symbols", "aria-label": "Search symbols" });

  const render = () => {
    const q = search.value.trim().toLowerCase();
    grid.replaceChildren(
      ...SYMBOL_GROUPS.map((g) => {
        const items = SYMBOLS.filter((s) => s.group === g && (!q || s.name.toLowerCase().includes(q)));
        if (!items.length) return "";
        return h(
          "section",
          {},
          h("h3", {}, g),
          h(
            "div",
            { class: "picker-grid" },
            ...items.map((s) => {
              const b = h("button", { type: "button", class: "pick", "aria-pressed": chosen.has(s.id) ? "true" : "false", title: s.name });
              b.innerHTML = symbolIcon(s.id, 28);
              b.append(h("span", {}, s.name));
              b.addEventListener("click", () => {
                const on = !chosen.has(s.id);
                if (on) chosen.add(s.id);
                else chosen.delete(s.id);
                b.setAttribute("aria-pressed", String(on));
                onToggle(s.id, on);
              });
              return b;
            }),
          ),
        );
      }),
    );
    if (!grid.childElementCount || [...grid.children].every((c) => !c.textContent)) grid.replaceChildren(h("p", { class: "hint" }, `No symbol matches “${search.value}”.`));
  };
  search.addEventListener("input", render);
  render();

  const dialog = h(
    "dialog",
    { class: "picker", "aria-label": "Structures and fossils" },
    h("header", {}, h("h2", {}, "Structures & fossils"), search),
    grid,
    h("footer", {}, h("p", { class: "hint" }, "Click to add or remove. Symbols are drawn beside the bed in this order."), h("button", { class: "primary", type: "button", onclick: () => dialog.close() }, "Done")),
  );
  dialog.addEventListener("close", () => dialog.remove());
  dialog.addEventListener("click", (e) => {
    if (e.target === dialog) dialog.close(); // click on the backdrop
  });
  document.body.append(dialog);
  dialog.showModal();
  search.focus();
}
