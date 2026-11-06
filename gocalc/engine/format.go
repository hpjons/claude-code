package engine

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
)

// minPlainExponent is the smallest decimal exponent shown in plain
// fixed-point form; anything smaller (e.g. 1e-11) switches to scientific
// notation so results don't turn into long runs of leading zeros.
const minPlainExponent = -10

// formatDecimal renders a decimal result as a clean string, trimming
// redundant trailing zeros (e.g. 5.00 -> 5) and never showing "-0".
//
// Values whose magnitude needs more integer digits than the context's
// precision, or that are smaller than 10^minPlainExponent, are shown in
// scientific notation (e.g. 1e100, 1.5e-12) rather than as a huge
// fixed-point string. Both forms parse back as valid input, which matters
// because results are fed back in as "ans".
func formatDecimal(ctx *apd.Context, d *apd.Decimal) string {
	reduced := new(apd.Decimal)
	ctx.Reduce(reduced, d)

	if reduced.IsZero() {
		return "0"
	}
	if reduced.Form != apd.Finite {
		return reduced.Text('f')
	}

	digits := reduced.Coeff.String()
	// adjusted is the exponent of the leading digit: d = x.xxx * 10^adjusted.
	adjusted := int64(reduced.Exponent) + int64(len(digits)) - 1

	if adjusted >= minPlainExponent && adjusted < int64(ctx.Precision) {
		return reduced.Text('f')
	}

	var b strings.Builder
	if reduced.Negative {
		b.WriteByte('-')
	}
	b.WriteByte(digits[0])
	if len(digits) > 1 {
		b.WriteByte('.')
		b.WriteString(digits[1:])
	}
	b.WriteByte('e')
	b.WriteString(strconv.FormatInt(adjusted, 10))
	return b.String()
}
