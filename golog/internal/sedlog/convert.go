// Package sedlog imports logs, templates and CSV files from SedLog (Royal Holloway), working out
// their layout from sample files and files saved by the program, not from its code.
package sedlog

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"golog/internal/model"
)

// Result is an imported log: its name, its content and what could not be brought over exactly.
type Result struct {
	Format   string            `json:"format"` // "slg", "tem", "csv" (or "golog" for GoLog's own files)
	Name     string            `json:"name"`
	Content  model.LogContent  `json:"content"`
	Warnings []string          `json:"warnings"`
	CSV      *CSVInfo          `json:"csv,omitempty"`
	notes    map[string][]int  // warning text → beds it applies to
	order    []string          // warnings in the order first seen
	others   [3]string         // titles of SedLog's three "other" columns
	mapped   map[string]string // what an unknown name was drawn as
}

// rawBed is one bed as SedLog stores it: text fields, thickness in cm.
type rawBed struct {
	thickness            string
	boundary             string
	lith                 [3]string
	share                [3]float64 // 0–1
	grainBase, phiBase   string
	grainTop, phiTop     string
	inBed                string
	symbols              []string
	notes                string
	bioType, bioStrength string
	palaeocurrents       string
	facies               string
	other                [3]string
	age, formation       string
}

func newResult(format string) *Result {
	return &Result{Format: format, Content: model.NewLogContent(), notes: map[string][]int{}, mapped: map[string]string{}, others: [3]string{"Other 1", "Other 2", "Other 3"}}
}

// note records a warning for bed n (0 = about the whole file).
func (r *Result) note(n int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if _, ok := r.notes[msg]; !ok {
		r.order = append(r.order, msg)
	}
	if n > 0 && !slices.Contains(r.notes[msg], n) {
		r.notes[msg] = append(r.notes[msg], n)
	}
}

// finish turns the collected notes into readable warnings ("… (beds 2–4, 7)").
func (r *Result) finish() {
	r.Warnings = []string{}
	for _, msg := range r.order {
		if beds := r.notes[msg]; len(beds) > 0 {
			msg += " (" + bedList(beds) + ")"
		}
		r.Warnings = append(r.Warnings, msg)
	}
}

func bedList(beds []int) string {
	sort.Ints(beds)
	var parts []string
	for i := 0; i < len(beds); {
		j := i
		for j+1 < len(beds) && beds[j+1] == beds[j]+1 {
			j++
		}
		if j > i {
			parts = append(parts, fmt.Sprintf("%d–%d", beds[i], beds[j]))
		} else {
			parts = append(parts, strconv.Itoa(beds[i]))
		}
		i = j + 1
	}
	word := "bed "
	if len(beds) > 1 {
		word = "beds "
	}
	if len(parts) > 8 {
		parts = append(parts[:8], "…")
	}
	return word + strings.Join(parts, ", ")
}

func none(s string) bool {
	s = strings.TrimSpace(s)
	return s == "" || s == "<none>"
}

func num(s string) (float64, bool) {
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v, err == nil && !math.IsNaN(v) && !math.IsInf(v, 0)
}

