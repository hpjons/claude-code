# GoLog

Draw graphic sediment logs in the browser. GoLog is one program: run it and it opens in your
browser, keeps your work in a small database on your computer, and needs no internet.

The idea comes from SedLog (Royal Holloway), but GoLog is a separate program with its own code,
design, patterns and file format. See [PLAN.md](PLAN.md).

GoLog was written with the AI coding assistant Claude Code (Anthropic), directed and tested by
Helgi Páll Jónsson. The code has automated tests (`make test`, `make e2e`) and a manual checklist
([docs/manual-tests.md](docs/manual-tests.md)), but no person has reviewed every line.

## Download and start

Release packages (`make release`, see below) hold one program and a short README for each
platform: `golog.exe` on Windows, `golog` on macOS and Linux. Double-click it (or run `./golog`)
and GoLog opens in the browser. The first start on Windows and macOS shows a warning because the
program isn't signed; the package README and the user guide say how to get past it.

**User guide:** the Projects page links to it (`/guide.html`, inside the program, works offline).
Its pictures are in `web/public/guide/`; `make guide-shots` retakes them from a sample project.

## Use

```bash
./golog              # starts GoLog and opens http://127.0.0.1:7341
./golog -no-browser  # start without opening the browser
./golog -data DIR    # keep the database somewhere else
```

Stop it with **Stop GoLog** at the bottom of the Projects page, or Ctrl+C in its terminal (not
Ctrl+Z, which only pauses it and keeps the address taken). Starting it again while it runs just
opens the browser on the running copy.

The database is `golog.db` in `~/.local/share/GoLog` (Linux), `~/Library/Application Support/GoLog`
(macOS) or `%AppData%\GoLog` (Windows). Copy that file to back up everything.

### Editing on the drawing

- Drag a bed's **top edge** to change its thickness, the **grain-size edge** to change grain size
  (lower half: at the base, upper half: at the top), a **bed** to move it, or a **column edge** in
  the header to change the column's width.
- **Undo / redo:** the ↶ ↷ buttons, Ctrl+Z and Ctrl+Shift+Z (or Ctrl+Y).
- With the drawing focused: ↑ ↓ selects the next bed, Alt+↑ ↓ moves it, Delete deletes it, Esc closes it.

### Export and print

**Export & print…** in a log's top bar saves the log as:

- **PDF**: pages on the paper chosen in the Layout tab, at the log's scale, with the column
  headers repeated on every page.
- **SVG** or **PNG**: the whole log as one drawing or picture (PNG at 300 dpi).
- **CSV**: the beds as a table in SedLog's columns (plus formation and age), either for
  spreadsheets or for SedLog's "Import Log from CSV file" (SedLog refuses the marker Excel needs
  for Icelandic letters, so there are two copies).
- **GoLog file**: everything in the log, to back up or to give to someone with GoLog.

**Print…** sends the same pages as the PDF to a printer. A key to the patterns, symbols and contacts
used is added below the log unless you turn it off. If the columns are wider than the paper, the
pages are shrunk to fit and the window says so.

### Statistics

**Statistics** in a log's top bar shows what the section is made of: bed thicknesses (mean,
median, range), and how much of the section each lithology, grain-size group (mud, sand,
gravel), grading trend, contact type and facies takes up. Then two tests for order, from
Burgess (2016), *Journal of Sedimentary Research* 86, 148–167:

- **Are the facies ordered?** The facies transition matrix (how often each facies is followed by
  each other one) and the Markov order metric *m*: 1 for a strict cycle, near 0 for random order.
- **Do beds thicken or thin upward?** The runs order metric *r*: long thickening or thinning runs
  give a high value.

