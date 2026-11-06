package engine

import (
	"sync"

	"github.com/cockroachdb/apd/v3"

	"gocalc/i18n"
)

// Trigonometric functions computed in decimal at working precision. They
// used to go through float64, which limited every result to ~16 digits
// and made identities like sin(x)^2 + cos(x)^2 = 1 visibly fail.

// maxTrigExponent is the largest adjusted exponent (x ≈ d.ddd × 10^e)
// accepted for sin/cos/tan. Reducing x modulo π/2 needs π to about that
// many digits beyond the working precision (see piTo); sin(10^1000)
// works, sin(10^1001) is "out of range".
const maxTrigExponent = 1000

// trigGuardDigits are extra digits carried inside series evaluation and
// argument reduction, beyond the working precision, so rounding inside
// the algorithms doesn't show in the result.
const trigGuardDigits = 20

var (
	piMu     sync.Mutex
	piCached *apd.Decimal // π to at least the most digits asked for so far
)

// piTo returns π rounded to the given number of significant digits.
//
// π is computed on demand and cached; the cache only grows when more
// digits are needed, so ordinary use stays cheap (~100 digits) and the
// cost of ~1,100 digits (~50 ms) is paid only once an argument near
// 10^maxTrigExponent actually needs them.
func piTo(digits uint32) *apd.Decimal {
	piMu.Lock()
	if piCached == nil || piCached.NumDigits() < int64(digits) {
		// Round the size up so a run of slightly larger requests
		// doesn't recompute π each time.
		piCached = computePi((digits/100 + 1) * 100)
	}
	cached := piCached
	piMu.Unlock()

	d := new(apd.Decimal)
	if _, err := apd.BaseContext.WithPrecision(digits).Round(d, cached); err != nil {
		panic(err) // rounding a finite constant can't fail
	}
	return d
}

// computePi computes π to the given number of significant digits with
// Machin's formula, π = 16·arctan(1/5) − 4·arctan(1/239).
func computePi(digits uint32) *apd.Decimal {
	c := newCalc(digits + 10)
	a := c.mul(apd.New(16, 0), arctanInv(c, 5))
	b := c.mul(apd.New(4, 0), arctanInv(c, 239))
	pi := c.sub(a, b)
	pi = newCalc(digits).round(pi)
	if c.err != nil {
		panic(c.err) // fixed constants; can't fail
	}
	return pi
}

// arctanInv returns arctan(1/n) = Σ (−1)^k / ((2k+1)·n^(2k+1)).
func arctanInv(c *calc, n int64) *apd.Decimal {
	n2 := apd.New(n*n, 0)
	power := c.quo(apd.New(1, 0), apd.New(n, 0)) // 1/n^(2k+1)
	limit := apd.New(1, -int32(c.ctx.Precision)-2)
	sum := new(apd.Decimal)
	for k := int64(0); c.err == nil; k++ {
		term := c.quo(power, apd.New(2*k+1, 0))
		if term.Cmp(limit) < 0 {
			break
		}
		if k%2 == 0 {
			sum = c.add(sum, term)
		} else {
			sum = c.sub(sum, term)
		}
		power = c.quo(power, n2)
	}
	return sum
}

// sinCos returns sin(x) and cos(x), rounded to the working precision.
func sinCos(x *apd.Decimal) (sin, cos *apd.Decimal, err error) {
	e := max(adjustedExponent(x), 0)

	// Reduce: x = k·(π/2) + r with |r| ≤ π/4. k has up to e+1 digits, so
	// the reduction needs e extra digits to keep r accurate.
	red := newCalc(workingPrecision + trigGuardDigits + uint32(e))
	halfPi := red.quo(piTo(red.ctx.Precision), apd.New(2, 0))
	k := new(apd.Decimal)
	if red.err == nil {
		_, red.err = red.ctx.RoundToIntegralValue(k, red.quo(x, halfPi))
	}
	r := red.sub(x, red.mul(k, halfPi))
	quadrant := new(apd.Decimal)
	if red.err == nil {
		_, red.err = red.ctx.Rem(quadrant, k, apd.New(4, 0))
	}
	if red.err != nil {
		return nil, nil, red.err
	}
	q, _ := quadrant.Int64() // in (−4, 4)
	q = (q%4 + 4) % 4

	c := newCalc(workingPrecision + trigGuardDigits)
	s, co := sinSeries(c, r), cosSeries(c, r)
	if c.err != nil {
		return nil, nil, c.err
	}

	// sin(r + q·π/2) and cos(r + q·π/2).
	switch q {
	case 1:
		s, co = co, c.neg(s)
	case 2:
		s, co = c.neg(s), c.neg(co)
	case 3:
		s, co = c.neg(co), s
	}

	w := newCalc(workingPrecision)
	sin, cos = w.round(s), w.round(co)
	return sin, cos, w.err
}