// addBed converts a SedLog bed and appends it. n is the bed number counted from the base.
func (r *Result) addBed(n int, rb rawBed, unitsPerMetre float64) {
	t, ok := num(rb.thickness)
	if !ok || t <= 0 {
		r.note(n, "Beds without a thickness were left out")
		return
	}
	bed := model.Bed{
		ID:        fmt.Sprintf("b%d", n),
		Thickness: math.Round(t/unitsPerMetre*1e6) / 1e6,
		Contact:   contact(rb.boundary),
		Lithology: []model.Lithology{},
	}
	if bed.Thickness > model.MaxThicknessM {
		bed.Thickness = model.MaxThicknessM
		r.note(n, "A bed was thicker than %d m and was cut to that", model.MaxThicknessM)
	}

	total := 0.0
	for i := range 3 {
		if none(rb.lith[i]) || rb.share[i] <= 0 {
			continue
		}
		id, exact := lithology(rb.lith[i])
		switch {
		case id == "":
			id = "sandstone"
			r.note(n, "GoLog has no pattern for %q; drawn as sandstone, change it in the Bed panel", rb.lith[i])
		case !exact:
			r.note(n, "%q is drawn with GoLog's %s pattern", rb.lith[i], strings.ReplaceAll(id, "-", " "))
		}
		pct := math.Round(rb.share[i]*1000) / 10
		if total+pct > 100 {
			pct = 100 - total
		}
		if pct <= 0 {
			continue
		}
		total += pct
		bed.Lithology = append(bed.Lithology, model.Lithology{Pattern: id, Percent: pct})
	}
	if len(bed.Lithology) == 1 {
		bed.Lithology[0].Percent = 100 // SedLog leaves single lithologies at 100 % but rounding may not
	}

	bed.Grain.Base = grain(rb.grainBase, rb.phiBase)
	bed.Grain.Top = grain(rb.grainTop, rb.phiTop)
	if bed.Grain.Base == nil && bed.Grain.Top != nil {
		bed.Grain.Base, bed.Grain.Top = bed.Grain.Top, nil
	}
	if bed.Grain.Top != nil && bed.Grain.Base != nil && *bed.Grain.Top == *bed.Grain.Base {
		bed.Grain.Top = nil
	}
	for _, g := range []string{rb.grainBase, rb.grainTop} {
		if carbonateTextures[strings.ToLower(strings.TrimSpace(g))] {
			r.note(n, "Limestone textures (mudstone to rudstone) are drawn at the matching width on GoLog's grain-size scale")
		}
	}

	var syms []string
	addSym := func(name string) {
		if none(name) {
			return
		}
		id, ok := symbol(name)
		if !ok {
			r.note(n, "GoLog has no symbol for %q; it was left out", strings.TrimSpace(name))
			return
		}
		if !slices.Contains(syms, id) {
			syms = append(syms, id)
		}
	}
	if !none(rb.inBed) {
		if id, ok := symbol(rb.inBed); ok && inBedSymbols[id] {
			bed.InBed = id
		} else {
			addSym(rb.inBed) // GoLog draws only some structures inside the bed; others go beside it
		}
	}
	for _, s := range rb.symbols {
		addSym(s)
	}
	addSym(rb.bioType)
	if v, ok := num(rb.bioStrength); ok && v > 0 {
		bi := int(math.Round(math.Min(6, v)))
		bed.Bioturbation = &bi
		if v > 6 {
			r.note(n, "Bioturbation intensities above 6 were set to 6")
		}
	}
	if len(syms) > 0 {
		bed.Symbols = syms
	}

	if pc := strings.TrimSpace(rb.palaeocurrents); pc != "" {
		for _, part := range strings.FieldsFunc(pc, func(c rune) bool { return c == ',' || c == ';' || c == ' ' }) {
			if v, ok := num(part); ok && v >= 0 && v <= 360 {
				bed.Palaeocurrents = append(bed.Palaeocurrents, v)
			} else {
				r.note(n, "Palaeocurrent readings that are not directions were left out")
			}
		}
	}
	if f := strings.TrimSpace(rb.facies); f != "" && f != "0" && f != "<none>" {
		bed.Facies = f
	}
	bed.Age = clip(rb.age, model.MaxNameLen)
	bed.Formation = clip(rb.formation, model.MaxNameLen)

	notes := strings.TrimSpace(rb.notes)
	for i, o := range rb.other {
		if o = strings.TrimSpace(o); o != "" {
			// GoLog has no free columns yet, so these go into the notes with the column's name.
			if notes != "" {
				notes += "\n"
			}
			notes += r.others[i] + ": " + o
			r.note(0, "Text from the %s column was added to the notes", r.others[i])
		}
	}
	bed.Notes = clip(notes, model.MaxNotesLen)
	r.Content.Beds = append(r.Content.Beds, bed)
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
