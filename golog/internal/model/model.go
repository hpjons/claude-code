// Package model defines GoLog's data: projects, logs and beds, and checks them.
package model

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"
	"time"
)

// Limits that keep a log sensible and the database small.
const (
	MaxNameLen    = 200
	MaxNotesLen   = 5000
	MaxBeds       = 10000
	MaxLithology  = 3
	MaxThicknessM = 10000 // 10 km of section in one bed is certainly a typo
	MinPhi        = -12
	MaxPhi        = 14
)

// Contact types at the base of a bed.
const (
	ContactUnknown     = "unknown"
	ContactSharp       = "sharp"
	ContactErosional   = "erosional"
	ContactGradational = "gradational"
)

var contacts = map[string]bool{ContactUnknown: true, ContactSharp: true, ContactErosional: true, ContactGradational: true}

// Project groups the logs from one field area or exercise.
type Project struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Created time.Time    `json:"created"`
	Updated time.Time    `json:"updated"`
	Logs    []LogSummary `json:"logs"`
}

// ProjectSummary is a project as shown in the project list.
type ProjectSummary struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Updated  time.Time `json:"updated"`
	LogCount int       `json:"logCount"`
}

// LogSummary is a log as shown in a project's list of logs.
type LogSummary struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Updated  time.Time `json:"updated"`
	BedCount int       `json:"bedCount"`
}

// Log is one measured section. Beds are listed from the bottom up.
type Log struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"projectId"`
	Name      string    `json:"name"`
	Revision  int64     `json:"revision"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	LogContent
}

// LogContent is the part of a log the user edits; it is stored as one JSON document.
type LogContent struct {
	Units         string `json:"units"`         // "m" or "cm": how thicknesses are shown; always stored in metres
	Scale         int    `json:"scale"`         // print scale 1:Scale
	Paper         string `json:"paper"`         // A4, A3 or Letter (portrait)
	Palaeocurrent string `json:"palaeocurrent"` // "arrows" or "numbers"
	Fill          string `json:"fill"`          // "colour" or "white": how lithologies are filled
	// The grain-size scale above the grain-size column: the class names (clay … boulder) and the
	// group names (mud, sand, gravel) are shown unless hidden, and so is the column's title.
	HideGrainClasses bool     `json:"hideGrainClasses,omitempty"`
	HideGrainGroups  bool     `json:"hideGrainGroups,omitempty"`
	HideGrainTitle   bool     `json:"hideGrainTitle,omitempty"`
	Columns          []Column `json:"columns"` // left to right
	Beds             []Bed    `json:"beds"`
}

// Column is one column of the drawn log.
type Column struct {
	Kind    string  `json:"kind"`
	Title   string  `json:"title"`
	Width   float64 `json:"width"` // px at 100 %
	Visible bool    `json:"visible"`
}

// Bed is one layer of the section.
type Bed struct {
	ID             string      `json:"id"`
	Thickness      float64     `json:"thickness"` // metres
	Contact        string      `json:"contact"`   // at the base of the bed
	Lithology      []Lithology `json:"lithology"`
	Grain          Grain       `json:"grain"`
	InBed          string      `json:"inBed,omitempty"`        // structure drawn inside the grain-size profile
	Symbols        []string    `json:"symbols,omitempty"`      // structures, fossils and traces beside the bed
	Bioturbation   *int        `json:"bioturbation,omitempty"` // bioturbation index 0–6
	Palaeocurrents []float64   `json:"palaeocurrents,omitempty"`
	Facies         string      `json:"facies,omitempty"`
	Formation      string      `json:"formation,omitempty"` // a formation starts at this bed
	Age            string      `json:"age,omitempty"`       // an age unit starts at this bed
	Notes          string      `json:"notes,omitempty"`
}

