# GoLog – plan

A new program for drawing and analysing graphic sediment logs: a Go backend and a browser
interface in TypeScript and SVG, in English. The idea comes from SedLog 3.1 (Royal Holloway,
2015), but this is its own program with its own design, data model, file format, patterns and
code. It can import SedLog files so existing logs aren't lost.

## 1. What SedLog does – the starting point (tested 2026-10-13)

SedLog 3.1 runs on Java 25 (needs the full `java-25-openjdk`, not just the headless package).
It was driven by a small helper (`docs/original/Drive.java`) that clicks its menus from code.
Screenshots are in `docs/original/` (kept locally, not in the repository).

| Area | Findings |
|---|---|
| Opening files | Ignores a file name on the command line; opens via File › Open. `example1.slg` and `example2.slg` load fine |
| Log view | Columns drawn side by side, beds bottom-up, a page grid ("Page 1 of 2"), zoom 100 % / fit width / fit height, click selects a bed (red outline) |
| Bed editor | Two tabs. **Properties**: thickness, facies number, base boundary, notes, palaeocurrent values, 3 lithologies (group, type, %), grain size base/top (clastic and limestone scales), symbols in bed (tiled or not, or use lithology 1), structures/fossils list, bioturbation type + intensity, age and formation (text + boundary). **Other**: 3 user columns (text, symbols, "add base line") |
| Preferences | Log title, scale start and increment, palaeocurrents as numbers or arrows, text alignment, grain-size header (clastic, limestone or both), symbols wrap, measure from top (cores), log scale, pattern and symbol size, header and cell font sizes, used/unused columns, column order and names |
| Log key | Automatic key of the lithologies, symbols and boundaries used; print, page setup, export |
| Lithologies/symbols | Groups and types, import SVG/SVGZ, edit, delete, edit groups |
| Statistics | "Order analysis": facies transition matrix, Markov order metric and runs analysis of bed thickness, each against a histogram from shuffled sections. **Not in the help pages** |
| Export | PDF (one page, the whole log), SVG, JPEG, CSV. All four tested: files written correctly |
| CSV | Export matches `example1.csv` byte for byte. Import has *automatic* (SedLog's own 25 columns) and *custom* (map any spreadsheet column to a log column) |
| Save | Writes `.slg` format **2.1.1** (`example1.slg` is format **1.0**; saving adds one field per bed). Line endings change from CRLF to LF |
| Templates | `.tem` stores layout only (format 2.1) |

Built-in content: 4 lithology groups (Basic, Carbonates, Other, Specialist), 4 symbol groups
(sedimentary structures, fossils, trace fossils, other), 23 lithology and 47 symbol SVGZ files,
22 grain-size classes with phi values.

## 2. Copyright and how we stay clear of it

The About box says *Copyright (c) 2005–2015 Dimitrios Zervas, Robert Hall, Fionn Murtagh,
Gary Nichols, Peter Burgess. All rights reserved.* No source code or licence is published.

- **No decompiling.** The new app is written from the help pages, the screenshots, and the
  behaviour of the running program (black-box testing), not from SedLog's code.
- **File formats** (`.slg`, `.tem`, `.csv`, `.DAT`) are worked out from sample files, including new
  ones made with the old program for testing. Reading another program's files for compatibility is fine.
- **Patterns and symbols** are SedLog's artwork, so the app does not ship them. Instead:
  - ship our own set, based on the public-domain US FGDC geologic map patterns (USGS) plus
    drawings of our own for the symbols;
  - let a user who has SedLog installed point the app at its `patterns/` folder, so their old
    logs look the same on their own computer.
- If the app is ever to be shared widely, it's worth telling the authors (Gary Nichols and
  Peter Burgess are still active in sedimentology) – they may even be glad to give permission
  for the patterns.

## 3. How it differs from SedLog

Graphic logs themselves are a shared convention in sedimentology (lithology column, grain-size
profile, symbols beside the bed); that part looks familiar in any log program. Everything else
is designed afresh:

