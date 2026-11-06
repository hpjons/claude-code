package engine

import (
	"math"

	"github.com/cockroachdb/apd/v3"

	"gocalc/i18n"
)

// piDigits is pi to more digits than defaultContext's precision, so it is
// rounded rather than truncated short.
const piDigits = "3.14159265358979323846264338327950288419716939937510582097494459230781640628620899862803482534211706798"

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

// float64Functions are single-argument functions with no practical pure
// decimal implementation; they're computed via float64, trading arbitrary
// precision for feasibility. Inverse trig functions are named arcsin/
// arccos/arctan rather than SpeedCrunch's asin/acos/atan, per preference.
var float64Functions = map[string]func(float64) float64{
	"sin":    math.Sin,
	"cos":    math.Cos,
	"tan":    math.Tan,
	"arcsin": math.Asin,
	"arccos": math.Acos,
	"arctan": math.Atan,
}

func evalFloatFunc(name string, fn func(float64) float64, arg *apd.Decimal) (*apd.Decimal, error) {
	f, err := arg.Float64()
	if err != nil {
		return nil, i18n.Errorf("%s: argument out of range", name)
	}

	result := fn(f)
	if math.IsNaN(result) || math.IsInf(result, 0) {
		return nil, i18n.Errorf("%s: argument out of domain", name)
	}

	res := new(apd.Decimal)
	if _, err := res.SetFloat64(result); err != nil {
		return nil, err
	}
	return res, nil
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
		if err != nil {
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
			d, _, err := ctx.NewFromString(piDigits)
			if err != nil {
				return nil, err
			}
			return d, nil
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
			return evalFloatFunc(n.Name, fn, arg)
		}

		return nil, i18n.Errorf("unknown function %q", n.Name)

	case *UnaryNode:
		operand, err := eval(ctx, n.Operand, ans)
		if err != nil {
			return nil, err
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
