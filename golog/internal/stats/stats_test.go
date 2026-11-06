package stats

import (
	"math"
	"math/rand/v2"
	"slices"
	"strconv"
	"testing"

	"golog/internal/model"
)

func near(t *testing.T, what string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Errorf("%s = %.6f, want %.6f", what, got, want)
	}
}

// log builds beds from thicknesses (cm) and facies codes ("" = none).
func log(thick []float64, facies []string) model.LogContent {
	c := model.NewLogContent()
	for i, th := range thick {
		c.Beds = append(c.Beds, model.Bed{ID: strconv.Itoa(i), Thickness: th / 100, Contact: model.ContactSharp, Facies: facies[i]})
	}
	return c
}

func codes(s ...int) []string {
	out := make([]string, len(s))
	for i, x := range s {
		if x != 0 {
			out[i] = strconv.Itoa(x)
		}
	}
	return out
}

// The cases below were also run through SedLog 3.1's "Order Analysis Single Section"; the
// comments give what SedLog showed.

func TestExample1(t *testing.T) {
	// SedLog's example1: facies 1 2 3 4 4 4 2 5 2. SedLog: m = 0.5, r = 0.89.
	r := Analyse(log([]float64{100, 45, 45, 135, 65, 50, 60, 35, 65}, codes(1, 2, 3, 4, 4, 4, 2, 5, 2)), 500)
	near(t, "m", r.Facies.M.Observed, 0.5)
	near(t, "r", r.Runs.R.Observed, 8.0/9)
	if want := []int{1, 3, 1, 3, 1}; !slices.Equal(r.Facies.Beds, want) {
		t.Errorf("beds per facies %v, want %v", r.Facies.Beds, want)
	}
	// rows as SedLog showed them: f2 → f3 and f5 half each, every other row a single 1.00
	want := [][]float64{{0, 1, 0, 0, 0}, {0, 0, 0.5, 0, 0.5}, {0, 0, 0, 1, 0}, {0, 1, 0, 0, 0}, {0, 1, 0, 0, 0}}
	for i := range want {
		if !slices.Equal(r.Facies.Prob[i], want[i]) {
			t.Errorf("row %d = %v, want %v", i, r.Facies.Prob[i], want[i])
		}
	}
}

func TestNoFaciesBreaksSuccession(t *testing.T) {
	// facies 1 0 2 1 0 2 1 (0 = none). SedLog: only 2 → 1 is counted, m = 0, and every bed
	// is in the runs analysis (r = 3.0 for steadily thickening beds).
	r := Analyse(log([]float64{10, 20, 30, 40, 50, 60, 70}, codes(1, 0, 2, 1, 0, 2, 1)), 200)
	if want := [][]int{{0, 0}, {2, 0}}; !slices.Equal(r.Facies.Counts[0], want[0]) || !slices.Equal(r.Facies.Counts[1], want[1]) {
		t.Errorf("counts %v, want %v", r.Facies.Counts, want)
	}
	near(t, "m", r.Facies.M.Observed, 0)
	near(t, "r", r.Runs.R.Observed, 3)
	if want := []int{3, 2}; !slices.Equal(r.Facies.Beds, want) {
		t.Errorf("beds per facies %v, want %v", r.Facies.Beds, want)
	}
}

func TestMainDiagonalNotInM(t *testing.T) {
	// 1 2 3 1 3 2 1 2 3 1: every row 2/3 one step on, 1/3 two steps on. The offset sums are
	// 2 and 1, so m = (2 − 1)/3. (Counting the main diagonal's 0 as the smallest would give 2/3.)
	r := Analyse(log([]float64{10, 20, 30, 25, 15, 35, 40, 30, 50, 45}, codes(1, 2, 3, 1, 3, 2, 1, 2, 3, 1)), 200)
	near(t, "m", r.Facies.M.Observed, 1.0/3)
}

func TestGapsInFaciesNumbersClosed(t *testing.T) {
	// facies 1 3 1 3 7 1. SedLog makes a 7 × 7 matrix with empty rows for 2, 4, 5 and 6 and
	// gets m = 1/7; GoLog numbers the facies 1, 3, 7 consecutively, as Burgess advises.
	r := Analyse(log([]float64{10, 20, 30, 40, 50, 60}, codes(1, 3, 1, 3, 7, 1)), 200)
	if want := []string{"1", "3", "7"}; !slices.Equal(r.Facies.Codes, want) {
		t.Errorf("codes %v, want %v", r.Facies.Codes, want)
	}
	near(t, "m", r.Facies.M.Observed, 2.0/3)
	near(t, "r", r.Runs.R.Observed, 2.5) // SedLog: 2.5
}