| | SedLog | This app |
|---|---|---|
| Structure | One desktop window per log | **Projects** holding several logs (sections) from a field area, with locations and photos |
| Editing | Modal "Bed Properties" dialog with drop-downs | Edit **directly on the log**: drag a bed's top to change thickness, drag the grain-size edge to change grain size, drag beds to reorder; a side panel for the rest |
| Data entry | Desktop only | *(later)* A **field mode** for tablets/phones: big buttons, quick "next bed" entry, works offline and syncs later |
| Columns | Fixed set, plus Other1–Other3 | Any number of **custom columns** of chosen kinds (text, number, symbol, sample, photo, interval) |
| Units | cm | m or cm, decimals; depth (core) or height (outcrop) |
| Correlation | – | *(later)* Several logs **side by side** with tie lines between beds or surfaces |
| Patterns | SedLog's own drawings | Our own set based on the public-domain US FGDC geologic patterns; users add their own SVGs. Each rock has a usual **colour** (any bed can have its own), or the log is drawn in black and white |
| File format | Line-by-line `.slg` / `.tem` | Versioned **JSON** project files; `.slg`, `.tem` and SedLog CSV are **import** formats |
| Statistics | Facies matrix, order metrics | Same analyses from the published methods, plus bed-thickness and grain-size summaries, with its own charts |
| Output | PDF/SVG/JPEG/CSV of one log | PDF/SVG/PNG/CSV of a log, a correlation panel, or a whole project; print with page splitting |

## 4. Goals

1. A clean, modern English interface that a student can use without a manual.
2. Import every SedLog `.slg` (1.0 and 2.1.1), `.tem` and CSV file, and draw it faithfully.
3. Create, edit, analyse and export logs and correlation panels.
4. Run on a single computer **offline** (one program, double-click, opens in the browser).

Later: the same program on a **school server** for a class, a tablet **field mode**, and **correlation panels**.

## 5. Architecture

```
 Browser (TypeScript + SVG)                 Go backend (one executable)
 ┌───────────────────────────┐   JSON/HTTP  ┌──────────────────────────────────┐
 │ log drawing (SVG)         │ ───────────▶ │ REST API  /api/projects, /logs…  │
 │ editor, field mode        │ ◀─────────── │ storage: SQLite (one file)       │
 │ PDF/PNG export            │              │ importers: .slg .tem .csv .svgz  │
 │ offline queue (IndexedDB) │              │ statistics (Markov, runs)        │
 └───────────────────────────┘              │ pattern library, photos          │
                                            │ serves the web app (go:embed)    │
                                            └──────────────────────────────────┘
```

**Division of work:** the browser draws; Go stores, imports, checks and calculates. Drawing lives
in one place only (the TypeScript renderer), so screen, PDF, SVG and print always match.

| Part | Choice | Why |
|---|---|---|
| Backend language | **Go** (1.26+) | One executable per OS (Windows/macOS/Linux, cross-compiled), no runtime to install, fast, easy to deploy on a server |
| HTTP | Standard library `net/http` (method + path routing) | No framework needed |
| Database | SQLite via `modernc.org/sqlite` (pure Go, no C compiler) | One file per installation; easy backup; enough for a school |
| Web app inside the binary | `go:embed` | The executable *is* the whole app |
| Local mode | `app` → starts on `localhost`, opens the browser, data in the user's folder | Works offline, no login |
| Server mode | `app serve` → accounts (teacher/student), class projects, sharing | Same code, extra login layer |
| Frontend | TypeScript (strict) + Vite; small hand-written components | Small, fast, readable |
| Drawing | SVG generated from the data model; SVG `<pattern>` fills | Sharp at any zoom; same output everywhere |
| PDF | jsPDF + svg2pdf.js in the browser | Vector PDF from the same SVG |
| Offline in the field | Service worker + IndexedDB queue, synced when back online | Tablets without a network |
| Tests | Go: `go test` (importers, statistics, API). TS: Vitest + Playwright | Importers tested against SedLog's own saved files |

