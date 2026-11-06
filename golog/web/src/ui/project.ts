import { api } from "../api";
import { ago, h, plural, toast } from "./dom";
import { importIntoProject } from "./importdialog";

/** One project: its logs, a form for a new log, rename and delete. */
export async function projectView(root: HTMLElement, id: string): Promise<void> {
  const project = await api.project(id);
  document.title = `${project.name} – GoLog`;

  const input = h("input", { name: "name", required: true, maxlength: 200, placeholder: "e.g. Section A, Kleifarvatn", "aria-label": "Log name" });
  const form = h(
    "form",
    {
      class: "inline-form",
      onsubmit: async (e: Event) => {
        e.preventDefault();
        try {
          const l = await api.createLog(project.id, input.value);
          location.hash = `#/l/${l.id}`;
        } catch (err) {
          toast((err as Error).message, "error");
        }
      },
    },
    input,
    h("button", { class: "primary" }, "New log"),
    h("button", { type: "button", name: "import", onclick: () => importIntoProject(project.id) }, "Import file…"),
  );

  const rename = async () => {
    const name = prompt("Project name", project.name);
    if (name === null || name.trim() === project.name) return;
    try {
      await api.renameProject(project.id, name);
      await projectView(root, id);
    } catch (err) {
      toast((err as Error).message, "error");
    }
  };

  const remove = async () => {
    const what = project.logs.length ? ` and its ${plural(project.logs.length, "log")}` : "";
    if (!confirm(`Delete the project “${project.name}”${what}? This can't be undone.`)) return;
    try {
      await api.deleteProject(project.id);
      location.hash = "#/";
    } catch (err) {
      toast((err as Error).message, "error");
    }
  };

  const deleteLog = async (logId: string, name: string) => {
    if (!confirm(`Delete the log “${name}”? This can't be undone.`)) return;
    try {
      await api.deleteLog(logId);
      await projectView(root, id);
    } catch (err) {
      toast((err as Error).message, "error");
    }
  };

  const list =
    project.logs.length === 0
      ? h("p", { class: "empty-note" }, "No logs in this project yet. Name the first section you measured, or import a file: a GoLog file, or a SedLog log, template or CSV table.")
      : h(
          "ul",
          { class: "rows" },
          ...project.logs.map((l) =>
            h(
              "li",
              {},
              h("a", { href: `#/l/${l.id}`, class: "row-main" }, l.name),
              h("span", { class: "meta" }, `${plural(l.bedCount, "bed")}, changed ${ago(l.updated)}`),
              h("button", { class: "quiet small", onclick: () => deleteLog(l.id, l.name), "aria-label": `Delete log ${l.name}` }, "Delete"),
            ),
          ),
        );

  root.replaceChildren(
    h(
      "main",
      { class: "page" },
      h("nav", { class: "crumbs" }, h("a", { href: "#/" }, "Projects")),
      h(
        "header",
        { class: "page-head" },
        h("h1", {}, project.name),
        h("div", { class: "actions" }, h("button", { class: "quiet", onclick: rename }, "Rename"), h("button", { class: "quiet danger", onclick: remove }, "Delete project")),
      ),
      form,
      list,
    ),
  );
  if (project.logs.length === 0) input.focus();
}
