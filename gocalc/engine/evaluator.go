package engine

import (
	"math"
	"strconv"

	"github.com/cockroachdb/apd/v3"

	"gocalc/i18n"
)

// piDigits is pi to more digits than defaultContext's precision, so it is
// rounded rather than truncated short.
const piDigits = "3.14159265358979323846264338327950288419716939937510582097494459230781640628620899862803482534211706798"

// pi is piDigits rounded to defaultContext's precision, parsed once at
// startup. eval hands out copies so no caller can mutate the shared value.
var pi = func() *apd.Decimal {
	d, _, err := defaultContext().NewFromString(piDigits)
	if err != nil {
		panic(err) // piDigits is a constant; this can't fail
	}
	return d
}()

// defaultContext returns a fresh decimal context with enough precision for
// a calculator REPL, matching standard calculator rounding behavior.
func defaultContext() *apd.Context {
	return apd.BaseContext.WithPrecision(50)
}

// decimalFunctions are single-argument functions apd implements natively
// at arbitrary precision, so no float64 round-trip is needed.
var decimalFunctions = map[string]func(ctx *apd.Context, d, x *apd.Decimal) (apd.Condition, error){
	"sqrt": (*apd.Context).Sqrt,
	"exp":  (*apd.Context).Exp,
}

// float64Function is a single-argument function with no practical pure
// decimal implementation; it's computed via float64, trading arbitrary
// precision for feasibility.
type float64Function struct {
	fn func(float64) float64
	// periodic marks sin/cos/tan: their output error tracks the argument's
	// float64 conversion error (slope near zero crossings is ~1), so
	// results inside that error are snapped to zero, and the result is
	// rejected outright once the argument error is large.
	periodic bool
}

// float64Functions maps function names to their float64 implementations.
// Inverse trig functions are named arcsin/arccos/arctan rather than
// SpeedCrunch's asin/acos/atan, per preference.
var float64Functions = map[string]float64Function{
	"sin":    {math.Sin, true},
	"cos":    {math.Cos, true},
	"tan":    {math.Tan, true},
	"arcsin": {math.Asin, false},
	"arccos": {math.Acos, false},
	"arctan": {math.Atan, false},
}

// float64SignificantDigits is how many significant digits a float64-based
// function result is rounded to. float64 carries ~15.95 decimal digits, so
// anything past 15 is noise that would otherwise be displayed as if it
// were exact.
const float64SignificantDigits = 15

// float64NoiseFactor scales the argument's float64 conversion error into
// the threshold below which a sin/cos/tan result is treated as exactly
// zero. For
// example pi -> float64 is off by ~1.2e-16, so sin(pi) comes back as
// ~1.2e-16 rather than 0; a result no bigger than the error carried in by
// the argument holds no information. (|d/dx| of sin and cos is at most 1,
// so the argument error bounds their output error.)
const float64NoiseFactor = 4

// maxPeriodicArgError is the largest float64 conversion error tolerated in
// a sin/cos/tan argument. Beyond it (e.g. sin(10^100), where the nearest
// float64 is off by ~10^84) the result would be arbitrary, so an error is
// returned instead of a confident-looking wrong number.
const maxPeriodicArgError = 1e-6

func evalFloatFunc(ctx *apd.Context, name string, f float64Function, arg *apd.Decimal) (*apd.Decimal, error) {
	x, err := arg.Float64()
	if err != nil || math.IsInf(x, 0) {
		return nil, i18n.Errorf("%s: argument out of range", name)
	}

	// argErr is how far the float64 argument is from the real decimal one.
	argErr, err := float64ConversionError(ctx, arg, x)
	if err != nil {
		return nil, err
	}
	if f.periodic && argErr > maxPeriodicArgError {
		return nil, i18n.Errorf("%s: argument out of range", name)
	}

	result := f.fn(x)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, i18n.Errorf("%s: argument out of domain", name)
	}
	if f.periodic && math.Abs(result) <= float64NoiseFactor*argErr {
		result = 0
	}

	res := new(apd.Decimal)
	if _, err := res.SetFloat64(result); err != nil {
		return nil, err
	}
	if _, err := apd.BaseContext.WithPrecision(float64SignificantDigits).Round(res, res); err != nil {
		return nil, i18n.WrapErr(err)
	}
	return res, nil
}