## 6. Data model (sketch, stored as JSON in SQLite)

```ts
interface Project { id: string; name: string; logs: Log[]; patterns: PatternRef[]; }
interface Log {
  id: string; name: string; location?: { lat: number; lon: number; note?: string };
  units: 'm' | 'cm'; direction: 'up' | 'down';           // outcrop height or core depth
  beds: Bed[];                                            // beds[0] = lowest
  columns: Column[];                                      // order, width, kind, title
}
interface Bed {
  id: string; thickness: number;
  base: { contact: 'sharp' | 'erosional' | 'gradational' | 'unknown' };
  lithology: { pattern: string; percent: number }[];
  grain: { base?: GrainClass; top?: GrainClass };       // clastic or carbonate scale, stored as phi
  structures: SymbolUse[]; fossils: SymbolUse[]; traceFossils: SymbolUse[];
  bioturbation?: number;                                  // index 0–6
  palaeocurrents: number[]; facies?: string; notes?: string;
  custom: Record<string, unknown>;                        // values for custom columns
  photos: string[]; samples: { id: string; height: number }[];
}
interface Correlation { logs: string[]; ties: { from: BedRef; to: BedRef }[]; }
```

## 7. Screens

```
┌ Projects ▸ Field area ▸ Log A ───────────────────────── Export ▾  Share ─┐
├──────────┬──────────────────────────────────────────┬───────────────────┤
│ Logs     │  log drawing — drag edges to edit         │ Bed 7             │
│  Log A   │                                           │  lithology grid   │
│  Log B   │                                           │  grain size slider│
│ Panels   │                                           │  structures …     │
│  A–B     │                                           │  notes, photos    │
└──────────┴──────────────────────────────────────────┴───────────────────┘
```

- **Project view:** logs, correlation panels, map of log locations.
- **Log view:** direct manipulation plus side panel; ↑/↓ moves between beds; undo for everything.
- *(later)* **Field mode:** one bed per screen, big buttons, "add bed above", photo from camera.
- *(later)* **Correlation view:** logs side by side, hang on a chosen surface, draw tie lines.
- **Library:** patterns and symbols in searchable grids; import SVG/SVGZ.

## 8. Milestones – first version (local mode)