// sinSeries returns sin(r) = r − r³/3! + r⁵/5! − … for |r| ≤ π/4.
func sinSeries(c *calc, r *apd.Decimal) *apd.Decimal {
	if r.IsZero() || adjustedExponent(r) < -int64(c.ctx.Precision) {
		return r // r³/6 is below the precision
	}
	r2 := c.mul(r, r)
	sum, term := r, r
	stop := adjustedExponent(r) - int64(c.ctx.Precision) - 2
	for n := int64(1); c.err == nil; n++ {
		term = c.neg(c.quo(c.mul(term, r2), apd.New((2*n)*(2*n+1), 0)))
		if term.IsZero() || adjustedExponent(term) < stop {
			break
		}
		sum = c.add(sum, term)
	}
	return sum
}

// cosSeries returns cos(r) = 1 − r²/2! + r⁴/4! − … for |r| ≤ π/4.
func cosSeries(c *calc, r *apd.Decimal) *apd.Decimal {
	one := apd.New(1, 0)
	if r.IsZero() || adjustedExponent(r) < -int64(c.ctx.Precision) {
		return one // r²/2 is below the precision
	}
	r2 := c.mul(r, r)
	sum, term := one, one
	stop := -int64(c.ctx.Precision) - 2 // cos(r) ≥ 0.7 here
	for n := int64(1); c.err == nil; n++ {
		term = c.neg(c.quo(c.mul(term, r2), apd.New((2*n-1)*(2*n), 0)))
		if term.IsZero() || adjustedExponent(term) < stop {
			break
		}
		sum = c.add(sum, term)
	}
	return sum
}

// arctan returns arctan(x), rounded to the working precision.
func arctan(x *apd.Decimal) (*apd.Decimal, error) {
	c := newCalc(workingPrecision + trigGuardDigits)
	res := arctanFull(c, x)
	w := newCalc(workingPrecision)
	out := w.round(res)
	if c.err != nil {
		return nil, c.err
	}
	return out, w.err
}

// arctanFull computes arctan(x) at c's precision.
func arctanFull(c *calc, x *apd.Decimal) *apd.Decimal {
	prec := int64(c.ctx.Precision)
	// For tiny |x|, arctan(x) = x − x³/3 + … = x at this precision. (Also
	// avoids exponents beyond apd's range in the steps below.)
	if x.IsZero() || adjustedExponent(x) < -prec {
		return x
	}
	one := apd.New(1, 0)
	abs := new(apd.Decimal).Abs(x)

	// arctan(x) = sign(x)·π/2 − arctan(1/x) for |x| > 1.
	if abs.Cmp(one) > 0 {
		halfPi := c.quo(piTo(c.ctx.Precision), apd.New(2, 0))
		var res *apd.Decimal
		if adjustedExponent(abs) > prec {
			res = halfPi // arctan(1/x) is below the precision
		} else {
			res = c.sub(halfPi, arctanFull(c, c.quo(one, abs)))
		}
		if x.Negative {
			res = c.neg(res)
		}
		return res
	}

	// Halve the angle until |x| ≤ 0.1 so the series converges quickly:
	// arctan(x) = 2·arctan(x / (1 + √(1 + x²))).
	tenth := apd.New(1, -1)
	halvings := int64(0)
	for c.err == nil && new(apd.Decimal).Abs(x).Cmp(tenth) > 0 {
		x = c.quo(x, c.add(one, c.sqrt(c.add(one, c.mul(x, x)))))
		halvings++
	}

	// arctan(x) = x − x³/3 + x⁵/5 − …
	sum := x
	if adjustedExponent(x) >= -prec {
		x2 := c.mul(x, x)
		power := x
		stop := adjustedExponent(x) - int64(c.ctx.Precision) - 2
		for k := int64(1); c.err == nil; k++ {
			power = c.neg(c.mul(power, x2))
			term := c.quo(power, apd.New(2*k+1, 0))
			if term.IsZero() || adjustedExponent(term) < stop {
				break
			}
			sum = c.add(sum, term)
		}
	}
	return c.mul(sum, apd.New(1<<halvings, 0))
}

