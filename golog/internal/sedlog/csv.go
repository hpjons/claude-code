package sedlog

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"
)

// CSVField is a bed field a CSV column can fill.
type CSVField struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// CSVFields are the fields in the order the column-matching table shows them.
var CSVFields = []CSVField{
	{"thickness", "Thickness"},
	{"boundary", "Base contact"},
	{"lith1", "Lithology"}, {"pct1", "Lithology %"},
	{"lith2", "Lithology 2"}, {"pct2", "Lithology 2 %"},
	{"lith3", "Lithology 3"}, {"pct3", "Lithology 3 %"},
	{"grainBase", "Grain size at base"}, {"phiBase", "Phi at base"},
	{"grainTop", "Grain size at top"}, {"phiTop", "Phi at top"},
	{"inBed", "Structure inside the bed"},
	{"symbols", "Structures & fossils"},
	{"notes", "Notes"},
	{"bioType", "Bioturbation type"}, {"bioStrength", "Bioturbation index"},
	{"palaeocurrents", "Palaeocurrents"},
	{"facies", "Facies"},
	{"formation", "Formation"}, {"age", "Age"},
	{"other1", "Other 1 text"}, {"other2", "Other 2 text"}, {"other3", "Other 3 text"},
}

// Header names that identify a field: SedLog's own export headers first, then common alternatives.
var csvHeaders = map[string][]string{
	"thickness":      {"thickness (cm)", "thickness (m)", "thickness", "thick", "thickness_cm", "thickness_m"},
	"boundary":       {"base boundary", "boundary", "contact", "base contact", "basal contact"},
	"lith1":          {"lithology", "lithology1", "lithology 1", "lith", "rock", "rock type"},
	"pct1":           {"lithology %", "lithology1 %", "lithology 1 %"},
	"lith2":          {"lithology2", "lithology 2"},
	"pct2":           {"lithology2 %", "lithology 2 %"},
	"lith3":          {"lithology3", "lithology 3"},
	"pct3":           {"lithology3 %", "lithology 3 %"},
	"grainBase":      {"grain size base", "grain base", "grain size", "grainsize", "grain"},
	"phiBase":        {"phi values base", "phi base", "phi"},
	"grainTop":       {"grain size top", "grain top"},
	"phiTop":         {"phi values top", "phi top"},
	"inBed":          {"symbols in bed", "structure in bed"},
	"symbols":        {"symbols/structures", "symbols", "structures", "fossils", "structures/fossils"},
	"notes":          {"notes column", "notes", "description", "remarks", "comments"},
	"bioType":        {"bioturbation type"},
	"bioStrength":    {"intensity", "bioturbation", "bioturbation index", "bi"},
	"palaeocurrents": {"palaeocurrent values", "palaeocurrent", "paleocurrent", "palaeocurrents", "paleocurrents"},
	"facies":         {"facies"},
	"formation":      {"formation"},
	"age":            {"age"},
	"other1":         {"other1 text", "other1"},
	"other2":         {"other2 text", "other2"},
	"other3":         {"other3 text", "other3"},
}

// CSVInfo describes a CSV file for the column-matching table.
type CSVInfo struct {
	Headers  []string       `json:"headers"`
	Rows     int            `json:"rows"`
	Sample   [][]string     `json:"sample"`  // the first rows, to help match columns
	Mapping  map[string]int `json:"mapping"` // field key → column index
	Units    string         `json:"units"`   // thickness in "cm" or "m"
	TopFirst bool           `json:"topFirst"`
	Fields   []CSVField     `json:"fields"`
}

// CSVOptions overrides the automatic column matching.
type CSVOptions struct {
	Mapping  map[string]int `json:"mapping"`
	Units    string         `json:"units"`
	TopFirst bool           `json:"topFirst"`
}

func readCSV(data []byte) ([][]string, error) {
	text := string(newLines(data).joined())
	first, _, _ := strings.Cut(text, "\n")
	r := csv.NewReader(strings.NewReader(text))
	r.FieldsPerRecord = -1
	r.LazyQuotes = true
	switch {
	case strings.Count(first, ";") > strings.Count(first, ","):
		r.Comma = ';'
	case strings.Count(first, "\t") > strings.Count(first, ","):
		r.Comma = '\t'
	}
	rows, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("could not read the CSV file: %w", err)
	}
	// Drop empty rows (e.g. a blank last line from a spreadsheet).
	out := rows[:0]
	for _, row := range rows {
		if strings.TrimSpace(strings.Join(row, "")) != "" {
			out = append(out, row)
		}
	}
	if len(out) < 1 {
		return nil, fmt.Errorf("the CSV file is empty")
	}
	return out, nil
}

