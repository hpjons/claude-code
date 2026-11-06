package stats

import (
	"hash/fnv"
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"

	"golog/internal/model"
)

// DefaultShuffles is how many shuffled sections the order tests compare with (Burgess uses 5000).
const DefaultShuffles = 5000

// Result is everything the statistics window shows.
type Result struct {
	Summary Summary  `json:"summary"`
	Facies  *Facies  `json:"facies,omitempty"` // nil when fewer than two facies are recorded
	Runs    *Runs    `json:"runs,omitempty"`   // nil with fewer than three beds
	Notes   []string `json:"notes,omitempty"`  // why a test was left out, and similar
}

// Summary describes the beds of a log.
type Summary struct {
	Beds        int     `json:"beds"`
	Thickness   float64 `json:"thickness"` // metres
	MeanBed     float64 `json:"meanBed"`
	MedianBed   float64 `json:"medianBed"`
	MinBed      float64 `json:"minBed"`
	MaxBed      float64 `json:"maxBed"`
	SDBed       float64 `json:"sdBed"` // standard deviation of bed thickness
	Lithologies []Share `json:"lithologies"`
	Facies      []Share `json:"facies"`
	Grain       []Share `json:"grain"`    // mud, sand, gravel, not recorded
	Trends      []Share `json:"trends"`   // fining up, coarsening up, no change, not recorded
	Contacts    []Share `json:"contacts"` // at the base of beds
}

// Share is the part of the section that has one property.
type Share struct {
	Key       string  `json:"key"`
	Beds      int     `json:"beds"`
	Thickness float64 `json:"thickness"` // metres
	Percent   float64 `json:"percent"`   // of the whole section's thickness
}

// Facies is the facies transition analysis.
type Facies struct {
	Codes       []string    `json:"codes"`       // facies in matrix order
	Beds        []int       `json:"beds"`        // beds of each facies
	Counts      [][]int     `json:"counts"`      // transitions from row to column
	Prob        [][]float64 `json:"prob"`        // transition probabilities; each row adds up to 1
	Transitions int         `json:"transitions"` // number of transitions counted
	M           Test        `json:"m"`
}

// Runs is the runs analysis of bed thickness.
type Runs struct {
	Thickening []int `json:"thickening"` // place of each bed in a thickening-upward run (0 = none)
	Thinning   []int `json:"thinning"`
	R          Test  `json:"r"`
}

// Analyse calculates the statistics for a log. The shuffles are drawn from a random sequence
// seeded by the beds themselves, so the same log always gives the same numbers.
func Analyse(c model.LogContent, shuffles int) Result {
	beds := c.Beds
	res := Result{Summary: summarise(beds)}
	if shuffles <= 0 {
		shuffles = DefaultShuffles
	}
	codes, seq := faciesNumbers(beds)
	thick := make([]float64, len(beds))
	for i, b := range beds {
		thick[i] = b.Thickness
	}
	rng := rand.New(rand.NewPCG(seed(beds), 0x601065))

	doFacies := len(codes) >= 2
	doRuns := len(beds) >= 3
	if !doFacies {
		if len(codes) == 1 {
			res.Notes = append(res.Notes, "The facies test needs at least two different facies; only one is recorded.")
		} else {
			res.Notes = append(res.Notes, "The facies test needs facies. Turn on the Facies column in the Layout tab, then give each bed a facies code (for example 1, 2, 3 or St, Sh, Fm) in the bed panel.")
		}
	}
	if !doRuns {
		res.Notes = append(res.Notes, "The thickness test needs at least three beds.")
	}
	if !doFacies && !doRuns {
		return res
	}

	var mShuffled, rShuffled []float64
	f, t := slices.Clone(seq), slices.Clone(thick)
	for range shuffles {
		copy(f, seq)
		copy(t, thick)
		shuffle(rng, f, t)
		if doFacies {
			counts, _, _ := markov(f, len(codes))
			mShuffled = append(mShuffled, metricM(counts))
		}
		if doRuns {
			rShuffled = append(rShuffled, metricR(t))
		}
	}

	if doFacies {
		counts, prob, m := markov(seq, len(codes))
		fa := &Facies{Codes: codes, Beds: make([]int, len(codes)), Counts: counts, Prob: prob, M: newTest(m, mShuffled, 0.02)}
		for _, x := range seq {
			if x >= 0 {
				fa.Beds[x]++
			}
		}
		for _, row := range counts {
			for _, n := range row {
				fa.Transitions += n
			}
		}
		if fa.Transitions == 0 {
			res.Notes = append(res.Notes, "No bed is followed by a bed of another facies, so there are no transitions to test.")
			fa = nil
		}
		res.Facies = fa
	}
	if doRuns {
		ru := &Runs{Thickening: make([]int, len(thick)), Thinning: make([]int, len(thick)), R: newTest(metricR(thick), rShuffled, 0.05)}
		for k := 1; k < len(thick); k++ {
			switch {
			case thick[k] > thick[k-1]:
				ru.Thickening[k] = ru.Thickening[k-1] + 1
			case thick[k] < thick[k-1]:
				ru.Thinning[k] = ru.Thinning[k-1] + 1
			}
		}
		res.Runs = ru
	}
	return res
}

