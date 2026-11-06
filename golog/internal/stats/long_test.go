package stats

import (
	"strconv"
	"testing"
	"time"

	"golog/internal/model"
)

// A log at GoLog's bed limit, with the number of shuffles the server allows for it, stays quick.
func TestLongLogIsQuick(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	c := model.NewLogContent()
	for i := range model.MaxBeds {
		c.Beds = append(c.Beds, model.Bed{ID: strconv.Itoa(i), Thickness: float64(1+i*7919%97) / 100, Facies: strconv.Itoa(1 + i*31%7)})
	}
	start := time.Now()
	r := Analyse(c, DefaultShuffles)
	if d := time.Since(start); d > 10*time.Second {
		t.Errorf("10,000 beds took %v", d)
	} else {
		t.Logf("10,000 beds, %d shuffles: %v", r.Facies.M.Shuffles, d)
	}
}
