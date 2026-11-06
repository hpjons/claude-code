import { api } from "../api";
import { ago, h, plural, toast } from "./dom";

/** Start screen: the list of projects and a form for a new one. */
export async function homeView(root: HTMLElement): Promise<void> {
  document.title = "GoLog";
  const [projects, health] = await Promise.all([api.projects(), api.health().catch(() => null)]);

  const input = h("input", { name: "name", required: true, maxlength: 200, placeholder: "e.g. Reykjanes field course 2026", "aria-label": "Project name" });
  const form = h(
    "form",
    {
      class: "inline-form",
      onsubmit: async (e: Event) => {
        e.preventDefault();
        try {
          const p = await api.createProject(input.value);
          location.hash = `#/p/${p.id}`;
        } catch (err) {
          toast((err as Error).message, "error");
        }
      },
    },
    input,
    h("button", { class: "primary" }, "New project"),
  );

  const list =
    projects.length === 0
      ? h("p", { class: "empty-note" }, "No projects yet. A project holds the logs from one field area or exercise.")
      : h(
          "ul",
          { class: "rows" },
          ...projects.map((p) =>
            h(
              "li",
              {},
              h("a", { href: `#/p/${p.id}`, class: "row-main" }, p.name),
              h("span", { class: "meta" }, `${plural(p.logCount, "log")}, changed ${ago(p.updated)}`),
            ),
          ),
        );

  const stopBtn =
    health?.canStop &&
    h(
      "button",
      {
        type: "button", class: "quiet", name: "stop",
        onclick: async () => {
          try {
            await api.stop();
          } catch (err) {
            toast((err as Error).message, "error");
            return;
          }
          document.title = "GoLog has stopped";
          root.replaceChildren(
            h(
              "main",
              { class: "page stopped" },
              h("h1", {}, "GoLog has stopped"),
              h("p", {}, "Your work is saved. You can close this tab."),
              h("p", { class: "hint" }, "To carry on, start GoLog again: it opens a new tab."),
            ),
          );
        },
      },
      "Stop GoLog",
    );

  root.replaceChildren(
    h(
      "main",
      { class: "page" },
      h("header", { class: "page-head" }, h("h1", {}, "Projects"), h("a", { href: "/guide.html", target: "_blank", class: "guide-link" }, "User guide")),
      form,
      list,
      h(
        "footer",
        { class: "app-foot" },
        h(
          "span",
          { class: "hint" },
          health ? `GoLog ${health.version}` : "GoLog",
          " ",
          h("a", { href: "/notices.txt", target: "_blank", class: "licences" }, "MIT licence and credits"),
        ),
        stopBtn,
      ),
    ),
  );
  if (projects.length === 0) input.focus();
}