| # | Result | Checked by |
|---|---|---|
| 1 ✅ | **Skeleton.** `golog` Go program with SQLite and the embedded web app; opens the browser; create a project and a log, add beds, see them drawn | Runs from one executable on Linux and Windows |
| 2 ✅ | **Drawing.** All standard columns, own base pattern set, scale, page breaks | Visual tests (Playwright screenshots) |
| 3 ✅ | **Editing.** Side panel, direct manipulation, undo/redo | Manual test list ([docs/manual-tests.md](docs/manual-tests.md)) |
| 4 ✅ | **SedLog import.** `.slg` 1.0/2.1.1, `.tem`, CSV (automatic + column mapping). Formats identified by saving test logs from the old program | `example1`/`example2` look the same as in SedLog |
| 5 ✅ | **Output.** PDF (paper sizes, pages), SVG, PNG, CSV, print, automatic key | Open in PDF viewer / Inkscape |
| 6 ✅ | **Statistics.** Facies transitions, Markov order, runs analysis, summaries (in Go) | Same numbers as SedLog on example1 and 15 test logs, except where SedLog departs from the paper (see [docs/sedlog-format.md](docs/sedlog-format.md#order-analysis-statistics-menu)) |
| 7 ✅ | **Release.** Builds for Windows, macOS and Linux; short user guide; MIT licence and third-party notices (`make notices`) | Fresh machine test (Linux done in a clean container; Windows and macOS on the [manual list](docs/manual-tests.md)) |

Milestone 7 done 2026-10-14: version 0.1.0 (`VERSION` file); `make release` makes a zip or tar.gz for Linux (amd64, arm64), Windows and macOS (Apple Silicon, Intel) with a short README and SHA256SUMS.txt; a user guide inside the program (Projects page › User guide) with pictures retaken by `make guide-shots`; **Stop GoLog** on the Projects page for people who start GoLog by double-clicking; a clear message when a paused GoLog (Ctrl+Z) holds the address; on Windows the window stays open when GoLog can't start. The Linux package was tested on a clean Debian system; Windows and macOS still need a person.

Milestone 6 done 2026-10-14: `internal/stats` (Go) and `POST /api/stats`; **Statistics** window with a summary of the section (bed thickness, lithology, grain size, grading, contacts, facies), the facies transition grid and the Markov order metric *m*, the runs order metric *r* for bed thickness, each against 5000 shuffled copies with a p value in words, and **Save as CSV**. Method from Burgess (2016); SedLog's behaviour worked out from its charts on example1 and 15 test logs. example2 has no facies, so SedLog gives no statistics for it; the check used example1 instead. Two places where SedLog departs from the paper (gaps in facies numbers, a thinning run at the base) follow the paper in GoLog.

Milestone 5 done 2026-10-13: "Export & print" window — PDF (jsPDF 4.2.1 + svg2pdf.js 2.8.1, loaded only when used; pages at print scale with repeated headers, shrunk to fit when the columns are too wide), SVG, PNG (300 dpi), CSV in SedLog's columns (reads back into SedLog and GoLog), GoLog file (`.golog`, JSON; read back by Import file…), printing through the browser with the same pages; automatic key (patterns, symbols, contacts) under the log or on its own page. Rocks are now drawn as colour plus a transparent pattern, which removes tile seams in PDF viewers.

Milestone 4 done 2026-10-13: `internal/sedlog` reads `.slg` 1.0 and 2.1.1, `.tem` and CSV (SedLog's export automatically, other CSV files with a column-matching table); import window with preview and a list of what was not brought over exactly; "Use a SedLog layout…" in the Layout tab. Format worked out by black-box tests on SedLog 3.1 and written down in [docs/sedlog-format.md](docs/sedlog-format.md). 14 new symbols and 9 new patterns so that SedLog's own lists all have a GoLog drawing. example1 and example2 import with the same columns, scale and beds as SedLog shows. Not carried over: SedLog's facies chart (GoLog shows the facies number as text), the Other1–3 columns (their text goes into the notes), measuring from the top.

Milestone 3 done 2026-10-13: drag a bed's top (thickness, snapping to a step that suits the zoom), the grain-size edge (base or top, snapping to grain classes), whole beds (reorder) and column edges (width in mm); undo/redo for every change (typing in one field is one step); Alt+↑/↓ moves a bed, Delete deletes it with Undo in the message; Duplicate bed. Covered by `make e2e`; [docs/manual-tests.md](docs/manual-tests.md) lists what needs a person.

Milestone 2 done 2026-10-13: all standard columns (formation, age, structures & fossils with 50 own symbols, in-bed structures, bioturbation index, palaeocurrent arrows/numbers, facies), print scales 1:10–1:1000, A4/A3/Letter page breaks, column layout panel, browser test in `web/e2e` (`make e2e`). Lithology colours added the same day: a pale conventional colour per rock under the black pattern, a colour per lithology in a bed (click the swatch), and a colour / black-and-white switch in the Layout tab.

Milestone 1 done 2026-10-13 (tested in Firefox: create project and log, add/edit/delete beds, reload, phone width, dark mode; builds for Linux, Windows, macOS).

After milestone 2 is a good point to look at it together.

**Later (after the first version):** server mode (teacher/student accounts, class projects,
sharing), correlation panels, tablet field mode with offline sync. The data model already has
room for them (projects, log ids, locations, photos), so nothing needs redoing.

## 9. Decisions

- Name: **GoLog** (working name).
- Interface in English.
- Go backend; own codebase and JSON file format; SedLog files are imported.
- Local mode first; server mode, correlation panels and field mode come later.
