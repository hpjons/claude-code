package sedlog

import "strings"

// SedLog's names for rocks and symbols, matched to GoLog's own patterns and symbols.
// The names come from SedLog's plain-text lists and from saved files; the drawings are GoLog's.

var lithologies = map[string]string{
	"mudstone":                      "mudstone",
	"claystone":                     "claystone",
	"shale":                         "shale",
	"siltstone":                     "siltstone",
	"sandstone":                     "sandstone",
	"conglomerate":                  "conglomerate",
	"coal":                          "coal",
	"limestone":                     "limestone",
	"chert":                         "chert",
	"volcaniclastic":                "volcaniclastic",
	"lime mudstone":                 "lime-mudstone",
	"wackestone":                    "wackestone",
	"packstone":                     "packstone",
	"grainstone":                    "grainstone",
	"halite":                        "halite",
	"gypsum/anhydrite":              "gypsum",
	"dolomite":                      "dolomite",
	"breccia":                       "breccia",
	"matrix-supported conglomerate": "matrix-conglomerate",
	"clast-supported conglomerate":  "conglomerate",
	"lava":                          "lava",
	"fine ash":                      "tuff",
	"coarse ash":                    "coarse-ash",
	// GoLog's own names, as GoLog writes them in its CSV export
	"diamictite":                 "diamictite",
	"chalk":                      "chalk",
	"marl":                       "marl",
	"tuff":                       "tuff",
	"lava (basalt)":              "lava",
	"hyaloclastite":              "hyaloclastite",
	"evaporite (gypsum, halite)": "evaporite",
	"soil / palaeosol":           "soil",
	"volcaniclastic sediment":    "volcaniclastic",
	"coarse ash / lapilli tuff":  "coarse-ash",
	"gypsum / anhydrite":         "gypsum",
}

// Words that identify a rock in a name SedLog users made up themselves ("Tuffaceous sandstone").
var lithologyWords = []struct{ word, id string }{
	{"conglomerate", "conglomerate"}, {"breccia", "breccia"}, {"sand", "sandstone"}, {"silt", "siltstone"},
	{"clay", "claystone"}, {"shale", "shale"}, {"mud", "mudstone"}, {"marl", "marl"}, {"chalk", "chalk"},
	{"dolo", "dolomite"}, {"lime", "limestone"}, {"coal", "coal"}, {"lignite", "coal"}, {"chert", "chert"},
	{"tuff", "tuff"}, {"ash", "tuff"}, {"lava", "lava"}, {"basalt", "lava"}, {"hyalo", "hyaloclastite"},
	{"gypsum", "gypsum"}, {"anhydrite", "gypsum"}, {"halite", "halite"}, {"salt", "halite"},
	{"soil", "soil"}, {"diamict", "diamictite"}, {"till", "diamictite"},
}

// lithology returns GoLog's pattern for a SedLog lithology name and whether it was an exact match.
func lithology(name string) (id string, exact bool) {
	n := strings.ToLower(strings.TrimSpace(name))
	if id, ok := lithologies[n]; ok {
		return id, true
	}
	for _, w := range lithologyWords {
		if strings.Contains(n, w.word) {
			return w.id, false
		}
	}
	return "", false
}

