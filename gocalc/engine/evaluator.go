package engine

import (
	"math/big"

	"github.com/cockroachdb/apd/v3"

	"gocalc/i18n"
)

// workingPrecision is the number of significant digits every calculation
// is carried out with. It's deliberately higher than displayPrecision:
// the extra guard digits absorb rounding in intermediate steps, so e.g.
// 1/3*3 and sin(2)^2 + cos(2)^2 display as exactly 1.
const workingPrecision = 60

// displayPrecision is the number of significant digits results are
// rounded to for display (see formatDecimal).
const displayPrecision = 50

// pi is π at working precision, computed once at startup. eval hands out
// copies so no caller can mutate the shared value.
var pi = piTo(workingPrecision)

// defaultContext returns a fresh decimal context at working precision.
func defaultContext() *apd.Context {
	return apd.BaseContext.WithPrecision(workingPrecision)
}

// decimalFunctions are single-argument functions apd implements natively.
var decimalFunctions = map[string]func(ctx *apd.Context, d, x *apd.Decimal) (apd.Condition, error){
	"sqrt": (*apd.Context).Sqrt,
	"exp":  (*apd.Context).Exp,
}

// trigFunctions are the trigonometric functions, computed in decimal at
// working precision (see trig.go). Inverse functions are named arcsin/
// arccos/arctan rather than SpeedCrunch's asin/acos/atan, per preference.
var trigFunctions = map[string]func(x value) (*apd.Decimal, error){
	"sin":    func(x value) (*apd.Decimal, error) { return periodic("sin", x) },
	"cos":    func(x value) (*apd.Decimal, error) { return periodic("cos", x) },
	"tan":    func(x value) (*apd.Decimal, error) { return periodic("tan", x) },
	"arcsin": func(x value) (*apd.Decimal, error) { return arcsin(x.d) },
	"arccos": func(x value) (*apd.Decimal, error) { return arccos(x.d) },
	"arctan": func(x value) (*apd.Decimal, error) { return arctan(x.d) },
}

// periodicNoiseUlps is how many units in the last working digit a rounded
// sin/cos/tan argument is assumed to be off by. It covers rounding that
// builds up over a few operations, e.g. (pi/7)*7.
const periodicNoiseUlps = 10

// maxPeriodicNoise is the largest uncertainty tolerated in a rounded
// sin/cos/tan argument. Beyond it (e.g. sin(2^200), whose 61-digit
// argument had to be rounded to 60 digits) the result could be anything,
// so an error is returned instead of a confident-looking wrong number.
// Exact arguments, however large (10^100), are always computed.
var maxPeriodicNoise = apd.New(1, -6)

// periodic evaluates sin, cos or tan of x.
//
// If x was rounded (x.inexact, as with pi or pi/2), it differs from the
// value the user meant by a few units in its last working digit. sin and
// cos change by at most that much, so a sin or cos smaller than it is
// indistinguishable from zero and is reported as exactly 0: sin(pi) = 0,
// cos(pi/2) = 0, and tan(pi/2) is out of domain rather than a meaningless
// huge number. Exact arguments, like 10^100, are never adjusted.
func periodic(name string, x value) (*apd.Decimal, error) {
	if !x.d.IsZero() && adjustedExponent(x.d) > maxTrigExponent {
		return nil, i18n.Errorf("%s: argument out of range", name)
	}

	var noise *apd.Decimal
	if x.inexact && !x.d.IsZero() {
		noise = apd.New(periodicNoiseUlps, int32(adjustedExponent(x.d)-workingPrecision+1))
		if noise.Cmp(maxPeriodicNoise) > 0 {
			return nil, i18n.Errorf("%s: argument out of range", name)
		}
	}

	sin, cos, err := sinCos(x.d)
	if err != nil {
		return nil, i18n.WrapErr(err)
	}

	if noise != nil {
		for _, v := range []*apd.Decimal{sin, cos} {
			if new(apd.Decimal).Abs(v).Cmp(noise) <= 0 {
				v.SetInt64(0)
			}
		}
	}

	switch name {
	case "sin":
		return sin, nil
	case "cos":
		return cos, nil
	}
	if cos.IsZero() {
		return nil, i18n.Errorf("%s: argument out of domain", name)
	}
	res := new(apd.Decimal)
	if _, err := defaultContext().Quo(res, sin, cos); err != nil {
		return nil, i18n.WrapErr(err)
	}
	return res, nil
}

// value is an intermediate result. inexact records whether rounding has
// happened anywhere in computing it (apd's Inexact condition), i.e.
// whether d may differ from the exact value of the expression. It's what
// lets sin(pi) be 0 while sin(10^100) is computed exactly.
type value struct {
	d       *apd.Decimal
	inexact bool
}

