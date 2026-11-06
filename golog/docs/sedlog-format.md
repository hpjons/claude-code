# SedLog's file formats, as GoLog reads them

Worked out from SedLog 3.1's example files and from files saved by the running program after
changing one setting at a time (black-box testing). SedLog's code was not read. The importer is
`internal/sedlog`; its tests write small files in this format.

All files are plain text, one value per line, CRLF line ends (SedLog 3.1 itself writes LF).
Older files may be Latin-1. `<none>` means "not set".

## Log (`.slg`)

```
SEDLOG-VERSION-1.0  or  SEDLOG-VERSION-2.1.1
<number of beds>
<beds, lowest first>
<layout>
```

### One bed

| Field | Notes |
|---|---|
| thickness | centimetres, always |
| base boundary | `<none>`, `Sharp`, `Erosion`, `Gradational` |
| lithology 1, share 1 | name, then fraction 0–1 (CSV: percent) |
| lithology 2, share 2 | |
| lithology 3, share 3 | |
| grain size at base: name, value | clastic class (`clay`, `clay/silt`, … `boulder`) or limestone texture (`mudstone` … `rudstone`); the value sets the profile width (phi-like: clay 10, packstone 1.5, grainstone −0.5) |
| grain size at top: name, value | |
| symbol in bed | drawn inside the grain-size profile |
| *2.1.1:* not tiled | 1 = symbol drawn once instead of repeated |
| number of symbols *n*, then *n* names | structures and fossils column |
| notes | one line |
| bioturbation type | a trace-fossil symbol name |
| bioturbation intensity | whole number |
| palaeocurrents | `90, 100, 120` |
| facies | whole number, 0 = none |
| 3 × other column | text, symbol (0 = none), *2.1.1:* base line 0/1 |
| formation | text; a unit starts at this bed |
| age | text |
| formation boundary, age boundary | 0 none, 1 simple line, 2 dotted line |
| *2.1.1:* use lithology 1 | 1 = in-bed symbol drawn over lithology 1 |

### Layout (after the beds; all of a template)

| Field | Notes |
|---|---|
| 40.5 | header height (unused by GoLog) |
| lithology pattern size, symbol size | 1.0 = 100 % |
| px per cm of section | 37.795 / scale: 0.756 = 1:50, 3.78 = 1:10 |
| title | "Untitled" in templates |
| two font sizes | |
| scale start, scale step | whole numbers |
| palaeocurrents | 0 numbers, 1 numbers and arrows, 2 arrows |
| symbols wrap | 0/1 |
| *2.1.1:* measure from top, text alignment | 0/1, 0 left / 1 centre |
| used columns *u* | |
| 13 × column: title, position, width | in SedLog's fixed order: Age, Formation, Scale, Lithology, Grain size, Structures/fossils, Notes, Bioturbation, Palaeocurrent, Facies, Other1–3. Title and width belong to that column; width in tenths of a mm. The *position* number of entry *k* is the column shown at place *k* from the left, and places 0 to *u*−1 are shown. |
| *2.1.1:* two sizes, 1.0, grain-size scale | scale: 0 limestones and clastics, 1 limestones only, 2 clastics only |

The scale column's title carries the units: `SCALE (CM)` or `SCALE (m)`.

## Template (`.tem`)

`SEDLOG-TEMPLATE-VERSION-2.1` followed by a 2.1.1 layout.

## CSV export

One header row, then one row per bed from the base:
`THICKNESS (CM), BASE BOUNDARY, LITHOLOGY, LITHOLOGY %, LITHOLOGY2, LITHOLOGY2 %, LITHOLOGY3,
LITHOLOGY3 %, GRAIN SIZE BASE, PHI VALUES BASE, GRAIN SIZE TOP, PHI VALUES TOP, SYMBOLS IN BED,
SYMBOLS/STRUCTURES, NOTES COLUMN, BIOTURBATION TYPE, INTENSITY, PALAEOCURRENT VALUES, FACIES,
OTHER1 TEXT, OTHER1 SYMBOL, OTHER2 TEXT, OTHER2 SYMBOL, OTHER3 TEXT, OTHER3 SYMBOL`.
GoLog also reads other CSV files: it matches headings it knows and lets the person choose the rest.

SedLog's own CSV import (tested by importing files GoLog wrote): it refuses a file that starts
with a UTF-8 byte-order mark, and one whose FACIES cells are empty or not whole numbers; it drops
double quotes inside a quoted cell; extra columns after OTHER3 SYMBOL are ignored.

## Order analysis (Statistics menu)

SedLog's *Statistics › Order Analysis Single Section* shows the facies transition matrix, a bar
chart of beds per facies, and the Markov order metric *m* and runs order metric *r*, each against
a histogram from shuffled sections. The window has no numbers other than the matrix cells, so the
values below were read from where the red line sits on the charts, on example1 and 15 test logs
made for the purpose. GoLog's `internal/stats` uses the published method (Burgess 2016,
doi:10.2110/jsr.2016.10); its tests hold these cases.

What SedLog does, and GoLog does the same:

- Facies are SedLog's whole numbers; 0 means none. A bed with no facies **breaks** the succession:
  no transition is counted across it (facies 1 0 2 gives no 1 → 2).
- Two beds of the same facies in a row are not a transition (embedded matrix); the bar chart
  counts beds, not facies units.
- *m*: for each offset *j* = 1 … F−1 from the main diagonal, add up the transition probabilities
  *T*(*i*, (*i*+*j*) mod F); *m* = (largest sum − smallest sum) / F. The main diagonal is left out
  (facies 1 2 3 1 3 2 1 2 3 1 gives *m* = 1/3, not 2/3). With two facies *m* is always 0.
- *r*: every bed counts, with or without a facies, and each bed is a unit (beds of the same
  facies are not merged). A bed as thick as the one below ends a run and scores 0.
- Shuffling swaps two beds at random, *n* times, skipping swaps that would put beds of the same
  facies next to each other. Sections with many same-facies neighbours therefore hardly change
  when shuffled.

Where GoLog differs on purpose:

| | SedLog 3.1 | GoLog |
|---|---|---|
| Gaps in facies numbers (1, 3, 7) | Matrix sized to the largest number (7 × 7) with empty rows, which lowers *m* (1/7 for 1 3 1 3 7 1) | Facies numbered in order without gaps, as the paper advises (2/3 for the same beds). Text codes (St, Sh …) are allowed and ordered by first appearance |
| Thinning run at the base | While the beds at the base have not yet got thicker, bed *k* scores *k* − 1, one less than the formula (six steadily thinning beds: 10/6 instead of 15/6) | The paper's formula, whose own worked example scores the first thinning bed 1 |
| Histogram axes | Fixed 0–1 and 0–3; values above 3 fall off the chart | Axes widen to fit |
| p value | Not shown | Shown, with Burgess's wording (< 0.01 very strong evidence of order, < 0.05 strong, ≤ 0.10 weak, above that none) |

Shuffling is random, so the histograms only match in shape; GoLog seeds its random numbers from
the beds, so a log always gives the same numbers. example2 has no facies, so SedLog shows no
statistics for it.