func TestTextFaciesInOrderOfAppearance(t *testing.T) {
	r := Analyse(log([]float64{1, 2, 3, 4, 5, 6}, []string{"St", "Sh", "Fm", "St", "Sh", "Fm"}), 100)
	if want := []string{"St", "Sh", "Fm"}; !slices.Equal(r.Facies.Codes, want) {
		t.Errorf("codes %v, want %v", r.Facies.Codes, want)
	}
	near(t, "m", r.Facies.M.Observed, 1) // a strict cycle
}

func TestPaperExamples(t *testing.T) {
	// Burgess (2016): a strict asymmetric cycle gives m = 1, a symmetric one 1 2 3 4 5 4 3 2 1 … 0.5.
	asym := []int{}
	sym := []int{}
	for range 6 {
		asym = append(asym, 1, 2, 3, 4, 5)
		sym = append(sym, 1, 2, 3, 4, 5, 4, 3, 2)
	}
	sym = append(sym, 1)
	near(t, "m asymmetric", metricM(counts(asym)), 1)
	near(t, "m symmetric", metricM(counts(sym)), 0.5)

	// The runs examples in the paper: two thickening cycles (I 1 2 3 4 0 1 2 3 4, D 0 0 0 0 1 …),
	// and thicknesses from a random generator. For the second the paper lists I 0 0 1 2 0 1 2 3 0
	// and D 1 2 0 0 1 0 0 0 1, but 0.13 0.61 0.86 0.26 0.25 is two steps up and two down
	// (I … 1 2 0 0, D … 0 0 1 2), so r = (6 + 7) / 10.
	near(t, "r cycles", metricR([]float64{0.2, 0.4, 0.6, 0.8, 1.0, 0.2, 0.4, 0.6, 0.8, 1.0}), 2.1)
	near(t, "r random", metricR([]float64{0.87, 0.80, 0.35, 0.57, 0.86, 0.13, 0.61, 0.86, 0.26, 0.25}), 1.3)
}

func counts(seq []int) [][]int {
	s := make([]int, len(seq))
	f := 0
	for i, x := range seq {
		s[i] = x - 1
		f = max(f, x)
	}
	c, _, _ := markov(s, f)
	return c
}

func TestRunsEqualThicknessEndsRun(t *testing.T) {
	// 10 20 30 30 40 50 60 70: the second 30 ends the run, so 1+2+0+1+2+3+4 = 13 over 8 beds.
	near(t, "r", metricR([]float64{10, 20, 30, 30, 40, 50, 60, 70}), 13.0/8)
	r := Analyse(log([]float64{10, 20, 30, 30, 40}, codes(1, 2, 1, 2, 1)), 10)
	if want := []int{0, 1, 2, 0, 1}; !slices.Equal(r.Runs.Thickening, want) {
		t.Errorf("thickening %v, want %v", r.Runs.Thickening, want)
	}
}

func TestShuffleKeepsBedsAndAvoidsSameFacies(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	f := []int{0, 1, 2, 0, 1, 2, 0, 1, 2, 3, 3, -1}
	th := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
	pair := map[float64]int{}
	for i := range f {
		pair[th[i]] = f[i]
	}
	for range 200 {
		shuffle(rng, f, th)
		for i := range f {
			if pair[th[i]] != f[i] {
				t.Fatalf("a bed lost its facies: %v %v", f, th)
			}
		}
	}
	got := slices.Clone(th)
	slices.Sort(got)
	if !slices.Equal(got, []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}) {
		t.Fatalf("beds lost: %v", got)
	}
	// One facies everywhere: no swap is possible, and shuffle must still return.
	same := []int{0, 0, 0, 0}
	shuffle(rng, same, []float64{1, 2, 3, 4})
}

