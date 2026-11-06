// Tiny helpers for building the interface without a framework.

type Child = Node | string | null | undefined | false;
type Attrs = Record<string, string | number | boolean | EventListener | undefined>;

/** Creates an element: h("button", { class: "x", onclick: fn }, "Label"). */
export function h<K extends keyof HTMLElementTagNameMap>(tag: K, attrs: Attrs = {}, ...children: Child[]): HTMLElementTagNameMap[K] {
  const el = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (v === undefined || v === false) continue;
    if (k.startsWith("on") && typeof v === "function") el.addEventListener(k.slice(2), v);
    else if (v === true) el.setAttribute(k, "");
    else el.setAttribute(k, String(v));
  }
  for (const c of children) if (c !== null && c !== undefined && c !== false) el.append(c);
  return el;
}

/** "3 minutes ago", "yesterday" … */
export function ago(iso: string): string {
  const s = (Date.parse(iso) - Date.now()) / 1000;
  const rtf = new Intl.RelativeTimeFormat("en", { numeric: "auto" });
  const steps: [Intl.RelativeTimeFormatUnit, number][] = [["second", 60], ["minute", 60], ["hour", 24], ["day", 30], ["month", 12], ["year", Infinity]];
  let v = s;
  for (const [unit, size] of steps) {
    if (Math.abs(v) < size) return rtf.format(Math.round(v), unit);
    v /= size;
  }
  return "";
}

/** Shows a short message at the bottom of the window, optionally with one action button. */
export function toast(msg: string, kind: "info" | "error" = "info", action?: { label: string; run: () => void }): void {
  for (const old of document.querySelectorAll(".toast")) old.remove();
  const t = h(
    "div",
    { class: `toast ${kind}`, role: kind === "error" ? "alert" : "status" },
    msg,
    action && h("button", { class: "toast-action", onclick: () => (t.remove(), action.run()) }, action.label),
  );
  document.body.append(t);
  setTimeout(() => t.remove(), kind === "error" || action ? 7000 : 3000);
}

export const plural = (n: number, word: string) => `${n} ${word}${n === 1 ? "" : "s"}`;
