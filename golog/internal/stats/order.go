// Package stats calculates summaries of a log and tests whether its beds are ordered.
//
// The order tests follow Burgess, P.M. (2016), "Identifying ordered strata: evidence, methods,
// and meaning", Journal of Sedimentary Research 86, 148–167, doi:10.2110/jsr.2016.10: the
// Markov order metric m for the facies succession, the runs order metric r for bed thickness,
// and both compared with the same section shuffled many times.
package stats

import (
	"math"
	"slices"
)

// markov returns the facies transition counts, the transition probabilities (each row adds up
// to 1, or 0 for a facies nothing follows) and the Markov order metric m.
//
// seq lists the facies of the beds from the base as numbers 0…f-1, with −1 for a bed without a
// facies. A bed without a facies breaks the succession (as in SedLog: what lies between is not
// known), and a bed followed by one of the same facies is not a transition (an embedded matrix,
// so the diagonal is always zero).
func markov(seq []int, f int) (counts [][]int, prob [][]float64, m float64) {
	counts = make([][]int, f)
	for i := range counts {
		counts[i] = make([]int, f)
	}
	prev := -1
	for _, x := range seq {
		if x >= 0 && prev >= 0 && x != prev {
			counts[prev][x]++
		}
		prev = x
	}
	return counts, probabilities(counts), metricM(counts)
}

func probabilities(counts [][]int) [][]float64 {
	prob := make([][]float64, len(counts))
	for i, row := range counts {
		prob[i] = make([]float64, len(row))
		total := 0
		for _, c := range row {
			total += c
		}
		if total > 0 {
			for j, c := range row {
				prob[i][j] = float64(c) / float64(total)
			}
		}
	}
	return prob
}

// metricM is Burgess's m: for each offset j from the main diagonal, add up the transition
// probabilities on that diagonal, wrapping round the matrix (cells i → (i+j) mod F); m is the
// largest sum minus the smallest, divided by the number of facies F. A strict cycle 1→2→…→F→1
// puts every transition on one diagonal and gives 1; disordered beds spread them evenly and give
// a value near 0.
func metricM(counts [][]int) float64 {
	f := len(counts)
	if f < 2 {
		return 0
	}
	totals := make([]int, f)
	for i, row := range counts {
		for _, c := range row {
			totals[i] += c
		}
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for j := 1; j < f; j++ {
		s := 0.0
		for i := range f {
			if totals[i] > 0 {
				s += float64(counts[i][(i+j)%f]) / float64(totals[i])
			}
		}
		lo, hi = min(lo, s), max(hi, s)
	}
	return (hi - lo) / float64(f)
}

// metricR is Burgess's runs order metric r for a list of thicknesses from the base: each unit
// thicker than the one below continues (or starts) a thickening run and scores its place in that
// run (1, 2, 3 …); thinning runs score the same way. A unit as thick as the one below ends both
// runs and scores 0, as does the lowest unit. r is the sum of the scores divided by the number of
// units, so long runs, as in thickening- or thinning-upward cycles, give a high value.
func metricR(thick []float64) float64 {
	if len(thick) == 0 {
		return 0
	}
	sum, up, down := 0, 0, 0
	for k := 1; k < len(thick); k++ {
		switch {
		case thick[k] > thick[k-1]:
			up, down = up+1, 0
		case thick[k] < thick[k-1]:
			up, down = 0, down+1
		default:
			up, down = 0, 0
		}
		sum += up + down
	}
	return float64(sum) / float64(len(thick))
}

// shuffle reorders the units the way Burgess does: n times, pick two units at random and swap
// them, but not when the swap would put two units of the same facies next to each other; then
// pick another pair. units without a facies (−1) never clash.
func shuffle(rng interface{ IntN(int) int }, facies []int, thick []float64) {
	n := len(facies)
	if n < 2 {
		return
	}
	clash := func(i int) bool {
		x := facies[i]
		return x >= 0 && (i > 0 && facies[i-1] == x || i+1 < n && facies[i+1] == x)
	}
	for range n {
		// A section can be so dominated by one facies that no swap is allowed; give up on this
		// swap after a fair number of tries rather than looping for ever.
		for try := 0; try < 50*n+100; try++ {
			a, b := rng.IntN(n), rng.IntN(n)
			if a == b || facies[a] == facies[b] {
				continue
			}
			facies[a], facies[b] = facies[b], facies[a]
			if clash(a) || clash(b) {
				facies[a], facies[b] = facies[b], facies[a]
				continue
			}
			thick[a], thick[b] = thick[b], thick[a]
			break
		}
	}
}

// Test compares an observed metric with the same metric in shuffled copies of the section.
type Test struct {
	Observed float64   `json:"observed"`
	Shuffles int       `json:"shuffles"`
	Mean     float64   `json:"mean"`     // mean of the shuffled values
	P        float64   `json:"p"`        // share of shuffled sections with a value ≥ the observed one
	Evidence string    `json:"evidence"` // what P says, in words
	Bins     Histogram `json:"bins"`
}

// Histogram counts values in equal bins starting at Start.
type Histogram struct {
	Start  float64 `json:"start"`
	Width  float64 `json:"width"`
	Counts []int   `json:"counts"`
}

func newTest(observed float64, shuffled []float64, width float64) Test {
	t := Test{Observed: observed, Shuffles: len(shuffled)}
	if len(shuffled) == 0 {
		return t
	}
	top := observed
	ge := 0
	for _, v := range shuffled {
		t.Mean += v
		top = max(top, v)
		if v >= observed-1e-9 {
			ge++
		}
	}
	t.Mean /= float64(len(shuffled))
	t.P = float64(ge) / float64(len(shuffled))
	t.Evidence = evidence(t.P)
	nb := int(math.Floor(top/width+1e-9)) + 1
	t.Bins = Histogram{Width: width, Counts: make([]int, nb)}
	for _, v := range shuffled {
		t.Bins.Counts[min(nb-1, int(math.Floor(v/width+1e-9)))]++
	}
	return t
}

// evidence puts a p value in words, using the scale Burgess (2016) suggests.
func evidence(p float64) string {
	switch {
	case p < 0.01:
		return "very strong evidence of order"
	case p < 0.05:
		return "strong evidence of order"
	case p <= 0.10:
		return "weak evidence of order"
	}
	return "no evidence of order: shuffled beds do as well"
}

// median of a list that may be changed.
func median(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	slices.Sort(v)
	if len(v)%2 == 1 {
		return v[len(v)/2]
	}
	return (v[len(v)/2-1] + v[len(v)/2]) / 2
}
