import "./style.css";
import { homeView } from "./ui/home";
import { logView } from "./ui/logview";
import { projectView } from "./ui/project";
import { h } from "./ui/dom";

const root = document.getElementById("app")!;

/** Set by a view that may have unsaved changes; returns true if leaving now would lose them. */
let unsaved: () => boolean = () => false;

async function route(): Promise<void> {
  const [, kind, id] = location.hash.match(/^#\/([pl])\/([\w-]+)/) ?? [];
  unsaved = () => false;
  try {
    if (kind === "p") await projectView(root, id!);
    else if (kind === "l") unsaved = await logView(root, id!);
    else await homeView(root);
  } catch (err) {
    root.replaceChildren(
      h("main", { class: "page" }, h("h1", {}, "Something went wrong"), h("p", {}, (err as Error).message), h("p", {}, h("a", { href: "#/" }, "Back to projects"))),
    );
  }
}

window.addEventListener("hashchange", route);
window.addEventListener("beforeunload", (e) => {
  if (unsaved()) e.preventDefault();
});
route();