Each is compared with 5000 copies of the section shuffled into random order; *p* is the share of
copies that score as high or higher, put into words (a small *p* means the order is unlikely to
be chance). The facies test needs a facies code on the beds (turn on the Facies column in the
Layout tab). **Save as CSV** writes all the numbers for a spreadsheet. The same log always gives
the same numbers. How GoLog's numbers compare with SedLog's is in
[docs/sedlog-format.md](docs/sedlog-format.md#order-analysis-statistics-menu).

### Bringing logs over from SedLog

On a project's page, **Import file…** reads a GoLog file or a SedLog log (`.slg`, versions 1.0 and 2.1.1),
template (`.tem`) or CSV file. It shows a preview and lists anything that could not be brought
over exactly before the log is created. For CSV files from elsewhere, GoLog matches the columns
by their headings and you choose the rest. In a log's Layout tab, **Use a SedLog layout…** takes
the columns and scale from a template. The format is described in
[docs/sedlog-format.md](docs/sedlog-format.md).

## Build

Needs Go 1.26+ and Node 20+.

```bash
make           # builds ./golog for this computer (web app included)
make test      # Go and TypeScript unit tests
make e2e       # builds, then clicks through the app in headless Firefox (SHOTS=dir keeps screenshots)
make release   # dist/: zip or tar.gz per platform (Linux, Windows, macOS; version from VERSION) + SHA256SUMS.txt
make guide-shots  # retake the user guide's pictures
make notices   # rewrite the licence notices after changing dependencies (commit the result)
make dev       # live-reloading web app on http://localhost:5173 (run ./golog -no-browser too)
```

## Layout

| Path | What |
|---|---|
| `main.go` | Starts the server, finds the data folder, opens the browser |
| `internal/model` | Projects, logs and beds, and their validation |
| `internal/store` | SQLite storage (one file, schema migrations) |
| `internal/sedlog` | Reads SedLog `.slg`, `.tem` and CSV files |
| `internal/stats` | Summaries, facies transitions, Markov and runs order metrics with shuffling |
| `internal/server` | JSON API, localhost-only guard, serves the web app |
| `web/` | The web app (TypeScript, Vite); built into `web/dist` and embedded in the program |
| `web/src/render` | Draws a log as SVG at print scale; GoLog's own lithology patterns and symbols |
| `web/src/ui` | Screens: projects, log editor, bed panel, layout panel, symbol picker, import, export, statistics |
| `web/e2e` | Browser test (puppeteer-core driving the installed Firefox); `guide-shots.mjs` takes the guide's pictures |
| `web/public/guide.html` | The user guide |
| `packaging/` | README for the release packages |
| `web/scripts/notices.mjs` | Writes `web/public/notices.txt`, GoLog's licence and those of the libraries built into it (`make notices`) |

## What a log holds

Beds from the base up, each with thickness, up to three lithologies, grain size at base and top,
base contact, a structure drawn inside the bed, structures/fossils/traces beside it, bioturbation
index (0–6), palaeocurrent directions, facies, the formation or age unit it starts, and notes.

The layout belongs to the log: print scale (1:10–1:1000), paper (A4, A3, Letter), palaeocurrents
as arrows or numbers, whether the grain-size names, groups and title are shown above the grain-size
column, and the columns (shown or hidden, order, title, width in mm). Dashed blue
lines show where printed pages will end.

## API (local only)

| Call | Does |
|---|---|
| `GET /api/health` | Version, and whether the app may stop the program |
| `POST /api/quit` | Stop GoLog ("Stop GoLog" button) |
| `GET /api/projects` | List projects |
| `POST /api/projects` `{name}` | New project |
| `GET/PATCH/DELETE /api/projects/{id}` | Project with its logs / rename / delete |
| `POST /api/projects/{id}/logs` `{name}` | New log |
| `GET/PUT/DELETE /api/logs/{id}` | Log with its beds / save / delete |
| `POST /api/import` `{data, csv?}` | Read a GoLog or SedLog file (base64) without saving it |
| `POST /api/stats` `{content, shuffles?}` | Statistics for the log content sent (saved or not) |

A save sends the whole log with the `revision` it was based on; if the log changed in the
meantime the reply is `409` with the current version. The server only answers requests
addressed to localhost and refuses changes sent from other websites.

## Licence

GoLog is released under the [MIT licence](LICENSE). The libraries built into it keep their own
licences (MIT, BSD, ISC, Apache-2.0, zlib, MPL-2.0 for a few unchanged files in modernc.org,
and public domain for SQLite; none is GPL); `make notices` collects their texts into
`web/public/notices.txt`, which the program serves at `/notices.txt` (linked at the bottom of
the Projects page) and the release packages carry as `NOTICES.txt`.
