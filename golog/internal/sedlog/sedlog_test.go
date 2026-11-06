package sedlog

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// bed211 writes one bed the way SedLog 2.1.1 does (thickness in cm, CRLF line ends).
func bed211(thick, boundary, lith, grainBase, phiBase, grainTop, phiTop, inBed string, symbols []string, notes, pc, facies, formation string) []string {
	l := []string{thick, boundary, lith, "1.0", "<none>", "0.0", "<none>", "0.0", grainBase, phiBase, grainTop, phiTop, inBed, "0", itoa(len(symbols))}
	l = append(l, symbols...)
	l = append(l, notes, "<none>", "0", pc, facies)
	l = append(l, "", "0", "0", "", "0", "0", "", "0", "0") // other columns: text, symbol, base line
	return append(l, formation, "", "0", "0", "0")
}

// layout211 writes the settings block: 1:50, the given used count and column positions.
func layout211(title string, palaeo int, used int, show [13]int) []string {
	names := []string{"AGE", "FORMATION", "SCALE (CM)", "LITHOLOGY", "GRAIN SIZE", "STRUCTURES / FOSSILS", "NOTES", "BIOTURBATION", "PALAEOCURRENT", "FACIES", "OTHER1", "OTHER2", "OTHER3"}
	l := []string{"40.5", "1.0", "1.0", "0.7559055", title, "11.0", "11.0", "0", "1", itoa(palaeo), "0", "0", "0", itoa(used)}
	for k, n := range names {
		w := "200.0"
		if k == 4 {
			w = "400.0"
		}
		l = append(l, n, itoa(show[k]), w)
	}
	return append(l, "50.0", "380.0", "1.0", "0")
}

var itoa = strconv.Itoa

func file(parts ...[]string) []byte {
	var all []string
	for _, p := range parts {
		all = append(all, p...)
	}
	return []byte(strings.Join(all, "\r\n") + "\r\n")
}