func TestOrderedSectionHasSmallP(t *testing.T) {
	var th []float64
	var fa []string
	for range 12 {
		for k, f := range []string{"1", "2", "3", "4"} {
			th = append(th, float64(10+10*k))
			fa = append(fa, f)
		}
	}
	r := Analyse(log(th, fa), 1000)
	near(t, "m", r.Facies.M.Observed, 1)
	if r.Facies.M.P > 0.01 || r.Runs.R.P > 0.01 {
		t.Errorf("p for a strict thickening cycle: m %.3f r %.3f, want < 0.01", r.Facies.M.P, r.Runs.R.P)
	}
	if r.Facies.M.Evidence != "very strong evidence of order" {
		t.Errorf("evidence %q", r.Facies.M.Evidence)
	}
	n := 0
	for _, c := range r.Facies.M.Bins.Counts {
		n += c
	}
	if n != 1000 {
		t.Errorf("histogram holds %d values, want 1000", n)
	}
	// The same log gives the same numbers every time.
	again := Analyse(log(th, fa), 1000)
	if again.Facies.M.P != r.Facies.M.P || again.Runs.R.Mean != r.Runs.R.Mean {
		t.Error("results differ between two runs on the same log")
	}
}

func TestSummary(t *testing.T) {
	phi := func(v float64) *float64 { return &v }
	c := model.NewLogContent()
	c.Beds = []model.Bed{
		{ID: "a", Thickness: 1, Contact: "erosional", Facies: "1", Grain: model.Grain{Base: phi(-2), Top: phi(1)},
			Lithology: []model.Lithology{{Pattern: "sandstone", Percent: 75}, {Pattern: "conglomerate", Percent: 25}}},
		{ID: "b", Thickness: 3, Contact: "sharp", Grain: model.Grain{Base: phi(6), Top: phi(6)},
			Lithology: []model.Lithology{{Pattern: "mudstone", Percent: 100}}},
	}
	r := Analyse(c, 10)
	s := r.Summary
	near(t, "thickness", s.Thickness, 4)
	near(t, "median", s.MedianBed, 2)
	if s.Lithologies[0].Key != "mudstone" || s.Lithologies[1].Key != "sandstone" {
		t.Errorf("lithologies %+v", s.Lithologies)
	}
	near(t, "sandstone", s.Lithologies[1].Thickness, 0.75)
	near(t, "mud %", s.Grain[1].Percent, 75)
	if s.Trends[0].Key != "fining" || s.Trends[0].Beds != 1 {
		t.Errorf("trends %+v", s.Trends)
	}
	if r.Facies != nil || len(r.Notes) == 0 {
		t.Error("one facies should give a note, not a facies test")
	}
	if r.Runs != nil {
		t.Error("two beds should not give a runs test")
	}
}

func TestRunsWhereSedLogDiffers(t *testing.T) {
	// SedLog 3.1 scores a run of thinning (or equal) beds at the very base of the section one
	// lower per bed than Burgess's formula: 0, 1, 2 … instead of 1, 2, 3 …. Elsewhere it agrees.
	// GoLog follows the paper, whose own example scores the first thinning bed 1.
	for _, c := range []struct {
		thick  []float64
		golog  float64 // Burgess (2016)
		sedlog float64 // shown by SedLog 3.1
	}{
		{[]float64{100, 90, 80, 70, 60, 50}, 15.0 / 6, 10.0 / 6},
		{[]float64{50, 40, 30, 20, 10, 60, 55, 45, 35}, 17.0 / 9, 13.0 / 9},
		{[]float64{30, 20, 10, 20, 10, 5}, 7.0 / 6, 5.0 / 6},
		{[]float64{50, 40, 40, 30, 60, 60, 70}, 4.0 / 7, 5.0 / 7},
		// the same in both
		{[]float64{10, 20, 30, 40, 30, 20, 10, 5, 40, 30}, 1.8, 1.8},
		{[]float64{50, 50, 40, 30}, 0.75, 0.75},
		{[]float64{10, 20, 30, 20, 20, 10}, 5.0 / 6, 5.0 / 6},
	} {
		near(t, "r", metricR(c.thick), c.golog)
		near(t, "SedLog's r", sedlogR(c.thick), c.sedlog)
	}
}

// sedlogR is the runs metric as SedLog 3.1 calculates it, worked out from 16 test sections:
// while the beds have not yet got thicker, bed k scores k−1.
func sedlogR(thick []float64) float64 {
	k := 1
	sum := 0
	for ; k < len(thick) && thick[k] <= thick[k-1]; k++ {
		sum += k - 1
	}
	if k < len(thick) {
		sum += int(math.Round(metricR(thick[k-1:]) * float64(len(thick)-k+1)))
	}
	return float64(sum) / float64(len(thick))
}