func eval(ctx *apd.Context, node Node, ans *value) (value, error) {
	switch n := node.(type) {
	case *NumberNode:
		d, cond, err := ctx.NewFromString(n.Value)
		if err != nil {
			return value{}, i18n.Errorf("invalid number %q", n.Value)
		}
		return value{d, cond.Inexact()}, nil

	case *IdentNode:
		switch n.Name {
		case "pi":
			return value{new(apd.Decimal).Set(pi), true}, nil
		case "ans":
			if ans == nil {
				return value{}, i18n.Errorf("ans is not available yet")
			}
			return value{new(apd.Decimal).Set(ans.d), ans.inexact}, nil
		default:
			return value{}, i18n.Errorf("unknown identifier %q", n.Name)
		}

	case *CallNode:
		arg, err := eval(ctx, n.Arg, ans)
		if err != nil {
			return value{}, err
		}

		if fn, ok := decimalFunctions[n.Name]; ok {
			res := new(apd.Decimal)
			cond, err := fn(ctx, res, arg.d)
			if err != nil {
				return value{}, i18n.WrapErr(err)
			}
			return value{res, arg.inexact || cond.Inexact()}, nil
		}

		if fn, ok := trigFunctions[n.Name]; ok {
			res, err := fn(arg)
			if err != nil {
				return value{}, i18n.WrapErr(err)
			}
			// Trig results are irrational except at a few special points,
			// so they're always treated as rounded.
			return value{res, true}, nil
		}

		return value{}, i18n.Errorf("unknown function %q", n.Name)

	case *UnaryNode:
		operand, err := eval(ctx, n.Operand, ans)
		if err != nil {
			return value{}, err
		}
		if n.Op != TokenMinus {
			return value{}, i18n.Errorf("unsupported operator")
		}
		res := new(apd.Decimal)
		if _, err := ctx.Neg(res, operand.d); err != nil {
			return value{}, i18n.WrapErr(err)
		}
		return value{res, operand.inexact}, nil

	case *BinaryNode:
		left, err := eval(ctx, n.Left, ans)
		if err != nil {
			return value{}, err
		}
		right, err := eval(ctx, n.Right, ans)
		if err != nil {
			return value{}, err
		}

		res := new(apd.Decimal)
		var cond apd.Condition
		switch n.Op {
		case TokenPlus:
			cond, err = ctx.Add(res, left.d, right.d)
		case TokenMinus:
			cond, err = ctx.Sub(res, left.d, right.d)
		case TokenStar:
			cond, err = ctx.Mul(res, left.d, right.d)
		case TokenSlash:
			cond, err = ctx.Quo(res, left.d, right.d)
		case TokenPercent:
			// The remainder is smaller than the divisor, so it's exact.
			err = remainder(ctx, res, left.d, right.d)
		case TokenCaret:
			// apd follows the General Decimal Arithmetic spec, under which
			// 0 to a negative power is Infinity with no condition raised,
			// so no trap catches it. 0^-n = 1/0^n, so report it the same
			// way as 1/0.
			if left.d.IsZero() && right.d.Negative && !right.d.IsZero() {
				return value{}, i18n.Errorf("division by zero")
			}
			// apd treats 0^0 as an invalid operation; like most
			// calculators (and Go's math.Pow), use the convention 0^0 = 1.
			if left.d.IsZero() && right.d.IsZero() {
				res.SetInt64(1)
				return value{res, false}, nil
			}
			cond, err = ctx.Pow(res, left.d, right.d)
		default:
			return value{}, i18n.Errorf("unsupported operator")
		}
		if err != nil {
			return value{}, i18n.WrapErr(err)
		}
		return value{res, left.inexact || right.inexact || cond.Inexact()}, nil

	default:
		return value{}, i18n.Errorf("invalid expression")
	}
}

// remainder sets d to x % y: a truncated remainder, as in Go and C (not
// floored, as in Python), so the result has the sign of x: -5 % 3 = -2
// and 5 % -3 = 2. Decimal operands work too (5.5 % 2 = 1.5).
//
// apd's Rem implements the General Decimal Arithmetic "remainder", which
// refuses whenever the integer quotient has more digits than the context
// precision (10^60 % 7 fails with "division impossible"), and fails
// outright when the operands' exponents are too far apart to align. Yet
// the remainder itself is always smaller than |y|, so it fits; only the
// quotient, which is thrown away, is large. So the remainder is computed
// here exactly with big integers instead.
//
// Both operands are scaled to integers over their common exponent:
// x = a*10^e, y = b*10^e, and x % y = (a rem b) * 10^e. Exponents are
// bounded by apd.MaxExponent, so a and b have at most ~200,000 digits.
func remainder(ctx *apd.Context, d, x, y *apd.Decimal) error {
	if y.IsZero() || x.Form != apd.Finite || y.Form != apd.Finite {
		// Let apd report its usual conditions (e.g. 5 % 0 is an invalid
		// operation and 0 % 0 is division undefined).
		_, err := ctx.Rem(d, x, y)
		return err
	}

	e := min(x.Exponent, y.Exponent)
	a := scaledCoefficient(x, int64(x.Exponent)-int64(e))
	b := scaledCoefficient(y, int64(y.Exponent)-int64(e))

	d.Coeff.SetMathBigInt(a.Rem(a, b))
	d.Exponent = e
	d.Negative = x.Negative
	d.Form = apd.Finite
	_, err := ctx.Round(d, d)
	return err
}

// scaledCoefficient returns |coefficient of x| * 10^shift.
func scaledCoefficient(x *apd.Decimal, shift int64) *big.Int {
	c := x.Coeff.MathBigInt()
	if shift > 0 {
		c.Mul(c, new(big.Int).Exp(big.NewInt(10), big.NewInt(shift), nil))
	}
	return c.Abs(c)
}
