package sedlog

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Files are one value per line. Version 2.1.1 (SedLog 2.1 and later) adds a few fields per bed
// and per layout over version 1.0; templates hold only the layout.
const (
	headLog10     = "SEDLOG-VERSION-1.0"
	headLog211    = "SEDLOG-VERSION-2.1.1"
	headTemplate  = "SEDLOG-TEMPLATE-VERSION-2.1"
	pxPerCmAt1to1 = 37.795277 // SedLog stores the scale as screen px per cm of section
)

// lines reads a SedLog text file: Windows or Unix line ends, UTF-8 or (older files) Latin-1.
type lines struct {
	l   []string
	pos int
}

func newLines(data []byte) *lines {
	s := ""
	if utf8.Valid(data) {
		s = string(data)
	} else {
		var b strings.Builder
		for _, c := range data {
			b.WriteRune(rune(c)) // Latin-1: each byte is its own code point
		}
		s = b.String()
	}
	s = strings.TrimPrefix(s, "\uFEFF")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return &lines{l: strings.Split(s, "\n")}
}

var errShort = errors.New("the file ends too early")

func (ls *lines) next() string {
	if ls.pos >= len(ls.l) {
		panic(errShort)
	}
	ls.pos++
	return ls.l[ls.pos-1]
}

func (ls *lines) int() int {
	s := strings.TrimSpace(ls.next())
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("line %d: expected a whole number, found %q", ls.pos, s))
	}
	return v
}

func (ls *lines) float() float64 {
	s := strings.TrimSpace(ls.next())
	v, ok := num(s)
	if !ok {
		panic(fmt.Errorf("line %d: expected a number, found %q", ls.pos, s))
	}
	return v
}

// Sniff tells what kind of SedLog file data is: "slg", "tem", "csv" or "".
func Sniff(data []byte) string {
	first, _, _ := bytes.Cut(bytes.TrimPrefix(data, []byte("\uFEFF")), []byte("\n"))
	head := strings.TrimSpace(string(first))
	switch {
	case head == headLog10 || head == headLog211 || strings.HasPrefix(head, "SEDLOG-VERSION-"):
		return "slg"
	case strings.HasPrefix(head, "SEDLOG-TEMPLATE-VERSION-"):
		return "tem"
	case strings.Contains(head, ","), strings.Contains(head, ";"), strings.Contains(head, "\t"):
		return "csv"
	}
	return ""
}

// ParseSLG reads a SedLog log (.slg).
func ParseSLG(data []byte) (res *Result, err error) {
	ls := newLines(data)
	defer catch(&err, ls)
	head := strings.TrimSpace(ls.next())
	v211 := head != headLog10
	if head != headLog10 && head != headLog211 {
		return nil, fmt.Errorf("this is not a SedLog log file (it starts with %q)", clip(head, 40))
	}
	res = newResult("slg")
	n := ls.int()
	if n < 0 || n > 100000 {
		return nil, fmt.Errorf("line 2: %d beds is not a sensible number", n)
	}
	beds := make([]rawBed, n)
	for i := range beds {
		beds[i] = readBed(ls, v211)
	}
	if ls.pos < len(ls.l) && strings.TrimSpace(ls.l[ls.pos]) != "" {
		readLayout(ls, res, v211)
	}
	if len(beds) > 10000 {
		res.note(0, "Only the first 10,000 of %d beds were imported", len(beds))
		beds = beds[:10000]
	}
	for i, rb := range beds {
		res.addBed(i+1, rb, 100) // SedLog keeps thicknesses in cm
	}
	res.finish()
	return res, nil
}

// ParseTemplate reads a SedLog template (.tem): a layout without beds.
func ParseTemplate(data []byte) (res *Result, err error) {
	ls := newLines(data)
	defer catch(&err, ls)
	head := strings.TrimSpace(ls.next())
	if !strings.HasPrefix(head, "SEDLOG-TEMPLATE-VERSION-") {
		return nil, fmt.Errorf("this is not a SedLog template (it starts with %q)", clip(head, 40))
	}
	res = newResult("tem")
	readLayout(ls, res, true)
	res.finish()
	return res, nil
}

func catch(err *error, ls *lines) {
	if r := recover(); r != nil {
		e, ok := r.(error)
		if !ok {
			panic(r)
		}
		if errors.Is(e, errShort) {
			e = fmt.Errorf("the file ends too early (after line %d); it may be cut short or from a newer SedLog", ls.pos)
		}
		*err = e
	}
}

