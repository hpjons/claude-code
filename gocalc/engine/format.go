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

// formatDecimal renders a result as a clean string: rounded (half up) to
// displayPrecision significant digits, with redundant trailing zeros
// trimmed (e.g. 5.00 -> 5) and never "-0".
//
// Values that need more integer digits than displayPrecision, or that are
// smaller than 10^minPlainExponent, are shown in scientific notation (e.g.
// 1e100, 1.5e-12) rather than as a huge fixed-point string. Both forms
// parse back as valid input.
//
// The rounding works on the digit string rather than through apd: near
// apd's smallest exponent (results around 1e-99999), a 60-digit
// coefficient's exponent is below apd's limit and apd's Round fails.
func formatDecimal(d *apd.Decimal) string {
	if d.Form != apd.Finite {
		return d.Text('f')
	}
	if d.IsZero() {
		return "0"
	}

	digits, exp := roundDigits(d.Coeff.String(), int64(d.Exponent), displayPrecision)
	trimmed := strings.TrimRight(digits, "0")
	exp += int64(len(digits) - len(trimmed))
	digits = trimmed
	// adjusted is the exponent of the leading digit: d = x.xxx * 10^adjusted.
	adjusted := exp + int64(len(digits)) - 1

	var b strings.Builder
	if d.Negative {
		b.WriteByte('-')
	}

	if adjusted >= minPlainExponent && adjusted < displayPrecision {
		// Fixed point: digits × 10^exp.
		switch point := int64(len(digits)) + exp; {
		case exp >= 0:
			b.WriteString(digits)
			b.WriteString(strings.Repeat("0", int(exp)))
		case point > 0:
			b.WriteString(digits[:point])
			b.WriteByte('.')
			b.WriteString(digits[point:])
		default:
			b.WriteString("0.")
			b.WriteString(strings.Repeat("0", int(-point)))
			b.WriteString(digits)
		}
		return b.String()
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

// roundDigits rounds the coefficient digits × 10^exp to at most n
// significant digits, half up, returning the new digits and exponent.
func roundDigits(digits string, exp int64, n int) (string, int64) {
	if len(digits) <= n {
		return digits, exp
	}
	roundUp := digits[n] >= '5'
	exp += int64(len(digits) - n)
	digits = digits[:n]
	if !roundUp {
		return digits, exp
	}

	// Add one to the last kept digit, carrying left.
	b := []byte(digits)
	i := len(b) - 1
	for ; i >= 0 && b[i] == '9'; i-- {
		b[i] = '0'
	}
	if i >= 0 {
		b[i]++
		return string(b), exp
	}
	// All nines: 99…9 + 1 = 100…0, one digit longer; drop a zero.
	return "1" + string(b[:len(b)-1]), exp + 1
}