// Column kinds, and the default layout of a new log.
var defaultColumns = []Column{
	{Kind: "age", Title: "Age", Width: 34, Visible: false},
	{Kind: "formation", Title: "Formation", Width: 34, Visible: false},
	{Kind: "scale", Title: "Height", Width: 46, Visible: true},
	{Kind: "lithology", Title: "Lithology", Width: 64, Visible: true},
	{Kind: "grain", Title: "Grain size", Width: 210, Visible: true},
	{Kind: "symbols", Title: "Structures & fossils", Width: 90, Visible: true},
	{Kind: "bioturbation", Title: "Bioturbation", Width: 44, Visible: false},
	{Kind: "palaeocurrent", Title: "Palaeocurrent", Width: 60, Visible: false},
	{Kind: "facies", Title: "Facies", Width: 44, Visible: false},
	{Kind: "notes", Title: "Notes", Width: 200, Visible: true},
}

// Scales a log can be drawn at (1:n).
var Scales = []int{10, 20, 25, 50, 100, 200, 250, 500, 1000}

var hexColour = regexp.MustCompile(`^#[0-9a-f]{6}$`)

var papers = map[string]bool{"A4": true, "A3": true, "Letter": true}

// Lithology is a rock type making up part of a bed.
type Lithology struct {
	Pattern string  `json:"pattern"`
	Percent float64 `json:"percent"`
	Colour  string  `json:"colour,omitempty"` // "#rrggbb"; empty = the pattern's own colour
}

// Grain holds grain size at the base and top of a bed as phi values (nil = not recorded).
type Grain struct {
	Base *float64 `json:"base,omitempty"`
	Top  *float64 `json:"top,omitempty"`
}

// NewLogContent is the content of a freshly created log.
func NewLogContent() LogContent {
	c := LogContent{}
	c.Normalize()
	return c
}

// Normalize fills in settings that are missing, e.g. in logs saved by an earlier GoLog.
func (c *LogContent) Normalize() {
	if c.Units == "" {
		c.Units = "m"
	}
	if c.Scale == 0 {
		c.Scale = 50
	}
	if c.Paper == "" {
		c.Paper = "A4"
	}
	if c.Palaeocurrent == "" {
		c.Palaeocurrent = "arrows"
	}
	if c.Fill == "" {
		c.Fill = "colour"
	}
	if c.Beds == nil {
		c.Beds = []Bed{}
	}
	// Keep the user's columns in their order and add any kinds they don't have yet.
	have := map[string]bool{}
	for _, col := range c.Columns {
		have[col.Kind] = true
	}
	for _, d := range defaultColumns {
		if !have[d.Kind] {
			c.Columns = append(c.Columns, d)
		}
	}
}

// CleanName trims a name and checks its length.
func CleanName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("name is empty")
	}
	if len([]rune(name)) > MaxNameLen {
		return "", fmt.Errorf("name is longer than %d characters", MaxNameLen)
	}
	return name, nil
}

// Validate checks log content and fills in defaults. Errors name the bed (counted from the base).
func (c *LogContent) Validate() error {
	c.Normalize()
	if c.Units != "m" && c.Units != "cm" {
		return fmt.Errorf("units must be m or cm")
	}
	if !slices.Contains(Scales, c.Scale) {
		return fmt.Errorf("scale 1:%d is not one of the offered scales", c.Scale)
	}
	if !papers[c.Paper] {
		return fmt.Errorf("unknown paper size %q", c.Paper)
	}
	if c.Palaeocurrent != "arrows" && c.Palaeocurrent != "numbers" {
		return fmt.Errorf("palaeocurrents are shown as arrows or numbers")
	}
	if c.Fill != "colour" && c.Fill != "white" {
		return fmt.Errorf("lithologies are filled with colour or white")
	}
	kinds := map[string]bool{}
	for _, d := range defaultColumns {
		kinds[d.Kind] = true
	}
	seen := map[string]bool{}
	for i := range c.Columns {
		col := &c.Columns[i]
		if !kinds[col.Kind] || seen[col.Kind] {
			return fmt.Errorf("unknown or repeated column %q", col.Kind)
		}
		seen[col.Kind] = true
		col.Title = strings.TrimSpace(col.Title)
		if len([]rune(col.Title)) > 60 {
			return fmt.Errorf("column title %q is longer than 60 characters", col.Title)
		}
		if math.IsNaN(col.Width) || col.Width < 16 || col.Width > 800 {
			return fmt.Errorf("column %q must be 16–800 px wide", col.Title)
		}
	}
	if len(c.Beds) > MaxBeds {
		return fmt.Errorf("a log can have at most %d beds", MaxBeds)
	}
	ids := map[string]bool{}
	for i := range c.Beds {
		b := &c.Beds[i]
		if err := b.validate(); err != nil {
			return fmt.Errorf("bed %d: %w", i+1, err)
		}
		if b.ID == "" || ids[b.ID] {
			return fmt.Errorf("bed %d: missing or repeated id", i+1)
		}
		ids[b.ID] = true
	}
	return nil
}

