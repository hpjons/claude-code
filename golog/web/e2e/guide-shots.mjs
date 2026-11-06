// Takes the pictures in the user guide (web/public/guide/) from a sample project.
// Run `make guide-shots` after the interface changes, then check the pictures and commit them.
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import puppeteer from "puppeteer-core";

const here = dirname(fileURLToPath(import.meta.url));
const out = join(here, "..", "public", "guide");
const BASE = "http://127.0.0.1:7352";
const data = mkdtempSync(join(tmpdir(), "golog-guide-"));
const server = spawn(join(here, "..", "..", "golog"), ["-addr", "127.0.0.1:7352", "-data", data, "-no-browser"], { stdio: "ignore" });
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const call = async (method, path, body) => (await fetch(BASE + path, { method, headers: { "Content-Type": "application/json" }, body: body && JSON.stringify(body) })).json();

// A made-up section of shallow-marine to fluvial beds capped by a lava, from the base up:
// thickness (m), contact, rock, grain base/top (phi), in-bed structure, symbols, facies, notes.
const BEDS = [
  [0.8, "sharp", "mudstone", 6, 6, "", ["bivalve", "bioturbation"], "1", "Grey, shelly"],
  [0.5, "gradational", "siltstone", 5, 4, "planar-lamination", [], "2", ""],
  [1.1, "sharp", "sandstone", 3, 2, "current-ripples", ["shell-debris"], "3", "Ripple tops eroded"],
  [0.6, "erosional", "conglomerate", -3, -1, "", ["imbrication"], "4", "Basalt pebbles"],
  [0.9, "sharp", "mudstone", 6, 6, "", ["bioturbation"], "1", ""],
  [0.6, "gradational", "siltstone", 5, 4, "planar-lamination", [], "2", ""],
  [1.4, "sharp", "sandstone", 3, 1.5, "trough-cross-bedding", [], "3", "Sets 10–20 cm"],
  [0.5, "erosional", "conglomerate", -2.5, -1, "", [], "4", ""],
  [0.7, "sharp", "mudstone", 6, 6, "", ["roots", "plant"], "1", "Plant remains"],
  [2.2, "sharp", "lava", undefined, undefined, "", ["vesicles", "columnar-joints"], "5", "Olivine basalt, columnar base"],
];

let browser;
try {
  for (let i = 0; i < 100; i++) {
    try {
      if ((await fetch(BASE + "/api/health")).ok) break;
    } catch {}
    await sleep(100);
  }
  const project = await call("POST", "/api/projects", { name: "Snæfellsnes field course" });
  await call("POST", "/api/projects", { name: "Reykjanes 2026" });
  const log = await call("POST", `/api/projects/${project.id}/logs`, { name: "Búðir section" });
  await call("POST", `/api/projects/${project.id}/logs`, { name: "Arnarstapi cliff" });
  log.beds = BEDS.map(([thickness, contact, pattern, base, top, inBed, symbols, facies, notes], i) => ({
    id: `g${i}`, thickness, contact, lithology: [{ pattern, percent: 100 }], grain: base === undefined ? {} : { base, top },
    ...(inBed && { inBed }), symbols, facies, ...(notes && { notes }),
  }));
  log.beds[3].formation = "Búðir Formation";
  log.beds[9].formation = "Lava group";
  log.columns = log.columns.map((c) => (c.kind === "formation" || c.kind === "facies" ? { ...c, visible: true } : c));
  await call("PUT", `/api/logs/${log.id}`, log);

  browser = await puppeteer.launch({ browser: "firefox", executablePath: process.env.FIREFOX ?? "/usr/bin/firefox", headless: true });
  const page = await browser.newPage();
  await page.setViewport({ width: 1280, height: 800 });
  const shot = async (name) => {
    const png = join(out, `${name}.png`);
    await page.screenshot({ path: png });
    // WebP keeps the pictures small enough to build into the program.
    execFileSync("python3", ["-I", "-c", `from PIL import Image; import os; p=${JSON.stringify(png)}; Image.open(p).save(p[:-4]+'.webp', quality=82); os.remove(p)`]);
    console.log("•", name);
  };

  await page.goto(BASE + "/");
  await page.waitForSelector(".rows a");
  await shot("projects");

  await page.goto(`${BASE}/#/l/${log.id}`);
  await page.waitForSelector(".editor .logsvg");
  await sleep(300);
  await page.$eval(".drawing", (d) => (d.scrollTop = d.scrollHeight));
  const hit = await page.$$(".logsvg .hit");
  await hit[6].click();
  await page.waitForSelector(".panel input");
  await sleep(200);
  await shot("log-editor");

  await page.click("#tab-layout");
  await sleep(200);
  await shot("layout");

  await page.click("button[name=export]");
  await page.waitForSelector("dialog.export[open]");
  await shot("export");
  await page.keyboard.press("Escape");

  await page.click("button[name=stats]");
  await page.waitForSelector("dialog.stats .stats-verdict");
  await page.$eval("#st-fac", (e) => e.scrollIntoView());
  await shot("statistics");
} finally {
  await browser?.close();
  server.kill();
  rmSync(data, { recursive: true, force: true });
}