func TestSLG211(t *testing.T) {
	data := file(
		[]string{"SEDLOG-VERSION-2.1.1", "2"},
		bed211("120.0", "<none>", "Sandstone", "m", "1.5", "f", "2.5", "Trough cross bedding", []string{"Bivalves", "Unicorns"}, "Cross-bedded", "85, 120", "3", "Lower Fm"),
		bed211("35.5", "Erosion", "Wackestone", "packstone", "1.5", "packstone", "1.5", "<none>", nil, "", "", "0", ""),
		layout211("Test section", 2, 3, [13]int{5, 3, 0, 1, 2, 6, 7, 8, 9, 10, 11, 12, 4}),
	)
	r, err := Parse(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Format != "slg" || r.Name != "Test section" || r.Content.Scale != 50 || r.Content.Units != "cm" || r.Content.Palaeocurrent != "arrows" {
		t.Fatalf("settings: %+v", r)
	}
	var shown []string
	for _, c := range r.Content.Columns {
		if c.Visible {
			shown = append(shown, c.Kind)
		}
	}
	// Positions 0–2 show SedLog columns 5, 3 and 0: structures, lithology, age.
	if !reflect.DeepEqual(shown, []string{"symbols", "lithology", "age"}) {
		t.Errorf("shown columns: %v", shown)
	}
	if len(r.Content.Columns) != 10 {
		t.Errorf("all GoLog columns should be kept, got %d", len(r.Content.Columns))
	}
	b := r.Content.Beds
	if len(b) != 2 || b[0].Thickness != 1.2 || b[1].Thickness != 0.355 {
		t.Fatalf("beds: %+v", b)
	}
	if *b[0].Grain.Base != 1.5 || *b[0].Grain.Top != 2.5 || b[0].InBed != "trough-cross-bedding" || b[0].Formation != "Lower Fm" || b[0].Facies != "3" {
		t.Errorf("bed 1: %+v", b[0])
	}
	if !reflect.DeepEqual(b[0].Symbols, []string{"bivalve"}) || !reflect.DeepEqual(b[0].Palaeocurrents, []float64{85, 120}) {
		t.Errorf("bed 1 symbols/palaeocurrents: %v %v", b[0].Symbols, b[0].Palaeocurrents)
	}
	if b[1].Contact != "erosional" || b[1].Lithology[0].Pattern != "wackestone" || b[1].Facies != "" || b[1].Grain.Top != nil {
		t.Errorf("bed 2: %+v", b[1])
	}
	warn := strings.Join(r.Warnings, "\n")
	for _, want := range []string{`no symbol for "Unicorns"`, "Limestone textures", "(bed 2)"} {
		if !strings.Contains(warn, want) {
			t.Errorf("warnings lack %q:\n%s", want, warn)
		}
	}
	if err := r.Content.Validate(); err != nil {
		t.Errorf("imported log is not valid: %v", err)
	}
}

func TestSLG10AndLatin1(t *testing.T) {
	// Version 1.0: no tiling flag, two lines per "other" column, no last flag, a shorter layout.
	bed := []string{"100.0", "Sharp", "Siltstone", "0.6", "Claystone", "0.4", "<none>", "0.0", "clay/silt", "8.0", "clay/silt", "8.0", "<none>", "0",
		"Gr\xe1tt", "<none>", "0", "", "2", "", "0", "", "0", "", "0", "", "", "0", "0"}
	lay := []string{"40.5", "1.0", "1.0", "3.7795277", "Section", "11.0", "11.0", "0", "1", "0", "0", "2"}
	for k, n := range []string{"AGE", "FORMATION", "SCALE (m)", "LITHOLOGY", "GRAIN SIZE", "STRUCTURES / FOSSILS", "NOTES", "BIOTURBATION", "PALAEOCURRENT", "FACIES", "OTHER1", "OTHER2", "OTHER3"} {
		lay = append(lay, n, itoa((k+2)%13), "100.0")
	}
	lay = append(lay, "50.0", "360.0")
	r, err := ParseSLG(file([]string{"SEDLOG-VERSION-1.0", "1"}, bed, lay))
	if err != nil {
		t.Fatal(err)
	}
	b := r.Content.Beds[0]
	if b.Notes != "Grátt" {
		t.Errorf("Latin-1 notes: %q", b.Notes)
	}
	if len(b.Lithology) != 2 || b.Lithology[0].Percent != 60 || b.Lithology[1].Pattern != "claystone" {
		t.Errorf("lithology: %+v", b.Lithology)
	}
	if r.Content.Scale != 10 || r.Content.Units != "m" || r.Content.Palaeocurrent != "numbers" {
		t.Errorf("settings: scale %d units %s pc %s", r.Content.Scale, r.Content.Units, r.Content.Palaeocurrent)
	}
	if c := r.Content.Columns; !c[0].Visible || c[0].Kind != "scale" || !c[1].Visible || c[1].Kind != "lithology" || c[2].Visible {
		t.Errorf("columns: %+v", c[:3])
	}
}

func TestBrokenFiles(t *testing.T) {
	for name, data := range map[string]string{
		"cut short":   "SEDLOG-VERSION-2.1.1\r\n3\r\n100.0\r\nSharp\r\n",
		"bad count":   "SEDLOG-VERSION-2.1.1\r\nmany\r\n",
		"not sedlog":  "hello world",
		"empty":       "",
		"bad symbols": strings.Join(append([]string{"SEDLOG-VERSION-2.1.1", "1"}, bed211("1", "<none>", "Sandstone", "m", "1.5", "m", "1.5", "<none>", nil, "", "", "0", "")[:13]...), "\n") + "\n0\n99999999\n",
	} {
		if _, err := Parse([]byte(data), nil); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestTemplate(t *testing.T) {
	r, err := Parse(file([]string{"SEDLOG-TEMPLATE-VERSION-2.1"}, layout211("Untitled", 0, 4, [13]int{2, 3, 4, 5, 4, 5, 6, 7, 8, 9, 10, 11, 12})), nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.Format != "tem" || r.Name != "" || len(r.Content.Beds) != 0 {
		t.Fatalf("template: %+v", r)
	}
	var shown []string
	for _, c := range r.Content.Columns {
		if c.Visible {
			shown = append(shown, c.Kind)
		}
	}
	if !reflect.DeepEqual(shown, []string{"scale", "lithology", "grain", "symbols"}) {
		t.Errorf("template columns: %v", shown)
	}
	for _, c := range r.Content.Columns {
		if c.Kind == "grain" && c.Width != 151.2 {
			t.Errorf("grain width %.1f, want 40 mm = 151.2 px", c.Width)
		}
	}
}

func TestCSV(t *testing.T) {
	sedlog := "THICKNESS (CM),BASE BOUNDARY,LITHOLOGY,LITHOLOGY %,LITHOLOGY2,LITHOLOGY2 %,LITHOLOGY3,LITHOLOGY3 %,GRAIN SIZE BASE,PHI VALUES BASE,GRAIN SIZE TOP,PHI VALUES TOP,SYMBOLS IN BED,SYMBOLS/STRUCTURES,NOTES COLUMN,BIOTURBATION TYPE,INTENSITY,PALAEOCURRENT VALUES,FACIES,OTHER1 TEXT,OTHER1 SYMBOL,OTHER2 TEXT,OTHER2 SYMBOL,OTHER3 TEXT,OTHER3 SYMBOL\n" +
		`45.0,Erosion,Conglomerate,100,<none>,0,<none>,0,pebble,-3.0,granule,-1.5,<none>,"Roots, Bivalves","Polymict, coarse",<none>,0,"90, 100",4,"sample 7","","","","",""` + "\n"
	r, err := Parse([]byte(sedlog), nil)
	if err != nil {
		t.Fatal(err)
	}
	b := r.Content.Beds[0]
	if r.Format != "csv" || b.Thickness != 0.45 || b.Contact != "erosional" || *b.Grain.Base != -3 || b.Notes != "Polymict, coarse\nOther 1: sample 7" {
		t.Errorf("SedLog CSV bed: %+v", b)
	}
	if !reflect.DeepEqual(b.Symbols, []string{"roots", "bivalve"}) || !reflect.DeepEqual(b.Palaeocurrents, []float64{90, 100}) {
		t.Errorf("symbols %v palaeocurrents %v", b.Symbols, b.Palaeocurrents)
	}

	// A spreadsheet of one's own: semicolons, metres, top bed first, grain names without phi.
	own := "Bed;Thickness (m);Rock;Grain size;Description\n3;0.5;Siltstone;silt;top bed\n2;1,5;Sandstone;f/m;\n1;2;Tuffaceous sandstone;c;bottom bed\n"
	r, err = ParseCSV([]byte(own), nil)
	if err != nil {
		t.Fatal(err)
	}
	if r.CSV.Units != "m" || r.CSV.Mapping["thickness"] != 1 || r.CSV.Mapping["lith1"] != 2 || r.CSV.Mapping["notes"] != 4 || r.CSV.Rows != 3 {
		t.Fatalf("automatic matching: %+v", r.CSV)
	}
	r, err = ParseCSV([]byte(own), &CSVOptions{Mapping: r.CSV.Mapping, Units: "m", TopFirst: true})
	if err != nil {
		t.Fatal(err)
	}
	beds := r.Content.Beds
	// "1,5" is not a number here, so that bed is left out and reported.
	if len(beds) != 2 || beds[0].Notes != "bottom bed" || beds[0].Thickness != 2 || *beds[1].Grain.Base != 6 {
		t.Errorf("own CSV beds: %+v", beds)
	}
	if beds[0].Lithology[0].Pattern != "sandstone" {
		t.Errorf("tuffaceous sandstone drawn as %s", beds[0].Lithology[0].Pattern)
	}
	if !strings.Contains(strings.Join(r.Warnings, "\n"), "without a thickness") {
		t.Errorf("warnings: %v", r.Warnings)
	}

	// No thickness column: nothing to import until one is chosen.
	r, err = ParseCSV([]byte("Rock,Notes\nSandstone,x\n"), nil)
	if err != nil || len(r.Content.Beds) != 0 || !strings.Contains(r.Warnings[0], "thickness") {
		t.Errorf("no thickness column: %v %+v", err, r)
	}
}

func TestBedList(t *testing.T) {
	if got := bedList([]int{7, 2, 3, 4, 9}); got != "beds 2–4, 7, 9" {
		t.Errorf("bedList: %q", got)
	}
}

// SedLog's own examples, when SedLog is installed next to GoLog. They are not part of GoLog.
func TestSedLogExamples(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "sedlog-3.1", "examples")
	if _, err := os.Stat(dir); err != nil {
		t.Skip("SedLog examples not found")
	}
	for _, name := range []string{"example1", "example2"} {
		slg, _ := os.ReadFile(filepath.Join(dir, name+".slg"))
		csv, _ := os.ReadFile(filepath.Join(dir, name+".csv"))
		a, err := Parse(slg, nil)
		if err != nil {
			t.Fatal(err)
		}
		b, err := Parse(csv, nil)
		if err != nil {
			t.Fatal(err)
		}
		ja, _ := json.Marshal(a.Content.Beds)
		jb, _ := json.Marshal(b.Content.Beds)
		if string(ja) != string(jb) {
			t.Errorf("%s: the .slg and the CSV export give different beds", name)
		}
		if err := a.Content.Validate(); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
}

// Every pattern and symbol GoLog draws must come back by its own name, since GoLog's CSV
// export writes GoLog's name when SedLog has none.
func TestGoLogNamesReadBack(t *testing.T) {
	re := regexp.MustCompile(`id: "([^"]+)", name: "([^"]+)"`)
	for file, lookup := range map[string]func(string) (string, bool){
		"patterns.ts": func(n string) (string, bool) { id, _ := lithology(n); return id, id != "" },
		"symbols.ts":  symbol,
	} {
		src, err := os.ReadFile(filepath.Join("..", "..", "web", "src", "render", file))
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range re.FindAllStringSubmatch(string(src), -1) {
			if got, ok := lookup(m[2]); !ok || got != m[1] {
				t.Errorf("%s: %q reads back as %q, want %q", file, m[2], got, m[1])
			}
		}
	}
}
