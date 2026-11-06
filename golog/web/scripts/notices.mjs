// Writes web/public/notices.txt: GoLog's licence followed by the licence of every library built
// into the program. The lists come from the build itself (the Go modules linked on each release
// platform, and the npm packages in the web bundle), so nothing is missed or listed needlessly.
// Run `make notices` after changing dependencies, and commit the result.
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync, readdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { build } from "vite";

const web = join(dirname(fileURLToPath(import.meta.url)), "..");
const root = join(web, "..");
const PLATFORMS = ["linux/amd64", "linux/arm64", "windows/amd64", "darwin/arm64", "darwin/amd64"];
const LICENCE_FILE = /^(licen[cs]e|copying|notice|patents|authors)/i;

const licenceFiles = (dir) =>
  readdirSync(dir)
    .filter((f) => LICENCE_FILE.test(f) && !/logo/i.test(f))
    .sort()
    .map((f) => ({ name: f, text: readFileSync(join(dir, f), "utf8").replace(/\r\n/g, "\n").trim() }));

const go = (args, env = {}) => execFileSync("go", args, { cwd: root, encoding: "utf8", env: { ...process.env, ...env } }).trim();

// Go modules linked into the program on any release platform.
const goModules = new Set();
for (const p of PLATFORMS) {
  const [GOOS, GOARCH] = p.split("/");
  const out = go(["list", "-deps", "-f", "{{if not .Standard}}{{with .Module}}{{if not .Main}}{{.Path}}{{end}}{{end}}{{end}}", "."], { GOOS, GOARCH, CGO_ENABLED: "0" });
  for (const m of out.split("\n")) if (m) goModules.add(m);
}
const goEntries = [...goModules].sort().map((m) => {
  const [dir, version] = go(["list", "-m", "-f", "{{.Dir}} {{.Version}}", m]).split(" ");
  return { name: m, version, files: licenceFiles(dir) };
});

// The Go standard library is in every Go program. Its licence sits in GOROOT, or elsewhere when Go
// came from a Linux distribution's package.
const goroot = go(["env", "GOROOT"]);
const stdDir = [goroot, "/usr/share/licenses/golang", "/usr/share/doc/golang"].find((d) => existsSync(join(d, "LICENSE")));
if (!stdDir) throw new Error("can't find the Go licence (LICENSE in GOROOT)");
goEntries.unshift({ name: "Go standard library", version: go(["env", "GOVERSION"]).replace(/[- ].*/, ""), files: licenceFiles(stdDir).filter((f) => /^(license|patents)$/i.test(f.name)) });

// npm packages that end up in the web bundle.
const out = await build({ root: web, logLevel: "silent", build: { write: false } });
const packages = new Set();
for (const o of [out].flat())
  for (const chunk of o.output)
    for (const id of chunk.moduleIds ?? []) {
      const m = id.match(/.*node_modules\/((?:@[^/]+\/)?[^/]+)/);
      if (m) packages.add(m[1]);
    }
const npmEntries = [...packages].sort().map((p) => {
  const dir = join(web, "node_modules", p);
  const pkg = JSON.parse(readFileSync(join(dir, "package.json"), "utf8"));
  return { name: p, version: pkg.version, licence: pkg.license, files: licenceFiles(dir) };
});

const missing = [...goEntries, ...npmEntries].filter((e) => e.files.length === 0).map((e) => e.name);
if (missing.length) throw new Error(`no licence file found for: ${missing.join(", ")}`);

const rule = "=".repeat(78);
const section = (e) =>
  [rule, `${e.name} ${e.version}${e.licence ? `  (${e.licence})` : ""}`, rule, "", ...e.files.flatMap((f) => (e.files.length > 1 ? [`--- ${f.name} ---`, "", f.text, ""] : [f.text, ""]))].join("\n");

const text = [
  "GoLog licences",
  "",
  "GoLog is released under the MIT licence:",
  "",
  readFileSync(join(root, "LICENSE"), "utf8").trim(),
  "",
  "GoLog is built with the libraries listed below. Each keeps its own licence, given",
  "here in full. DOMPurify offers a choice of MPL-2.0 or Apache-2.0; GoLog uses it",
  "under Apache-2.0.",
  "",
  "Built into the program (Go):",
  ...goEntries.map((e) => `  ${e.name} ${e.version}`),
  "",
  "Built into the web app (JavaScript):",
  ...npmEntries.map((e) => `  ${e.name} ${e.version}`),
  "",
  ...goEntries.map(section),
  ...npmEntries.map(section),
].join("\n");

writeFileSync(join(web, "public", "notices.txt"), text.trimEnd() + "\n");
console.log(`notices.txt: ${goEntries.length} Go modules, ${npmEntries.length} npm packages`);