func readBed(ls *lines, v211 bool) rawBed {
	var b rawBed
	b.thickness = ls.next()
	b.boundary = ls.next()
	for i := range 3 {
		b.lith[i] = ls.next()
		b.share[i] = ls.float()
	}
	b.grainBase, b.phiBase = ls.next(), ls.next()
	b.grainTop, b.phiTop = ls.next(), ls.next()
	b.inBed = ls.next()
	if v211 {
		ls.next() // 1 = in-bed symbol drawn once instead of tiled
	}
	k := ls.int()
	if k < 0 || k > 1000 {
		panic(fmt.Errorf("line %d: %d symbols in one bed is not sensible", ls.pos, k))
	}
	for range k {
		b.symbols = append(b.symbols, ls.next())
	}
	b.notes = ls.next()
	b.bioType = ls.next()
	b.bioStrength = ls.next()
	b.palaeocurrents = ls.next()
	b.facies = ls.next()
	for i := range 3 {
		b.other[i] = ls.next()
		ls.next() // other column: symbol
		if v211 {
			ls.next() // other column: base line
		}
	}
	b.formation = ls.next()
	b.age = ls.next()
	ls.next() // formation boundary line: 0 none, 1 simple, 2 dotted
	ls.next() // age boundary line
	if v211 {
		ls.next() // 1 = in-bed symbol drawn over lithology 1
	}
	return b
}

// readLayout reads the settings after the beds (or the whole of a template).
func readLayout(ls *lines, res *Result, v211 bool) {
	c := &res.Content
	ls.float() // header height
	ls.float() // lithology pattern size
	ls.float() // symbol size
	pxPerCm := ls.float()
	if title := strings.TrimSpace(ls.next()); title != "" && title != "Untitled" {
		res.Name = clip(title, 200)
	}
	ls.float() // font sizes
	ls.float()
	ls.int() // scale numbering: start
	ls.int() // and step
	switch ls.int() {
	case 0:
		c.Palaeocurrent = "numbers"
	default: // 1 numbers and arrows, 2 arrows
		c.Palaeocurrent = "arrows"
	}
	ls.int() // symbols wrap
	if v211 {
		if ls.int() == 1 {
			res.note(0, "The log was measured from the top (core depth); GoLog draws it from the base up")
		}
		ls.int() // text alignment
	}
	used := ls.int()

	// Thirteen columns in SedLog's fixed order, each with its title and width; the number with
	// entry k is the column shown at position k, and the first `used` positions are shown.
	type col struct {
		title string
		width float64
	}
	var cols [13]col
	var show [13]int
	for k := range 13 {
		cols[k].title = strings.TrimSpace(ls.next())
		show[k] = ls.int()
		cols[k].width = ls.float()
	}
	for k := range 3 {
		if t := cols[10+k].title; t != "" && !defaultTitles[strings.ToUpper(t)] {
			res.others[k] = t
		}
	}

	if pxPerCm > 0 {
		c.Scale = nearestScale(pxPerCmAt1to1 / pxPerCm)
	}
	if strings.Contains(strings.ToUpper(cols[2].title), "(M)") {
		c.Units = "m"
	} else {
		c.Units = "cm"
	}

	// GoLog's columns: the shown SedLog columns first, in their order, then the rest hidden.
	byKind := map[string]int{}
	for i, gc := range c.Columns {
		byKind[gc.Kind] = i
	}
	var order []int
	seen := map[int]bool{}
	for k := 0; k < used && k < 13; k++ {
		m := show[k]
		if m < 0 || m >= 13 || seen[m] {
			continue
		}
		seen[m] = true
		order = append(order, m)
	}
	newCols := c.Columns[:0:0]
	taken := map[string]bool{}
	for _, m := range order {
		kind := columnKinds[m]
		i, ok := byKind[kind]
		if !ok {
			if strings.HasPrefix(kind, "other") {
				res.note(0, "GoLog has no free columns yet: the %s column is not shown", res.others[m-10])
			}
			continue
		}
		gc := c.Columns[i]
		gc.Visible = true
		if t := cols[m].title; t != "" && !defaultTitles[strings.ToUpper(t)] {
			gc.Title = clip(t, 60)
		}
		if w := cols[m].width; w > 0 {
			// SedLog widths are in tenths of a millimetre on paper.
			gc.Width = math.Max(16, math.Min(800, math.Round(w/10*96/25.4*10)/10))
		}
		newCols = append(newCols, gc)
		taken[kind] = true
	}
	for _, gc := range c.Columns {
		if !taken[gc.Kind] {
			gc.Visible = false
			newCols = append(newCols, gc)
		}
	}
	c.Columns = newCols

	// 2.1.1 ends with two more sizes and the grain-size scale (0 both, 1 limestones, 2 clastics).
	for ls.pos < len(ls.l) && strings.TrimSpace(ls.l[ls.pos]) != "" {
		ls.next()
	}
}

func nearestScale(s float64) int {
	best := 50
	for _, v := range []int{10, 20, 25, 50, 100, 200, 250, 500, 1000} {
		if math.Abs(math.Log(float64(v)/s)) < math.Abs(math.Log(float64(best)/s)) {
			best = v
		}
	}
	return best
}
