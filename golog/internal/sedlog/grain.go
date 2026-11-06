package sedlog

import (
	"strconv"
	"strings"
)

// Grain-size names SedLog writes, with phi values for when a file gives only the name.
// Clastic classes follow GoLog's own scale; "a/b" classes sit halfway between.
var grainPhi = map[string]float64{
	"clay": 9, "silt": 6, "vf": 3.5, "f": 2.5, "m": 1.5, "c": 0.5, "vc": -0.5,
	"granule": -1.5, "pebble": -4, "cobble": -7, "boulder": -9,
	// Carbonate (Dunham) textures are drawn at the width of a clastic class.
	"mudstone": 6, "wackestone": 3, "packstone": 1.5, "grainstone": -0.5,
	"rudstone fine": -1.5, "rudstone medium": -3, "rudstone": -4,
}

// carbonateTextures are the names SedLog uses on its limestone scale.
var carbonateTextures = map[string]bool{
	"mudstone": true, "wackestone": true, "packstone": true, "grainstone": true,
	"rudstone fine": true, "rudstone medium": true, "rudstone": true,
}

// grain turns a SedLog grain-size name and value into phi; nil if not recorded.
func grain(name, value string) *float64 {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" || n == "<none>" {
		return nil
	}
	if v, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err == nil && v >= -12 && v <= 14 {
		return &v
	}
	if v, ok := grainPhi[n]; ok {
		return &v
	}
	if a, b, ok := strings.Cut(n, "/"); ok {
		va, oka := grainPhi[a]
		vb, okb := grainPhi[b]
		if oka && okb {
			v := (va + vb) / 2
			return &v
		}
	}
	return nil
}