func (ls *lines) joined() []byte { return []byte(strings.Join(ls.l, "\n")) }

// autoMapping matches CSV headers to fields by name.
func autoMapping(headers []string) (map[string]int, string) {
	m := map[string]int{}
	units := "cm"
	norm := make([]string, len(headers))
	for i, h := range headers {
		norm[i] = strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(h))), " ")
	}
	for _, f := range CSVFields {
		for _, name := range csvHeaders[f.Key] {
			found := -1
			for i, h := range norm {
				if h == name && !taken(m, i) {
					found = i
					break
				}
			}
			if found >= 0 {
				m[f.Key] = found
				if f.Key == "thickness" && (strings.Contains(name, "(m)") || strings.HasSuffix(name, "_m")) {
					units = "m"
				}
				break
			}
		}
	}
	return m, units
}

func taken(m map[string]int, i int) bool {
	for _, v := range m {
		if v == i {
			return true
		}
	}
	return false
}

// ParseCSV reads beds from a CSV file. Without options the columns are matched by their headers.
func ParseCSV(data []byte, opt *CSVOptions) (*Result, error) {
	rows, err := readCSV(data)
	if err != nil {
		return nil, err
	}
	headers := rows[0]
	body := rows[1:]
	mapping, units := autoMapping(headers)
	topFirst := false
	if opt != nil {
		if opt.Mapping != nil {
			mapping = map[string]int{}
			for k, v := range opt.Mapping {
				if v >= 0 && v < len(headers) {
					mapping[k] = v
				}
			}
		}
		if opt.Units == "m" || opt.Units == "cm" {
			units = opt.Units
		}
		topFirst = opt.TopFirst
	}
	res := newResult("csv")
	sample := body
	if len(sample) > 5 {
		sample = sample[:5]
	}
	res.CSV = &CSVInfo{Headers: headers, Rows: len(body), Sample: sample, Mapping: mapping, Units: units, TopFirst: topFirst, Fields: CSVFields}
	res.Content.Units = units
	if _, ok := mapping["thickness"]; !ok {
		res.note(0, "Choose which column holds the bed thickness")
		res.finish()
		return res, nil
	}
	if topFirst {
		for i, j := 0, len(body)-1; i < j; i, j = i+1, j-1 {
			body[i], body[j] = body[j], body[i]
		}
	}
	if len(body) > 10000 {
		res.note(0, "Only the first 10,000 of %d beds were imported", len(body))
		body = body[:10000]
	}
	perMetre := 100.0
	if units == "m" {
		perMetre = 1
	}
	for i, row := range body {
		get := func(key string) string {
			if c, ok := mapping[key]; ok && c < len(row) {
				return strings.TrimSpace(row[c])
			}
			return ""
		}
		var rb rawBed
		rb.thickness = get("thickness")
		rb.boundary = get("boundary")
		for k, key := range []string{"1", "2", "3"} {
			rb.lith[k] = get("lith" + key)
			if none(rb.lith[k]) {
				continue
			}
			if p, ok := num(get("pct" + key)); ok {
				rb.share[k] = p / 100
			} else if k == 0 {
				rb.share[k] = 1 // a lone lithology without a percentage fills the bed
			}
		}
		rb.grainBase, rb.phiBase = get("grainBase"), get("phiBase")
		rb.grainTop, rb.phiTop = get("grainTop"), get("phiTop")
		rb.inBed = get("inBed")
		for _, s := range strings.FieldsFunc(get("symbols"), func(c rune) bool { return c == ',' || c == ';' }) {
			rb.symbols = append(rb.symbols, strings.TrimSpace(s))
		}
		rb.notes = get("notes")
		rb.bioType, rb.bioStrength = get("bioType"), get("bioStrength")
		rb.palaeocurrents = get("palaeocurrents")
		rb.facies = get("facies")
		rb.formation, rb.age = get("formation"), get("age")
		for k := range 3 {
			rb.other[k] = get(fmt.Sprintf("other%d", k+1))
		}
		res.addBed(i+1, rb, perMetre)
	}
	res.finish()
	return res, nil
}

// Parse reads any SedLog file, telling the kind from its content.
func Parse(data []byte, csvOpt *CSVOptions) (*Result, error) {
	switch Sniff(bytes.TrimSpace(data)) {
	case "slg":
		return ParseSLG(data)
	case "tem":
		return ParseTemplate(data)
	case "csv":
		return ParseCSV(data, csvOpt)
	}
	return nil, fmt.Errorf("this is not a SedLog log (.slg), template (.tem) or CSV file")
}
