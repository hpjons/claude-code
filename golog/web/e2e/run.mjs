// Browser test of GoLog: starts ../golog with an empty database, clicks through the app in
// headless Firefox and checks what was saved. Run `make e2e` (builds first) or `npm run e2e`.
// FIREFOX=/path/to/firefox overrides the browser; SHOTS=dir saves screenshots.
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import puppeteer from "puppeteer-core";

const here = dirname(fileURLToPath(import.meta.url));
const BASE = "http://127.0.0.1:7349";
const SHOTS = process.env.SHOTS;
const data = mkdtempSync(join(tmpdir(), "golog-e2e-"));
if (SHOTS) mkdirSync(SHOTS, { recursive: true });

const server = spawn(join(here, "..", "..", "golog"), ["-addr", "127.0.0.1:7349", "-data", data, "-no-browser"], { stdio: "ignore" });
const problems = [];
let browser;

const step = (s) => console.log("•", s);
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const api = async (path) => (await fetch(BASE + path)).json();

async function waitForServer() {
  for (let i = 0; i < 100; i++) {
    try {
      if ((await fetch(BASE + "/api/health")).ok) return;
    } catch {}
    await sleep(100);
  }
  throw new Error("golog did not start (run `make build` first)");
}

try {
  await waitForServer();
  browser = await puppeteer.launch({ browser: "firefox", executablePath: process.env.FIREFOX ?? "/usr/bin/firefox", headless: true });
  const page = await browser.newPage();
  await page.setViewport({ width: 1360, height: 900 });
  page.on("console", (m) => (m.type() === "error" || m.type() === "warn") && problems.push(`console ${m.type()}: ${m.text()}`));
  page.on("pageerror", (e) => problems.push(`page error: ${e.message}`));

  const shot = (name) => SHOTS && page.screenshot({ path: join(SHOTS, `${name}.png`) });
  const waitText = (sel, text) => page.waitForFunction((s, t) => document.querySelector(s)?.textContent?.includes(t), { timeout: 8000 }, sel, text);
  const saved = async () => {
    await sleep(50);
    await waitText(".save-status", "Saved");
  };
  const check = (ok, msg) => ok || problems.push(msg);
  const fill = async (selector, value) => {
    const el = await page.waitForSelector(selector);
    await el.click();
    await page.keyboard.down("Control");
    await page.keyboard.press("a");
    await page.keyboard.up("Control");
    await page.keyboard.press("Backspace");
    if (value !== "") await el.type(String(value));
  };
  const currentLog = () => page.evaluate(() => location.hash.split("/")[2]);

  step("create a project and a log");
  await page.goto(BASE + "/", { waitUntil: "networkidle0" });
  await page.type(".inline-form input", "Snæfellsnes field course");
  await page.click(".inline-form button");
  await waitText(".page-head h1", "Snæfellsnes");
  await page.type(".inline-form input", "Búðir section");
  await page.click(".inline-form button");
  await page.waitForSelector(".editor");

  step("layout: show bioturbation, palaeocurrent and facies, rename Notes, use A3");
  await page.click("#tab-layout");
  for (const name of ["Bioturbation", "Palaeocurrent", "Facies", "Formation"]) {
    const box = await page.$(`.column-list input[aria-label="Show ${name}"]`);
    if (!(await box.evaluate((b) => b.checked))) await box.click();
  }
  const notesTitle = await page.$$eval(".column-list .col-title", (els) => els.findIndex((e) => e.value === "Notes"));
  await fill(`.column-list li:nth-child(${notesTitle + 1}) .col-title`, "Description");
  await page.select("select[name=paper]", "A3");
  await saved();
  await shot("1-layout");

  step("bed 1: limestone with bivalves, starts a formation");
  await page.click("#tab-bed");
  await page.click(".panel button.primary");
  await fill("input[name=thickness]", 1.2);
  await page.select(".litho-row select", "limestone");
  await page.select("select[name=grain-base]", "silt");
  await fill("input[name=formation]", "Búðir Formation");
  await page.click("button[name=add-symbol]");
  await page.waitForSelector("dialog.picker[open]");
  await page.type("dialog.picker input[type=search]", "biv");
  await page.click('dialog.picker .pick[title="Bivalves"]');
  await fill("dialog.picker input[type=search]", "");
  await page.click('dialog.picker .pick[title="Gastropods"]');
  await shot("2-picker");
  await page.click("dialog.picker footer button");
  await page.waitForFunction(() => !document.querySelector("dialog.picker"));
  check((await page.$$(".chip")).length === 2, "expected two symbol chips after picking");

  step("bed 2: bioturbated mudstone");
  await page.click(".panel-actions button");
  await fill("input[name=thickness]", 0.6);
  await page.select(".litho-row select", "mudstone");
  await page.select("select[name=grain-base]", "clay");
  await page.click('input[name^="bi-"][value="4"]');

  step("bed 3: cross-bedded sandstone with palaeocurrents");
  await page.click(".panel-actions button");
  await fill("input[name=thickness]", 1.5);
  await page.select(".litho-row select", "sandstone");
  await page.select("select[name=grain-base]", "c");
  await page.select("select[name=grain-top]", "f");
  await page.click('.segmented input[value="erosional"]');
  await page.select("select[name=inBed]", "trough-cross-bedding");
  await fill("input[name=palaeocurrents]", "85, 120, 95");
  await fill("input[name=facies]", "St");
  await fill("input[name=palaeocurrents]", "85, north");
  const invalid = await page.$eval("input[name=palaeocurrents]", (i) => !i.validity.valid);
  check(invalid, "palaeocurrent 'north' should be marked invalid");
  await fill("input[name=palaeocurrents]", "85, 120, 95");

  step("bed 4: a lava flow starts a new formation");
  await page.click(".panel-actions button");
  await fill("input[name=thickness]", 2.4);
  await page.select(".litho-row select", "lava");
  await page.select("select[name=grain-base]", "");
  await page.select("select[name=inBed]", "");
  await fill("input[name=facies]", "");
  await fill("input[name=palaeocurrents]", "");
  await fill("input[name=formation]", "Staðarsveit lavas");
  await page.click("button[name=add-symbol]");
  await page.click('dialog.picker .pick[title="Columnar jointing"]');
  await page.click("dialog.picker footer button");
  await saved();
  await shot("3-log");

  step("check what was saved");
  let log = await api(`/api/logs/${await currentLog()}`);
  const [b1, b2, b3, b4] = log.beds;
  check(log.beds.length === 4, `expected 4 beds, got ${log.beds.length}`);
  check(JSON.stringify(b1.symbols) === '["bivalve","gastropod"]', `bed 1 symbols: ${JSON.stringify(b1.symbols)}`);
  check(b1.formation === "Búðir Formation", `bed 1 formation: ${b1.formation}`);
  check(b2.bioturbation === 4, `bed 2 bioturbation: ${b2.bioturbation}`);
  check(JSON.stringify(b3.palaeocurrents) === "[85,120,95]", `bed 3 palaeocurrents: ${JSON.stringify(b3.palaeocurrents)}`);
  check(b3.inBed === "trough-cross-bedding" && b3.facies === "St" && b3.contact === "erosional", `bed 3: ${JSON.stringify(b3)}`);
  check(b4.formation === "Staðarsveit lavas" && !b4.facies && !b4.palaeocurrents, `bed 4: ${JSON.stringify(b4)}`);
  check(log.paper === "A3", `paper: ${log.paper}`);
  check(log.columns.find((c) => c.kind === "notes").title === "Description", "notes column not renamed");
  check(log.columns.find((c) => c.kind === "bioturbation").visible, "bioturbation column not shown");

  step("drawing shows the new columns");
  const svg = await page.$eval(".logsvg", (s) => s.outerHTML);
  for (const [what, needle] of [["bivalve symbol", "#sym-bivalve"], ["in-bed pattern", "inbed-trough-cross-bedding"], ["formation name", "Búðir Formation"], ["renamed column", ">Description<"], ["arrow at 120°", "rotate(120)"]]) {
    check(svg.includes(needle), `drawing lacks the ${what}`);
  }

  step("colours: rocks have their usual colour, bed 1 gets its own, then black and white");
  const drawn = () => page.$eval(".logsvg", (s) => s.outerHTML);
  check((await drawn()).includes('fill="#c0a2ab"'), "lava not drawn in its colour");
  (await page.$$(".hit"))[0].click();
  await waitText(".panel h2", "Bed 1");
  const setColour = (v) =>
    page.$eval("input[name=colour-0]", (el, v) => {
      el.value = v;
      el.dispatchEvent(new Event("input", { bubbles: true }));
    }, v);
  await setColour("#c0392b");
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.beds[0].lithology[0].colour === "#c0392b", `bed 1 colour: ${log.beds[0].lithology[0].colour}`);
  check((await drawn()).includes('fill="#c0392b"'), "bed 1 not drawn in its own colour");
  check(await page.$eval("button[name=colour-reset-0]", (b) => !b.classList.contains("unused")), "no button to go back to the usual colour");
  await shot("3a-colour");
  await page.click("button[name=colour-reset-0]");
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.beds[0].lithology[0].colour === undefined, "usual colour not restored");
  await setColour("#c0392b");
  await page.click("#tab-layout");
  await page.select("select[name=fill]", "white");
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.fill === "white", `fill: ${log.fill}`);
  const bw = await drawn();
  check(!bw.includes("c0a2ab") && !bw.includes("c0392b"), "black-and-white log still has colours");
  await page.select("select[name=fill]", "colour");
  await saved();
  check((await drawn()).includes('fill="#c0392b"'), "bed 1's own colour lost after black and white");
  await page.click("#tab-bed");

  step("scale 1:10 makes the 5.7 m log run over two A3 pages (570 mm of section)");
  await page.select("select[name=scale]", "10");
  await saved();
  const pages = await page.$eval(".pages", (e) => e.textContent);
  check(pages.startsWith("2 pages"), `expected 2 pages at 1:10 on A3, got "${pages}"`);
  check((await page.$$(".logsvg .pagebreak")).length >= 1, "no page break drawn");
  await shot("3b-pages");
  await page.select("select[name=scale]", "50");

  step("move the Description column to the left of Structures & fossils");
  await page.click("#tab-layout");
  const idx = await page.$$eval(".column-list .col-title", (els) => els.findIndex((e) => e.value === "Description"));
  for (let k = 0; k < 4; k++) await page.click(`.column-list li:nth-child(${idx + 1 - k}) button[aria-label^="Move Description left"]`);
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  const order = log.columns.filter((c) => c.visible).map((c) => c.kind);
  check(order.indexOf("notes") < order.indexOf("symbols"), `column order: ${order.join(", ")}`);

  step("reload keeps everything, then select bed 3 by clicking");
  await page.reload({ waitUntil: "networkidle0" });
  await page.waitForSelector(".logsvg");
  (await page.$$(".hit"))[2].click();
  await waitText(".panel h2", "Bed 3");
  const pc = await page.$eval("input[name=palaeocurrents]", (i) => i.value);
  check(pc === "85, 120, 95", `palaeocurrent field after reload: ${pc}`);
  await shot("4-bed3");

  // ---- stage 3: editing on the drawing ----
  const beds = async () => (await api(`/api/logs/${await currentLog()}`)).beds;
  const center = async (sel) => {
    const r = await page.$eval(sel, (el) => {
      const b = el.getBoundingClientRect();
      return { x: b.x, y: b.y, w: b.width, h: b.height };
    });
    return r;
  };
  const dragFromTo = async (x0, y0, x1, y1) => {
    await page.mouse.move(x0, y0);
    await page.mouse.down();
    await page.mouse.move(x1, y1, { steps: 8 });
    await page.mouse.up();
    await saved();
  };
  const ids = (await beds()).map((b) => b.id);

  step("drag bed 3's top edge up by 0.5 m (1:50 at 100 %: 37.8 px)");
  let r = await center(`.h-top[data-bed="${ids[2]}"]`);
  await dragFromTo(r.x + 10, r.y + r.h / 2, r.x + 10, r.y + r.h / 2 - 37.8);
  check((await beds())[2].thickness === 2, `bed 3 thickness after drag: ${(await beds())[2].thickness}`);
  await shot("6-drag-top");

  step("undo and redo with the buttons and the keyboard");
  await page.click("button[name=undo]");
  await saved();
  check((await beds())[2].thickness === 1.5, "undo button did not put the thickness back");
  await page.click("button[name=redo]");
  await saved();
  check((await beds())[2].thickness === 2, "redo button did not redo");
  await page.focus(".drawing");
  await page.keyboard.down("Control");
  await page.keyboard.press("z");
  await page.keyboard.up("Control");
  await saved();
  check((await beds())[2].thickness === 1.5, "Ctrl+Z did not undo");
  await page.keyboard.down("Control");
  await page.keyboard.down("Shift");
  await page.keyboard.press("z");
  await page.keyboard.up("Shift");
  await page.keyboard.up("Control");
  await saved();
  check((await beds())[2].thickness === 2, "Ctrl+Shift+Z did not redo");

  step("drag bed 2's grain-size edge (clay) to fine sand");
  r = await center(`.h-grain[data-bed="${ids[1]}"]`);
  const fx = await page.$$eval(".logsvg text.tick", (els) => {
    const f = els.find((e) => e.textContent === "f").getBoundingClientRect();
    return f.x + f.width / 2;
  });
  await dragFromTo(r.x + r.w / 2, r.y + r.h * 0.75, fx, r.y + r.h * 0.75);
  let b2now = (await beds())[1];
  check(b2now.grain.base === 2.5 && b2now.grain.top === undefined, `bed 2 grain after drag: ${JSON.stringify(b2now.grain)}`);

  step("drag the upper half of bed 2's edge to medium sand: coarsening upward");
  const mx = await page.$$eval(".logsvg text.tick", (els) => {
    const m = els.find((e) => e.textContent === "m").getBoundingClientRect();
    return m.x + m.width / 2;
  });
  r = await center(`.h-grain[data-bed="${ids[1]}"]`);
  await dragFromTo(r.x + r.w / 2, r.y + r.h * 0.25, mx, r.y + r.h * 0.25);
  b2now = (await beds())[1];
  check(b2now.grain.base === 2.5 && b2now.grain.top === 1.5, `bed 2 grain after top drag: ${JSON.stringify(b2now.grain)}`);
  await shot("7-drag-grain");

  step("drag the Lithology column's right edge 10 mm wider");
  log = await api(`/api/logs/${await currentLog()}`);
  const litIdx = log.columns.findIndex((c) => c.kind === "lithology");
  const w0 = log.columns[litIdx].width;
  r = await center(`.h-col[data-col="${litIdx}"]`);
  await dragFromTo(r.x + r.w / 2, r.y + 40, r.x + r.w / 2 + 37.8, r.y + 40);
  log = await api(`/api/logs/${await currentLog()}`);
  const dmm = (log.columns[litIdx].width - w0) / (96 / 25.4);
  check(Math.abs(dmm - 10) <= 1, `lithology column grew by ${dmm.toFixed(1)} mm`);

  step("drag bed 1 to the top of the log, then undo");
  r = await center(`.hit[data-bed="${ids[0]}"]`);
  const top = await center(`.hit[data-bed="${ids[3]}"]`);
  await dragFromTo(r.x + 20, r.y + r.h / 2, r.x + 20, top.y + 4);
  check(JSON.stringify((await beds()).map((b) => b.id)) === JSON.stringify([ids[1], ids[2], ids[3], ids[0]]), "bed 1 did not move to the top");
  await page.click("button[name=undo]");
  await saved();
  check(JSON.stringify((await beds()).map((b) => b.id)) === JSON.stringify(ids), "undo did not put bed 1 back");

  step("keyboard: Alt+↑ moves the selected bed, Delete deletes it, Undo in the message brings it back");
  await page.click(`.hit[data-bed="${ids[0]}"]`);
  await waitText(".panel h2", "Bed 1");
  await page.focus(".drawing");
  await page.keyboard.down("Alt");
  await page.keyboard.press("ArrowUp");
  await page.keyboard.up("Alt");
  await saved();
  check((await beds())[1].id === ids[0], "Alt+↑ did not move bed 1 up");
  await waitText(".panel h2", "Bed 2");
  await page.keyboard.press("Delete");
  await saved();
  check((await beds()).length === 3, "Delete did not delete the bed");
  await page.click(".toast-action");
  await saved();
  check((await beds()).length === 4 && (await beds())[1].id === ids[0], "Undo in the message did not bring the bed back");

  step("duplicate a bed: the copy does not start the formation again");
  await page.click(`.hit[data-bed="${ids[0]}"]`);
  await page.$$eval(".panel-actions button", (bs) => bs.find((b) => b.textContent === "Duplicate").click());
  await saved();
  const all = await beds();
  check(all.length === 5 && all[2].lithology[0].pattern === "limestone" && !all[2].formation && all[1].formation === "Búðir Formation", `duplicate: ${JSON.stringify(all.map((b) => [b.lithology[0]?.pattern, b.formation]))}`);
  await shot("8-after-editing");

  // ---- stage 4: SedLog import ----
  // Small files in SedLog's format, written here (GoLog does not ship SedLog's own examples).
  const names = ["AGE", "FORMATION", "SCALE (CM)", "LITHOLOGY", "GRAIN SIZE", "STRUCTURES / FOSSILS", "NOTES", "BIOTURBATION", "PALAEOCURRENT", "FACIES", "OTHER1", "OTHER2", "OTHER3"];
  const layout = (title, used, show) => [
    "40.5", "1.0", "1.0", "0.37795275", title, "11.0", "11.0", "0", "1", "2", "0", "0", "0", String(used),
    ...names.flatMap((n, k) => [n, String(show[k]), k === 4 ? "400.0" : "200.0"]),
    "50.0", "380.0", "1.0", "0",
  ];
  const sedBed = (cm, contact, lith, grain, phi, symbols, notes, pc) => [
    cm, contact, lith, "1.0", "<none>", "0.0", "<none>", "0.0", grain, phi, grain, phi, "<none>", "0", String(symbols.length), ...symbols,
    notes, "<none>", "0", pc, "0", "", "0", "0", "", "0", "0", "", "0", "0", "", "", "0", "0", "0",
  ];
  const slg = ["SEDLOG-VERSION-2.1.1", "2",
    ...sedBed("150.0", "<none>", "Sandstone", "m", "1.5", ["Bivalves", "Unicorns"], "Cross-bedded", "45, 60"),
    ...sedBed("80.0", "Erosion", "Packstone", "c", "0.5", [], "Shelly", ""),
    ...layout("Hraunsvík", 4, [2, 3, 4, 8, 5, 6, 7, 8, 9, 10, 11, 12, 0])].join("\r\n") + "\r\n";
  const tem = ["SEDLOG-TEMPLATE-VERSION-2.1", ...layout("Untitled", 3, [3, 4, 6, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12])].join("\r\n") + "\r\n";
  const csv = "Thykkt (m);Berg;Lysing\n0.4;Siltstone;efst\n1.1;Sandstone;neðst\n";
  for (const [n, t] of [["hraunsvik.slg", slg], ["simple.tem", tem], ["own.csv", csv]]) writeFileSync(join(data, n), t);
  const choose = async (button, file) => {
    const [chooser] = await Promise.all([page.waitForFileChooser(), page.click(button)]);
    await chooser.accept([join(data, file)]);
  };

  step("import a SedLog log into the project, with a preview and warnings");
  await page.click(".crumbs a:nth-child(2)");
  await page.waitForSelector("button[name=import]");
  await choose("button[name=import]", "hraunsvik.slg");
  await page.waitForSelector("dialog.import[open] .import-preview svg");
  check((await page.$eval("input[name=import-name]", (i) => i.value)) === "Hraunsvík", "log name not taken from the file");
  const summary = await page.$eval(".import-summary", (e) => e.textContent);
  check(summary.includes("2 beds") && summary.includes("1:100"), `import summary: ${summary}`);
  const warns = await page.$eval(".import-warnings", (e) => e.textContent);
  check(warns.includes("Unicorns"), `warnings: ${warns}`);
  await shot("9-import");
  await page.click("button[name=import-ok]");
  await page.waitForSelector(".editor .logsvg");
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.name === "Hraunsvík" && log.beds.length === 2 && log.beds[0].thickness === 1.5 && log.beds[1].lithology[0].pattern === "packstone", `imported log: ${JSON.stringify(log.beds)}`);
  check(log.scale === 100 && log.palaeocurrent === "arrows", `imported settings: 1:${log.scale} ${log.palaeocurrent}`);
  const visible = log.columns.filter((c) => c.visible).map((c) => c.kind);
  check(JSON.stringify(visible) === JSON.stringify(["scale", "lithology", "grain", "palaeocurrent"]), `imported columns: ${visible}`);

  step("use a SedLog template's layout for this log, then undo it");
  await page.click("#tab-layout");
  await choose("button[name=load-layout]", "simple.tem");
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  const fromTem = log.columns.filter((c) => c.visible).map((c) => c.kind);
  check(JSON.stringify(fromTem) === JSON.stringify(["lithology", "grain", "notes"]) && log.beds.length === 2, `layout from template: ${fromTem}`);
  await page.click("button[name=undo]");
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.columns.find((c) => c.kind === "palaeocurrent").visible, "undo did not bring the old layout back");

  step("turn the grain-size scale above the log off and on again");
  const scaleText = () => page.$eval(".editor .logsvg", (svg) => [">vf<", ">gravel<", ">Grain size<"].map((t) => svg.innerHTML.includes(t)).join(","));
  check((await scaleText()) === "true,true,true", `grain-size scale before: ${await scaleText()}`);
  for (const name of ["grain-classes", "grain-groups", "grain-title"]) await page.click(`input[name=${name}]`);
  await saved();
  check((await scaleText()) === "false,false,false", `grain-size scale still drawn: ${await scaleText()}`);
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.hideGrainClasses && log.hideGrainGroups && log.hideGrainTitle, `grain-size scale settings not saved: ${log.hideGrainClasses} ${log.hideGrainGroups} ${log.hideGrainTitle}`);
  await page.reload();
  await page.waitForSelector(".editor .logsvg");
  await page.click("#tab-layout");
  check(!(await page.$eval("input[name=grain-classes]", (b) => b.checked)) && (await scaleText()) === "false,false,false", "grain-size scale back after reloading");
  await shot("6-no-grain-scale");
  for (const name of ["grain-classes", "grain-groups", "grain-title"]) await page.click(`input[name=${name}]`);
  await saved();
  log = await api(`/api/logs/${await currentLog()}`);
  check((await scaleText()) === "true,true,true" && !["hideGrainClasses", "hideGrainGroups", "hideGrainTitle"].some((k) => k in log), "grain-size scale not back on");

  step("import a CSV with headings GoLog doesn't know: choose the thickness column by hand");
  await page.click(".crumbs a:nth-child(2)");
  await page.waitForSelector("button[name=import]");
  await choose("button[name=import]", "own.csv");
  await page.waitForSelector("dialog.import[open] select[name=map-thickness]");
  check(await page.$eval("button[name=import-ok]", (b) => b.disabled), "Import should wait until a thickness column is chosen");
  await page.select("select[name=map-thickness]", "0");
  await page.select("select[name=map-lith1]", "1");
  await page.select("select[name=map-notes]", "2");
  await page.select("select[name=csv-units]", "m");
  await page.click("input[name=csv-top-first]");
  await page.waitForFunction(() => !document.querySelector("button[name=import-ok]").disabled && document.querySelector(".import-summary").textContent.includes("2 beds"));
  await shot("10-import-csv");
  await page.click("button[name=import-ok]");
  await page.waitForSelector(".editor .logsvg");
  log = await api(`/api/logs/${await currentLog()}`);
  check(log.name === "own" && log.beds[0].notes === "neðst" && log.beds[0].thickness === 1.1 && log.beds[1].lithology[0].pattern === "siltstone", `CSV import: ${JSON.stringify(log.beds)}`);

  // ---- stage 5: output ----
  step("export the Búðir section as PDF, SVG, PNG, CSV and a GoLog file");
  await page.click(".crumbs a:nth-child(2)");
  await page.waitForSelector(".rows a");
  await page.$$eval(".rows a", (as) => as.find((a) => a.textContent === "Búðir section").click());
  await page.waitForSelector(".editor .logsvg");
  const budir = await api(`/api/logs/${await currentLog()}`);
  await page.evaluate(() => {
    // Keep the files in the page instead of letting the browser save them.
    window.__files = {};
    window.addEventListener("golog-download", async (e) => {
      const bytes = new Uint8Array(await e.detail.blob.arrayBuffer());
      let bin = "";
      for (const b of bytes) bin += String.fromCharCode(b);
      const ext = e.detail.name.endsWith(".sedlog.csv") ? "sedlog.csv" : e.detail.name.split(".").pop();
      window.__files[ext] = { name: e.detail.name, b64: btoa(bin) };
    });
    const click = HTMLAnchorElement.prototype.click;
    HTMLAnchorElement.prototype.click = function () {
      if (!this.download) click.call(this);
    };
    window.print = () => (window.__printed = document.querySelectorAll("#print-root .print-sheet").length);
  });
  await page.click("button[name=export]");
  await page.waitForSelector("dialog.export[open]");
  const note = await page.$eval("dialog.export .hint", (e) => e.textContent);
  check(note.includes("page") && note.includes("A3"), `export note: ${note}`);
  await shot("11-export");
  const files = {};
  for (const [button, ext] of [["export-pdf", "pdf"], ["export-svg", "svg"], ["export-png", "png"], ["export-csv", "csv"], ["export-csv-sedlog", "sedlog.csv"], ["export-golog", "golog"]]) {
    await page.click(`button[name=${button}]`);
    await page.waitForFunction((x) => window.__files?.[x], { timeout: 30000 }, ext);
    const f = await page.evaluate((x) => window.__files[x], ext);
    files[ext] = Buffer.from(f.b64, "base64");
    check(f.name === `Búðir section.${ext}`, `file name: ${f.name}`);
  }
  const pdf = files.pdf.toString("latin1");
  check(pdf.startsWith("%PDF-"), "the PDF is not a PDF");
  check((pdf.match(/\/Type \/Page\b/g) ?? []).length >= 1, "the PDF has no pages");
  check(files.svg.toString().includes("Búðir Formation") && files.svg.toString().includes(">Key<"), "the SVG lacks the formation or the key");
  check(!/class="(hit|h-top|h-grain|h-col|knob|pagebreak)/.test(files.svg.toString()), "the SVG carries the editor's handles");
  const svgError = await page.evaluate((t) => new DOMParser().parseFromString(t, "image/svg+xml").querySelector("parsererror")?.textContent ?? "", files.svg.toString());
  check(!svgError, `the SVG is not valid XML: ${svgError.slice(0, 200)}`);
  check(files.pdf.length > 20000, `the PDF looks empty (${files.pdf.length} bytes)`);
  check(files.png.subarray(1, 4).toString() === "PNG" && files.png.length > 20000, `the PNG is not a picture (${files.png.length} bytes)`);
  const brightness = await page.evaluate(async (b64) => {
    const bytes = Uint8Array.from(atob(b64), (c) => c.charCodeAt(0));
    const img = await createImageBitmap(new Blob([bytes], { type: "image/png" }));
    const c = new OffscreenCanvas(60, 60);
    const g = c.getContext("2d");
    g.drawImage(img, 0, 0, 60, 60);
    const d = g.getImageData(0, 0, 60, 60).data;
    let sum = 0;
    for (let i = 0; i < d.length; i += 4) sum += (d[i] + d[i + 1] + d[i + 2]) / 3;
    return sum / (d.length / 4);
  }, files.png.toString("base64"));
  check(brightness > 170, `the PNG is mostly dark (average brightness ${brightness.toFixed(0)} of 255)`);
  const csvRows = files.csv.toString("utf8").replace(/^\uFEFF/, "").trim().split("\r\n");
  check(csvRows.length === budir.beds.length + 1 && csvRows[0].startsWith("THICKNESS (CM),BASE BOUNDARY") && csvRows.some((r) => r.includes("Búðir Formation")), `CSV: ${csvRows.slice(0, 2).join(" | ")}`);
  check(files.csv[0] === 0xef && files["sedlog.csv"].toString().startsWith("THICKNESS (CM),"), "CSV byte-order marks: spreadsheet copy needs one, SedLog copy must not have one");
  if (SHOTS) writeFileSync(join(SHOTS, "export.pdf"), files.pdf);

  step("print: one sheet per page");
  await page.click("button[name=export-print]");
  const sheets = await page.evaluate(() => window.__printed);
  check(sheets >= 1, "nothing was laid out for printing");
  // What a printer would get: Firefox prints the page with its print styles to PDF.
  const printed = Buffer.from(await page.pdf({ preferCSSPageSize: true, printBackground: true }));
  check(printed.subarray(0, 5).toString() === "%PDF-", "printing made no PDF");
  const printedFile = join(data, "printed.pdf");
  writeFileSync(printedFile, printed);
  if (SHOTS) writeFileSync(join(SHOTS, "printed.pdf"), printed);
  try {
    // Firefox compresses its PDFs, so pdfinfo (poppler) reads the page count and size if it is installed.
    const info = execFileSync("pdfinfo", [printedFile], { encoding: "utf8" });
    check(new RegExp(`Pages:\\s+${sheets}\\b`).test(info), `printed pages differ from the ${sheets} laid out:\n${info}`);
    check(/Page size:\s+84[12](\.\d+)? x 119[01]/.test(info), `printed pages are not A3:\n${info}`);
  } catch (e) {
    if (e.code !== "ENOENT") throw e;
  }
  await page.evaluate(() => window.dispatchEvent(new Event("afterprint")));
  check(!(await page.$("#print-root")), "the print pages were not removed after printing");

  step("the GoLog file and the CSV come back in as the same beds");
  writeFileSync(join(data, "budir.golog"), files.golog);
  writeFileSync(join(data, "budir.csv"), files.csv);
  for (const [file, what] of [["budir.golog", "GoLog file"], ["budir.csv", "CSV"]]) {
    await page.click(".crumbs a:nth-child(2)");
    await page.waitForSelector("button[name=import]");
    await choose("button[name=import]", file);
    await page.waitForSelector("dialog.import[open] .import-preview svg");
    await page.click("button[name=import-ok]");
    await page.waitForSelector(".editor .logsvg");
    const back = await api(`/api/logs/${await currentLog()}`);
    // A CSV has no column for a bed's own colour, and a bed with only a top grain size comes back with it as the base.
    const lith = (b) => (what === "CSV" ? b.lithology.map(({ colour: _c, ...l }) => l) : b.lithology);
    const grain = (g) => (what === "CSV" && g.base === undefined && g.top !== undefined ? { base: g.top } : g);
    const sig = (l) => JSON.stringify(l.beds.map((b) => [b.thickness, b.contact, lith(b), grain(b.grain), b.symbols ?? [], b.inBed ?? "", b.palaeocurrents ?? [], b.formation ?? "", b.notes ?? ""]));
    check(sig(back) === sig(budir), `${what} round trip changed the beds:\n${sig(back)}\n${sig(budir)}`);
    if (what === "GoLog file") check(back.paper === "A3" && back.scale === budir.scale && JSON.stringify(back.columns) === JSON.stringify(budir.columns), "GoLog file lost the layout");
  }

  step("statistics: SedLog's example1 facies and thicknesses give m = 0.50 and r = 0.89, as in SedLog");
  const ex1 = [[1, "1"], [0.45, "2"], [0.45, "3"], [1.35, "4"], [0.65, "4"], [0.5, "4"], [0.6, "2"], [0.35, "5"], [0.65, "2"]];
  const ordered = {
    golog: 1, kind: "log", name: "Facies test",
    content: { ...budir, beds: ex1.map(([thickness, facies], i) => ({ id: `s${i}`, thickness, contact: "sharp", lithology: [{ pattern: i % 2 ? "shale" : "limestone", percent: 100 }], grain: { base: 3.5, top: i % 3 ? 3.5 : 6 }, facies })) },
  };
  for (const k of ["id", "projectId", "name", "revision", "created", "updated"]) delete ordered.content[k];
  writeFileSync(join(data, "facies.golog"), JSON.stringify(ordered));
  await page.click(".crumbs a:nth-child(2)");
  await page.waitForSelector("button[name=import]");
  await choose("button[name=import]", "facies.golog");
  await page.waitForSelector("dialog.import[open] .import-preview svg");
  await page.click("button[name=import-ok]");
  await page.waitForSelector(".editor .logsvg");
  await page.click("button[name=stats]");
  await page.waitForSelector("dialog.stats[open] .stats-verdict");
  const verdicts = await page.$$eval("dialog.stats .stats-verdict", (ps) => ps.map((p) => p.textContent));
  check(verdicts.length === 2 && verdicts[0].startsWith("m = 0.50") && verdicts[1].startsWith("r = 0.89"), `statistics: ${verdicts.join(" | ")}`);
  const cells = await page.$$eval("dialog.stats .stats-matrix text.val", (ts) => ts.map((t) => t.textContent).join(" "));
  check(cells === "1.00 .50 .50 1.00 1.00 1.00", `transition matrix: ${cells}`);
  const lith = await page.$$eval("dialog.stats .stats-share:first-child tbody th", (ths) => ths.map((t) => t.textContent));
  check(lith.join(",") === "Limestone,Shale", `lithology table: ${lith}`);
  await shot("12-statistics");
  await page.$eval("#st-fac", (e) => e.scrollIntoView());
  await shot("13-statistics-facies");
  await page.$eval("dialog.stats .stats-body", (b) => (b.scrollTop = b.scrollHeight));
  await shot("14-statistics-runs");
  await page.click("button[name=stats-csv]");
  await page.waitForFunction(() => window.__files.csv?.name.endsWith(".statistics.csv"));
  const statsCsv = Buffer.from(await page.evaluate(() => window.__files.csv.b64), "base64").toString("utf8");
  check(statsCsv.includes("Markov order metric m,0.5\r\n") && statsCsv.includes("Runs order metric r,0.8889"), `statistics CSV:\n${statsCsv}`);
  await page.keyboard.press("Escape");
  await page.waitForFunction(() => !document.querySelector("dialog.stats"));

  step("statistics with only one facies explain why there is no facies test");
  await page.click(".crumbs a:nth-child(2)");
  await page.waitForSelector(".rows a");
  await page.$$eval(".rows a", (as) => as.find((a) => a.textContent === "Búðir section").click());
  await page.waitForSelector(".editor .logsvg");
  await page.click("button[name=stats]");
  await page.waitForSelector("dialog.stats[open] .stats-section");
  const notes = await page.$eval("dialog.stats .stats-body", (e) => e.textContent);
  check(notes.includes("at least two different facies") && !notes.includes("Are the facies ordered") && notes.includes("Do beds thicken"), `no-facies statistics: ${notes.slice(0, 300)}`);
  await page.keyboard.press("Escape");

  step("phone width");
  await page.setViewport({ width: 390, height: 844 });
  await shot("5-phone");
  const overflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth + 1);
  check(!overflow, "page scrolls sideways at phone width");
  await page.setViewport({ width: 1360, height: 900 });

  step("the user guide opens from the Projects page, with its pictures");
  await page.goto(BASE + "/", { waitUntil: "networkidle0" });
  const guideHref = await page.$eval("a.guide-link", (a) => a.getAttribute("href"));
  await page.goto(BASE + guideHref, { waitUntil: "networkidle0" });
  const pictures = await page.$$eval("main img", (is) => is.map((i) => i.naturalWidth));
  check(pictures.length >= 5 && pictures.every((w) => w > 0), `guide pictures: ${pictures}`);
  const anchors = await page.$$eval("nav a[href^='#']", (as) => as.filter((a) => !document.querySelector(a.getAttribute("href"))).map((a) => a.href));
  check(anchors.length === 0, `guide links to missing sections: ${anchors}`);

  step("the licence and credits open from the Projects page");
  await page.goto(BASE + "/", { waitUntil: "networkidle0" });
  const noticesHref = await page.$eval("a.licences", (a) => a.getAttribute("href"));
  const notices = await fetch(BASE + noticesHref);
  const noticesText = await notices.text();
  check(notices.ok && /charset=utf-8/i.test(notices.headers.get("content-type") ?? ""), `licences: ${notices.status} ${notices.headers.get("content-type")}`);
  check(["MIT License", "jspdf", "svg2pdf.js", "modernc.org/sqlite", "Go standard library"].every((t) => noticesText.includes(t)), "licences file is missing an entry");

  step("Stop GoLog ends the program");
  await page.goto(BASE + "/", { waitUntil: "networkidle0" });
  check((await page.$eval(".app-foot", (e) => e.textContent)).includes("GoLog "), "no version on the Projects page");
  const exited = new Promise((r) => server.once("exit", r));
  await page.click("button[name=stop]");
  await page.waitForFunction(() => document.title === "GoLog has stopped");
  check(await Promise.race([exited.then(() => true), sleep(5000).then(() => false)]), "GoLog still running after Stop GoLog");
} catch (e) {
  problems.push(`stopped: ${e.message}`);
} finally {
  await browser?.close();
  server.kill();
  rmSync(data, { recursive: true, force: true });
}

console.log(problems.length ? "\nPROBLEMS:\n- " + problems.join("\n- ") : "\nALL STEPS OK");
process.exit(problems.length ? 1 : 0);
