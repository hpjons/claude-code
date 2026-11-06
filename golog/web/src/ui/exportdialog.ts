import { download, fileName, printLog, toCSV, toGoLog, toPDF, toPNG, toSVG } from "../export/formats";
import { layoutPages } from "../export/pages";
import type { Log } from "../types";
import { h, plural, toast } from "./dom";

const KEY_SETTING = "golog.key";

function keyWanted(): boolean {
  try {
    return localStorage.getItem(KEY_SETTING) !== "no";
  } catch {
    return true;
  }
}

/** Saving the log as PDF, SVG, PNG, CSV or a GoLog file, and printing it. */
export function openExport(log: Log): void {
  const key = h("input", {
    type: "checkbox", name: "with-key", checked: keyWanted(),
    onchange: () => {
      try {
        localStorage.setItem(KEY_SETTING, key.checked ? "yes" : "no");
      } catch {
        /* still works for this time */
      }
      describe();
    },
  });
  const pagesNote = h("p", { class: "hint", role: "status" });
  const describe = () => {
    const { svgs, shrink } = layoutPages(log, key.checked);
    pagesNote.textContent =
      `${plural(svgs.length, "page")} on ${log.paper} at 1:${log.scale}.` +
      (shrink < 1 ? ` The columns are wider than the paper, so the pages are shrunk to ${Math.floor(shrink * 100)} % and the scale is no longer exact.` : "");
    pagesNote.classList.toggle("warn", shrink < 1);
  };

  const run = (label: string, make: () => Blob | Promise<Blob>, ext: string) => async (e: Event) => {
    const button = e.currentTarget as HTMLButtonElement;
    button.disabled = true;
    const text = button.firstChild!.textContent;
    button.firstChild!.textContent = "Making…";
    try {
      download(fileName(log, ext), await make());
      toast(`${label} saved as “${fileName(log, ext)}”.`);
    } catch (err) {
      toast(`Could not make the ${label}: ${(err as Error).message}`, "error");
    } finally {
      button.disabled = false;
      button.firstChild!.textContent = text;
    }
  };

  const option = (name: string, title: string, what: string, onclick: (e: Event) => void) =>
    h("button", { type: "button", class: "export-option", name, onclick }, h("strong", {}, title), h("span", {}, what));

  const dialog = h(
    "dialog",
    { class: "picker export", "aria-label": "Export and print" },
    h("header", {}, h("h2", {}, "Export and print")),
    h(
      "div",
      { class: "export-body" },
      h("label", { class: "inline" }, key, " Add a key to the patterns and symbols used"),
      pagesNote,
      h(
        "div",
        { class: "export-grid" },
        option("export-pdf", "PDF", "Pages on the chosen paper, to print or send", run("PDF", () => toPDF(log, key.checked), "pdf")),
        option("export-print", "Print…", "Straight to a printer, one page per sheet", () => {
          dialog.close();
          printLog(log, key.checked);
        }),
        option("export-svg", "SVG drawing", "The whole log as one drawing, for Inkscape or Illustrator", run("SVG drawing", () => toSVG(log, key.checked), "svg")),
        option("export-png", "PNG picture", "The whole log as a picture, for reports and slides", run("PNG picture", () => toPNG(log, key.checked), "png")),
        option("export-csv", "CSV table", "The beds as a table, for Excel and other spreadsheets", run("CSV table", () => toCSV(log), "csv")),
        option("export-csv-sedlog", "CSV for SedLog", "The same table for SedLog's Import Log from CSV file", run("CSV for SedLog", () => toCSV(log, true), "sedlog.csv")),
        option("export-golog", "GoLog file", "Everything in this log, to back up or give to someone with GoLog", run("GoLog file", () => toGoLog(log), "golog")),
      ),
    ),
    h("footer", {}, h("span", { class: "hint" }, "Paper size and columns are set in the Layout tab."), h("button", { type: "button", onclick: () => dialog.close() }, "Close")),
  );
  dialog.addEventListener("close", () => dialog.remove());
  document.body.append(dialog);
  describe();
  dialog.showModal();
}
