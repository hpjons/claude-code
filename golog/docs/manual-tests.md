# Manual tests

`make e2e` covers most of the editor in headless Firefox: adding and editing beds, colours,
layout, page breaks, dragging bed tops, grain-size edges, column edges and whole beds, undo and
redo (buttons and keys), Alt+↑/↓, Delete with Undo in the message, Duplicate, reload and phone
width. These checks need a person, a real screen or other devices. Run them before a release.

## Mouse and keyboard (desktop, Firefox and Chrome or Edge)

- [ ] Hovering a bed's top edge shows an up-down cursor and a faint line; the grain-size edge
      shows a left-right cursor; a column edge in the header shows a column-resize cursor.
- [ ] While dragging, the small label next to the pointer shows the new value, and the bed's base
      stays in place while its top follows the pointer.
- [ ] Dragging a bed far up or down in a long log: the drawing does not jump.
- [ ] Click the colour swatch of a lithology: the system colour picker opens; choosing a colour
      redraws the bed; one Ctrl+Z undoes the whole colour choice.
- [ ] Typing a sentence in Notes, then clicking the drawing and pressing Ctrl+Z, removes the
      sentence in one step. Inside the Notes box, Ctrl+Z undoes letters as usual.
- [ ] macOS: Cmd+Z and Cmd+Shift+Z undo and redo.
- [ ] Tab reaches the undo and redo buttons, the drawing and every panel field, with a visible
      focus ring.

## Touch (tablet or phone)

- [ ] Swiping over the drawing scrolls it; it does not move beds.
- [ ] Dragging a bed's top edge or the grain-size edge with a finger changes it.
- [ ] Tapping a bed selects it.

## Saving

- [ ] Open the same log in two windows, change it in both: the second window says the log was
      changed elsewhere and shows the latest version; undo there starts afresh.
- [ ] Stop `golog` while editing: the status says "Not saved – retrying", and the change is saved
      once `golog` runs again.
- [ ] Closing the tab right after a change asks before leaving if the save hasn't finished.

## Looks

- [ ] Dark mode (system setting): the editor is dark, the log itself stays white like paper.
- [ ] Zoom 50 % and 300 %: handles are still easy to grab, knobs stay small.

## SedLog import

- [ ] Import a real `.slg` from a colleague's SedLog (not the examples): compare the drawing with
      SedLog's PDF of it; read the "Not brought over exactly" list.
- [ ] Import a CSV saved from Excel or LibreOffice with Icelandic letters in it (UTF-8 and the
      older Windows encoding) and with semicolons as separators.
- [ ] Windows: the file chooser shows `.slg`, `.tem` and `.csv` files.

## Export and print

- [ ] Print a two-page log on a real printer from Firefox and from Chrome or Edge: one page per
      sheet, nothing cut at the edges, 1 m measures 2 cm at 1:50.
- [ ] Open the PDF in Adobe Reader and in the browser's viewer: patterns without seams, Icelandic
      letters right.
- [ ] Open the SVG in Inkscape: patterns and symbols show; text can be edited.
- [ ] Open the CSV in Excel and in LibreOffice: Icelandic letters right, notes with commas in one cell.
- [x] "CSV for SedLog" opens in SedLog 3.1 (Import Log from CSV file, automatic): checked 2026-10-13
      with example2 (39 beds) and a test log; everything comes back except formation, age and
      facies names (SedLog's CSV has no columns for them).
- [ ] Give a `.golog` file to another computer and import it there.

## Statistics

- [ ] A real log with facies from a field course: do the numbers and the wording make sense to a
      sedimentologist? Is the transition grid the right way round (rows: the facies below)?
- [ ] Compare with SedLog's Order Analysis on the same `.slg` (beds with facies 1, 2, 3 …): the
      matrix cells and *m* should match; *r* matches unless the section starts by thinning
      (see docs/sedlog-format.md).
- [ ] **Save as CSV** opens in Excel and LibreOffice with the tables in separate blocks.

## Release (fresh computers)

The Linux package was tested 2026-10-14 in an empty Debian container: unpack, `./golog`, no
browser installed (message shown), projects and beds, statistics, user guide with pictures,
Stop GoLog. Still to do on real machines:

- [ ] **Windows 10/11**, a computer that has never had GoLog: unzip, double-click `golog.exe`,
      get past SmartScreen ("More info › Run anyway"), the browser opens, make a log, print it,
      Stop GoLog closes the black window. Start it twice: the second start opens the running copy.
- [ ] **macOS on Apple Silicon** and, if one is at hand, an Intel Mac: unzip, double-click
      `golog`, "Open Anyway" in Privacy & Security, the browser opens, Stop GoLog.
- [ ] The user guide reads well to a student who hasn't used GoLog; every step in it works.
- [ ] Back up `golog.db`, copy it to another computer, and see the projects there.
