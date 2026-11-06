package engine

import "github.com/cockroachdb/apd/v3"

// formatDecimal renders a decimal result as a clean fixed-point string,
// trimming redundant trailing zeros (e.g. 5.00 -> 5).
func formatDecimal(ctx *apd.Context, d *apd.Decimal) string {
	reduced := new(apd.Decimal)
	ctx.Reduce(reduced, d)
	return reduced.Text('f')
}