// float64ConversionError returns |arg - x| where x is arg converted to
// float64, computed in decimal against x's exact binary value.
//
// apd's SetFloat64 can't be used for x: it takes the shortest decimal that
// round-trips (so float64(1e100) reads back as exactly 1e100, hiding a
// ~1e84 error). Every float64 has a finite exact decimal expansion of at
// most 767 significant digits, which 'e' with that precision yields.
func float64ConversionError(ctx *apd.Context, arg *apd.Decimal, x float64) (float64, error) {
	xDec, _, err := apd.NewFromString(strconv.FormatFloat(x, 'e', 767, 64))
	if err != nil {
		return 0, err
	}
	diff := new(apd.Decimal)
	if _, err := ctx.Sub(diff, arg, xDec); err != nil {
		return 0, i18n.WrapErr(err)
	}
	d, err := diff.Float64()
	if err != nil {
		return math.Inf(1), nil
	}
	return math.Abs(d), nil
}

// Evaluate parses and evaluates a single expression, returning a clean,
// human-readable decimal string. ans is the previous result (empty if
// there isn't one yet) and is made available to the expression as the
// "ans" identifier.
func Evaluate(expression string, ans string) (string, error) {
	node, err := Parse(expression)
	if err != nil {
		return "", err
	}

	ctx := defaultContext()

	var ansDecimal *apd.Decimal
	if ans != "" {
		ansDecimal, _, err = ctx.NewFromString(ans)
		if err != nil || ansDecimal.Form != apd.Finite {
			return "", i18n.Errorf("invalid previous result %q", ans)
		}
	}

	result, err := eval(ctx, node, ansDecimal)
	if err != nil {
		return "", err
	}

	return formatDecimal(ctx, result), nil
}

func eval(ctx *apd.Context, node Node, ans *apd.Decimal) (*apd.Decimal, error) {
	switch n := node.(type) {
	case *NumberNode:
		d, _, err := ctx.NewFromString(n.Value)
		if err != nil {
			return nil, i18n.Errorf("invalid number %q", n.Value)
		}
		return d, nil

	case *IdentNode:
		switch n.Name {
		case "pi":
			return new(apd.Decimal).Set(pi), nil
		case "ans":
			if ans == nil {
				return nil, i18n.Errorf("ans is not available yet")
			}
			return ans, nil
		default:
			return nil, i18n.Errorf("unknown identifier %q", n.Name)
		}

	case *CallNode:
		arg, err := eval(ctx, n.Arg, ans)
		if err != nil {
			return nil, err
		}

		if fn, ok := decimalFunctions[n.Name]; ok {
			res := new(apd.Decimal)
			if _, err := fn(ctx, res, arg); err != nil {
				return nil, i18n.WrapErr(err)
			}
			return res, nil
		}

		if fn, ok := float64Functions[n.Name]; ok {
			return evalFloatFunc(ctx, n.Name, fn, arg)
		}

		return nil, i18n.Errorf("unknown function %q", n.Name)

	case *UnaryNode:
		operand, err := eval(ctx, n.Operand, ans)
		if err != nil {
			return nil, err
		}
		if n.Op != TokenMinus {
			return nil, i18n.Errorf("unsupported operator")
		}
		res := new(apd.Decimal)
		if _, err := ctx.Neg(res, operand); err != nil {
			return nil, i18n.WrapErr(err)
		}
		return res, nil

	case *BinaryNode:
		left, err := eval(ctx, n.Left, ans)
		if err != nil {
			return nil, err
		}
		right, err := eval(ctx, n.Right, ans)
		if err != nil {
			return nil, err
		}

		res := new(apd.Decimal)
		switch n.Op {
		case TokenPlus:
			_, err = ctx.Add(res, left, right)
		case TokenMinus:
			_, err = ctx.Sub(res, left, right)
		case TokenStar:
			_, err = ctx.Mul(res, left, right)
		case TokenSlash:
			_, err = ctx.Quo(res, left, right)
		case TokenPercent:
			_, err = ctx.Rem(res, left, right)
		case TokenCaret:
			// apd follows the General Decimal Arithmetic spec, under which
			// 0 to a negative power is Infinity with no condition raised,
			// so no trap catches it. 0^-n = 1/0^n, so report it the same
			// way as 1/0.
			if left.IsZero() && right.Negative && !right.IsZero() {
				return nil, i18n.Errorf("division by zero")
			}
			_, err = ctx.Pow(res, left, right)
		default:
			return nil, i18n.Errorf("unsupported operator")
		}
		if err != nil {
			return nil, i18n.WrapErr(err)
		}
		return res, nil

	default:
		return nil, i18n.Errorf("invalid expression")
	}
}