var symbols = map[string]string{
	// fossils
	"shells":          "shell-debris",
	"bivalves":        "bivalve",
	"gastropods":      "gastropod",
	"cephalopods":     "ammonite",
	"brachiopods":     "brachiopod",
	"echinoids":       "echinoid",
	"crinoids":        "crinoid",
	"solitary corals": "coral",
	"colonial corals": "coral",
	"foraminifera":    "foraminifera",
	"algae":           "algae",
	"bryzoa":          "bryozoa", // SedLog's spelling
	"bryozoa":         "bryozoa",
	"stromatolites":   "stromatolite",
	"vertebrates":     "bone",
	"plant material":  "plant",
	"roots":           "roots",
	"logs":            "wood",
	"tree stumps":     "tree-stump",
	"ostracods":       "ostracod",
	"radiolaria":      "radiolaria",
	"sponges":         "sponge",
	// other
	"nodules and concretions": "nodules",
	"intraclasts":             "intraclasts",
	"mudclasts":               "rip-up-clasts",
	"flute marks":             "flute-casts",
	"groove marks":            "groove-casts",
	"scours":                  "scours",
	// sedimentary structures
	"current ripple cross-lamination": "current-ripples",
	"wave ripple cross-lamination":    "wave-ripples",
	"planar cross bedding":            "planar-cross-bedding",
	"trough cross bedding":            "trough-cross-bedding",
	"horizontal planar lamination":    "planar-lamination",
	"hummocky cross stratification":   "hummocky",
	"swaley cross stratification":     "swaley",
	"mudcracks":                       "mudcracks",
	"synaeresis cracks":               "synaeresis-cracks",
	"convolute lamination":            "convolute",
	"load casts":                      "load-casts",
	"water structures":                "water-escape",
	"herring-bone cross bedding":      "herringbone",
	// trace fossils
	"minor bioturbation":    "bioturbation",
	"moderate bioturbation": "bioturbation",
	"intense bioturbation":  "bioturbation",
	"tracks":                "footprints",
	"trails":                "trails",
	"vertical burrows":      "vertical-burrows",
	"horizontal burrrows":   "horizontal-burrows", // SedLog's spelling
	"horizontal burrows":    "horizontal-burrows",
	// GoLog's own names, as GoLog writes them in its CSV export
	"low-angle lamination":          "low-angle-lamination",
	"climbing ripples":              "climbing-ripples",
	"flaser bedding":                "flaser",
	"normal grading":                "graded",
	"reverse grading":               "reverse-graded",
	"massive (structureless)":       "massive",
	"clast imbrication":             "imbrication",
	"columnar jointing":             "columnar-joints",
	"pillow structures":             "pillows",
	"vesicles":                      "vesicles",
	"fish remains":                  "fish",
	"leaves":                        "leaf",
	"u-shaped burrows":              "u-burrows",
	"branching burrow systems":      "branching-burrows",
	"footprints":                    "footprints",
	"pumice":                        "pumice",
	"scoria":                        "scoria",
	"volcanic glass":                "glass",
	"dropstone":                     "dropstone",
	"sample taken":                  "sample",
	"photograph":                    "photo",
	"planar lamination":             "planar-lamination",
	"planar cross-bedding":          "planar-cross-bedding",
	"trough cross-bedding":          "trough-cross-bedding",
	"current ripple lamination":     "current-ripples",
	"wave ripple lamination":        "wave-ripples",
	"hummocky cross-stratification": "hummocky",
	"swaley cross-stratification":   "swaley",
	"herringbone cross-bedding":     "herringbone",
	"desiccation cracks":            "mudcracks",
	"groove casts":                  "groove-casts",
	"flute casts":                   "flute-casts",
	"water-escape structures":       "water-escape",
	"mud rip-up clasts":             "rip-up-clasts",
	"concretions / nodules":         "nodules",
	"corals":                        "coral",
	"crinoid ossicles":              "crinoid",
	"ammonites":                     "ammonite",
	"bone fragments":                "bone",
	"plant fragments":               "plant",
	"wood / logs":                   "wood",
	"shell fragments":               "shell-debris",
	"trails and tracks":             "trails",
	"bioturbation (general)":        "bioturbation",
}

// Structures GoLog can draw repeated inside the grain-size profile.
var inBedSymbols = map[string]bool{
	"planar-lamination": true, "low-angle-lamination": true, "planar-cross-bedding": true, "trough-cross-bedding": true,
	"current-ripples": true, "wave-ripples": true, "climbing-ripples": true, "hummocky": true, "swaley": true,
	"herringbone": true, "flaser": true, "convolute": true,
}

func symbol(name string) (string, bool) {
	id, ok := symbols[strings.ToLower(strings.TrimSpace(name))]
	return id, ok
}

// SedLog's column names (in its fixed column order) and GoLog's column kinds.
var columnKinds = []string{
	"age", "formation", "scale", "lithology", "grain", "symbols", "notes", "bioturbation", "palaeocurrent", "facies",
	"other1", "other2", "other3",
}

// SedLog's own titles; a column with one of these keeps GoLog's English title instead.
var defaultTitles = map[string]bool{
	"AGE": true, "FORMATION": true, "SCALE (CM)": true, "SCALE (M)": true, "LITHOLOGY": true, "GRAIN SIZE": true,
	"STRUCTURES / FOSSILS": true, "NOTES": true, "BIOTURBATION": true, "PALAEOCURRENT": true, "FACIES": true,
	"OTHER1": true, "OTHER2": true, "OTHER3": true,
}

func contact(s string) string {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sharp":
		return "sharp"
	case "erosion", "erosional":
		return "erosional"
	case "gradational":
		return "gradational"
	}
	return "unknown"
}