// arcsin returns arcsin(x) for |x| ≤ 1, rounded to the working precision.
func arcsin(x *apd.Decimal) (*apd.Decimal, error) {
	one := apd.New(1, 0)
	abs := new(apd.Decimal).Abs(x)
	switch abs.Cmp(one) {
	case 1:
		return nil, i18n.Errorf("%s: argument out of domain", "arcsin")
	case 0:
		w := newCalc(workingPrecision)
		res := w.quo(piTo(workingPrecision+trigGuardDigits), apd.New(2, 0))
		if x.Negative {
			res = w.neg(res)
		}
		return res, w.err
	}
	// For tiny |x|, arcsin(x) = x + x³/6 + … = x at working precision.
	if x.IsZero() || adjustedExponent(x) < -(workingPrecision+trigGuardDigits) {
		w := newCalc(workingPrecision)
		return w.round(x), w.err
	}
	// arcsin(x) = arctan(x / √((1−x)(1+x))). (1−x)(1+x) avoids the
	// cancellation in 1 − x² when x is close to ±1.
	c := newCalc(workingPrecision + trigGuardDigits)
	t := c.quo(x, c.sqrt(c.mul(c.sub(one, x), c.add(one, x))))
	if c.err != nil {
		return nil, c.err
	}
	return arctan(t)
}

// arccos returns arccos(x) for |x| ≤ 1, rounded to the working precision.
func arccos(x *apd.Decimal) (*apd.Decimal, error) {
	one := apd.New(1, 0)
	if new(apd.Decimal).Abs(x).Cmp(one) > 0 {
		return nil, i18n.Errorf("%s: argument out of domain", "arccos")
	}
	if x.Cmp(apd.New(-1, 0)) == 0 {
		w := newCalc(workingPrecision)
		return w.round(piTo(workingPrecision + trigGuardDigits)), w.err
	}
	// arccos(x) = 2·arctan(√((1−x)/(1+x))), which stays accurate near ±1,
	// unlike π/2 − arcsin(x).
	c := newCalc(workingPrecision + trigGuardDigits)
	t := c.sqrt(c.quo(c.sub(one, x), c.add(one, x)))
	res := c.mul(apd.New(2, 0), arctanFull(c, t))
	w := newCalc(workingPrecision)
	out := w.round(res)
	if c.err != nil {
		return nil, c.err
	}
	return out, w.err
}

// adjustedExponent returns e such that d = d.ddd × 10^e (for nonzero d).
func adjustedExponent(d *apd.Decimal) int64 {
	return int64(d.Exponent) + d.NumDigits() - 1
}

// calc wraps an apd.Context so a chain of operations can be written as
// expressions; the first error is kept and later operations are skipped.
type calc struct {
	ctx *apd.Context
	err error
}

func newCalc(precision uint32) *calc {
	return &calc{ctx: apd.BaseContext.WithPrecision(precision)}
}

func (c *calc) op(f func(d *apd.Decimal) (apd.Condition, error)) *apd.Decimal {
	d := new(apd.Decimal)
	if c.err == nil {
		_, c.err = f(d)
	}
	return d
}

func (c *calc) add(a, b *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Add(d, a, b) })
}

func (c *calc) sub(a, b *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Sub(d, a, b) })
}

func (c *calc) mul(a, b *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Mul(d, a, b) })
}

func (c *calc) quo(a, b *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Quo(d, a, b) })
}

func (c *calc) sqrt(a *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Sqrt(d, a) })
}

func (c *calc) neg(a *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Neg(d, a) })
}

func (c *calc) round(a *apd.Decimal) *apd.Decimal {
	return c.op(func(d *apd.Decimal) (apd.Condition, error) { return c.ctx.Round(d, a) })
}