func (b *Bed) validate() error {
	if math.IsNaN(b.Thickness) || b.Thickness <= 0 || b.Thickness > MaxThicknessM {
		return fmt.Errorf("thickness must be more than 0 and at most %d m", MaxThicknessM)
	}
	if b.Contact == "" {
		b.Contact = ContactUnknown
	}
	if !contacts[b.Contact] {
		return fmt.Errorf("unknown contact %q", b.Contact)
	}
	if b.Lithology == nil {
		b.Lithology = []Lithology{}
	}
	if len(b.Lithology) > MaxLithology {
		return fmt.Errorf("at most %d lithologies", MaxLithology)
	}
	total := 0.0
	for _, l := range b.Lithology {
		if l.Pattern == "" || len(l.Pattern) > 64 {
			return fmt.Errorf("lithology pattern name missing or too long")
		}
		if math.IsNaN(l.Percent) || l.Percent <= 0 || l.Percent > 100 {
			return fmt.Errorf("lithology percent must be between 0 and 100")
		}
		if l.Colour != "" && !hexColour.MatchString(l.Colour) {
			return fmt.Errorf("lithology colour %q is not of the form #rrggbb", l.Colour)
		}
		total += l.Percent
	}
	if total > 100.0001 {
		return fmt.Errorf("lithology percentages add up to more than 100")
	}
	for _, p := range []*float64{b.Grain.Base, b.Grain.Top} {
		if p != nil && (math.IsNaN(*p) || *p < MinPhi || *p > MaxPhi) {
			return fmt.Errorf("grain size must be between %d and %d phi", MinPhi, MaxPhi)
		}
	}
	if len([]rune(b.Notes)) > MaxNotesLen {
		return fmt.Errorf("notes are longer than %d characters", MaxNotesLen)
	}
	if len(b.InBed) > 64 {
		return fmt.Errorf("in-bed structure name too long")
	}
	if len(b.Symbols) > 40 {
		return fmt.Errorf("at most 40 symbols per bed")
	}
	for _, sym := range b.Symbols {
		if sym == "" || len(sym) > 64 {
			return fmt.Errorf("symbol name missing or too long")
		}
	}
	if b.Bioturbation != nil && (*b.Bioturbation < 0 || *b.Bioturbation > 6) {
		return fmt.Errorf("bioturbation index must be 0–6")
	}
	if len(b.Palaeocurrents) > 100 {
		return fmt.Errorf("at most 100 palaeocurrent readings per bed")
	}
	for _, a := range b.Palaeocurrents {
		if math.IsNaN(a) || a < 0 || a > 360 {
			return fmt.Errorf("palaeocurrent directions must be 0–360°")
		}
	}
	for name, v := range map[string]string{"facies": b.Facies, "formation": b.Formation, "age": b.Age} {
		if len([]rune(v)) > MaxNameLen {
			return fmt.Errorf("%s is longer than %d characters", name, MaxNameLen)
		}
	}
	return nil
}