// faciesNumbers numbers the facies 0…F-1 and lists each bed's number (−1 for none). Facies
// codes that are all whole numbers keep their numeric order; otherwise the order is that in
// which each facies first appears from the base. Gaps are closed up (SedLog keeps empty rows for
// unused numbers, which lowers m; Burgess numbers facies consecutively).
func faciesNumbers(beds []model.Bed) (codes []string, seq []int) {
	seen := map[string]bool{}
	numeric := true
	for _, b := range beds {
		code := strings.TrimSpace(b.Facies)
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		codes = append(codes, code)
		if _, err := strconv.Atoi(code); err != nil {
			numeric = false
		}
	}
	if numeric {
		slices.SortStableFunc(codes, func(a, b string) int {
			x, _ := strconv.Atoi(a)
			y, _ := strconv.Atoi(b)
			return x - y
		})
	}
	index := map[string]int{}
	for i, c := range codes {
		index[c] = i
	}
	seq = make([]int, len(beds))
	for i, b := range beds {
		if x, ok := index[strings.TrimSpace(b.Facies)]; ok {
			seq[i] = x
		} else {
			seq[i] = -1
		}
	}
	return codes, seq
}

func seed(beds []model.Bed) uint64 {
	h := fnv.New64a()
	for _, b := range beds {
		h.Write([]byte(strconv.FormatFloat(b.Thickness, 'g', -1, 64)))
		h.Write([]byte{0})
		h.Write([]byte(b.Facies))
		h.Write([]byte{0})
	}
	return h.Sum64()
}

func summarise(beds []model.Bed) Summary {
	s := Summary{Beds: len(beds)}
	if len(beds) == 0 {
		return s
	}
	th := make([]float64, len(beds))
	s.MinBed = math.Inf(1)
	for i, b := range beds {
		th[i] = b.Thickness
		s.Thickness += b.Thickness
		s.MinBed = min(s.MinBed, b.Thickness)
		s.MaxBed = max(s.MaxBed, b.Thickness)
	}
	s.MeanBed = s.Thickness / float64(len(beds))
	for _, t := range th {
		s.SDBed += (t - s.MeanBed) * (t - s.MeanBed)
	}
	if len(beds) > 1 {
		s.SDBed = math.Sqrt(s.SDBed / float64(len(beds)-1))
	} else {
		s.SDBed = 0
	}
	s.MedianBed = median(th)

	lith, fac, grain, trend, contact := tally{}, tally{}, tally{}, tally{}, tally{}
	for _, b := range beds {
		total := 0.0
		for _, l := range b.Lithology {
			total += l.Percent
		}
		counted := map[string]bool{}
		for _, l := range b.Lithology {
			if l.Pattern == "" || total <= 0 {
				continue
			}
			lith.add(l.Pattern, b.Thickness*l.Percent/total, !counted[l.Pattern])
			counted[l.Pattern] = true
		}
		if total <= 0 {
			lith.add("", b.Thickness, true)
		}
		fac.add(strings.TrimSpace(b.Facies), b.Thickness, true)
		grain.add(grainGroup(b.Grain), b.Thickness, true)
		trend.add(grainTrend(b.Grain), b.Thickness, true)
		contact.add(b.Contact, b.Thickness, true)
	}
	s.Lithologies = lith.shares(s.Thickness, nil)
	codes, _ := faciesNumbers(beds)
	s.Facies = fac.shares(s.Thickness, codes)
	s.Grain = grain.shares(s.Thickness, []string{"gravel", "sand", "mud", ""})
	s.Trends = trend.shares(s.Thickness, []string{"fining", "coarsening", "constant", ""})
	s.Contacts = contact.shares(s.Thickness, []string{model.ContactErosional, model.ContactSharp, model.ContactGradational, model.ContactUnknown})
	return s
}

type tally map[string]*Share

func (t tally) add(key string, thickness float64, newBed bool) {
	s := t[key]
	if s == nil {
		s = &Share{Key: key}
		t[key] = s
	}
	s.Thickness += thickness
	if newBed {
		s.Beds++
	}
}

// shares lists the tally in the given order, or thickest first; the empty key ("not recorded")
// always comes last.
func (t tally) shares(total float64, order []string) []Share {
	var out []Share
	for _, s := range t {
		if total > 0 {
			s.Percent = 100 * s.Thickness / total
		}
		out = append(out, *s)
	}
	rank := func(k string) int {
		if i := slices.Index(order, k); i >= 0 {
			return i
		}
		if k == "" {
			return len(order) + 1
		}
		return len(order)
	}
	slices.SortStableFunc(out, func(a, b Share) int {
		if ra, rb := rank(a.Key), rank(b.Key); ra != rb {
			return ra - rb
		}
		if a.Thickness != b.Thickness {
			if a.Thickness > b.Thickness {
				return -1
			}
			return 1
		}
		return strings.Compare(a.Key, b.Key)
	})
	return out
}

// grainGroup is the Wentworth group of a bed's mean grain size: gravel coarser than −1 φ,
// mud finer than 4 φ.
func grainGroup(g model.Grain) string {
	phi, ok := meanPhi(g)
	switch {
	case !ok:
		return ""
	case phi < -1:
		return "gravel"
	case phi > 4:
		return "mud"
	}
	return "sand"
}

func meanPhi(g model.Grain) (float64, bool) {
	switch {
	case g.Base != nil && g.Top != nil:
		return (*g.Base + *g.Top) / 2, true
	case g.Base != nil:
		return *g.Base, true
	case g.Top != nil:
		return *g.Top, true
	}
	return 0, false
}

// grainTrend says whether a bed fines or coarsens upwards (more than a quarter of a φ unit).
func grainTrend(g model.Grain) string {
	if g.Base == nil || g.Top == nil {
		return ""
	}
	switch d := *g.Top - *g.Base; {
	case d > 0.25:
		return "fining"
	case d < -0.25:
		return "coarsening"
	}
	return "constant"
}
