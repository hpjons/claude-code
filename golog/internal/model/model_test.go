package model

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestStageOneLogGetsDefaults(t *testing.T) {
	var c LogContent
	if err := json.Unmarshal([]byte(`{"units":"cm","beds":[]}`), &c); err != nil {
		t.Fatal(err)
	}
	c.Normalize()
	if c.Units != "cm" || c.Scale != 50 || c.Paper != "A4" || c.Palaeocurrent != "arrows" || c.Fill != "colour" {
		t.Fatalf("defaults: %+v", c)
	}
	if len(c.Columns) != len(defaultColumns) {
		t.Fatalf("got %d columns, want %d", len(c.Columns), len(defaultColumns))
	}
}

func TestNormalizeKeepsUserOrder(t *testing.T) {
	c := LogContent{Columns: []Column{{Kind: "notes", Title: "Remarks", Width: 120, Visible: true}}}
	c.Normalize()
	if c.Columns[0].Kind != "notes" || c.Columns[0].Title != "Remarks" {
		t.Fatalf("user column moved or changed: %+v", c.Columns[0])
	}
}

func TestValidateNewFields(t *testing.T) {
	bi := func(n int) *int { return &n }
	bed := func(mod func(*Bed)) LogContent {
		b := Bed{ID: "a", Thickness: 1}
		mod(&b)
		return LogContent{Beds: []Bed{b}}
	}
	ok := bed(func(b *Bed) {
		b.Symbols = []string{"ripples", "shell"}
		b.Bioturbation = bi(3)
		b.Palaeocurrents = []float64{0, 85.5, 360}
		b.Formation = "Hellisheiði"
		b.Lithology = []Lithology{{Pattern: "sandstone", Percent: 60, Colour: "#c0392b"}, {Pattern: "mudstone", Percent: 40}}
	})
	if err := ok.Validate(); err != nil {
		t.Fatalf("valid bed refused: %v", err)
	}
	for name, c := range map[string]LogContent{
		"bioturbation 7":    bed(func(b *Bed) { b.Bioturbation = bi(7) }),
		"palaeocurrent 400": bed(func(b *Bed) { b.Palaeocurrents = []float64{400} }),
		"empty symbol":      bed(func(b *Bed) { b.Symbols = []string{""} }),
		"long formation":    bed(func(b *Bed) { b.Formation = strings.Repeat("x", 201) }),
		"odd scale":         {Scale: 33},
		"unknown paper":     {Paper: "A0"},
		"unknown fill":      {Fill: "rainbow"},
		"bad colour":        bed(func(b *Bed) { b.Lithology = []Lithology{{Pattern: "sandstone", Percent: 100, Colour: "red"}} }),
		"repeated column":   {Columns: []Column{{Kind: "notes", Width: 50}, {Kind: "notes", Width: 50}}},
		"unknown column":    {Columns: []Column{{Kind: "colour", Width: 50}}},
		"too narrow column": {Columns: []Column{{Kind: "notes", Width: 2}}},
	} {
		if err := c.Validate(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}
