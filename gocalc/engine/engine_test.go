package engine

import (
	"strings"
	"testing"
)

func TestEvaluateBasic(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"0.1 + 0.2", "0.3"},
		{"2 + 3 * 4", "14"},
		{"(2 + 3) * 4", "20"},
		{"2 ^ 3 ^ 2", "512"},
		{"2 ** 2", "4"},
		{"2 ** 3 ** 2", "512"},
		{"-2 ** 2", "-4"},
		{"10 / 4", "2.5"},
		{"10 % 3", "1"},
		{"-5 + 3", "-2"},
		{"-2 ^ 2", "-4"},
		{"-2 * 3", "-6"},
		{"5.00 + 0", "5"},
		{"2 + -3", "-1"},
	}

	for _, c := range cases {
		got, err := Evaluate(c.expr)
		if err != nil {
			t.Errorf("Evaluate(%q) returned error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("Evaluate(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestEvaluateDivisionByZero(t *testing.T) {
	_, err := Evaluate("5 / 0")
	if err == nil {
		t.Fatal("expected an error for division by zero, got nil")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("expected an informative division-by-zero message, got: %v", err)
	}
}

func TestEvaluateModuloByZero(t *testing.T) {
	_, err := Evaluate("5 % 0")
	if err == nil {
		t.Fatal("expected an error for modulo by zero, got nil")
	}
}

func TestEvaluateSyntaxErrors(t *testing.T) {
	exprs := []string{
		"2 + * 3",
		"((2+3)",
		"2 + 3)",
		"",
		"2 3",
		"* 2",
	}

	for _, expr := range exprs {
		_, err := Evaluate(expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected a syntax error, got nil", expr)
		}
	}
}

func TestEvaluateConstantsAndFunctions(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"sqrt(4)", "2"},
		{"sqrt(9)", "3"},
		{"sqrt(2 + 2)", "2"},
		{"sin(0)", "0"},
		{"cos(0)", "1"},
		{"tan(0)", "0"},
		{"arcsin(0)", "0"},
		{"arccos(1)", "0"},
		{"arctan(0)", "0"},
		{"exp(0)", "1"},
		{"exp(1)", "2.7182818284590452353602874713526624977572470937"},
		{"pi", "3.1415926535897932384626433832795028841971693993751"},
		{"2 * pi", "6.2831853071795864769252867665590057683943387987502"},
	}

	for _, c := range cases {
		got, err := Evaluate(c.expr)
		if err != nil {
			t.Errorf("Evaluate(%q) returned error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("Evaluate(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

// sessionStep is one expression evaluated in a Session, with its expected
// result, or wantErr if it should fail.
type sessionStep struct {
	expr, want string
	wantErr    bool
}

func runSession(t *testing.T, s *Session, steps []sessionStep) {
	t.Helper()
	for _, step := range steps {
		got, err := s.Evaluate(step.expr)
		switch {
		case step.wantErr && err == nil:
			t.Errorf("Session.Evaluate(%q) = %q, want an error", step.expr, got)
		case !step.wantErr && err != nil:
			t.Errorf("Session.Evaluate(%q) returned error: %v", step.expr, err)
		case !step.wantErr && got != step.want:
			t.Errorf("Session.Evaluate(%q) = %q, want %q", step.expr, got, step.want)
		}
	}
}

func TestSessionAns(t *testing.T) {
	runSession(t, NewSession(), []sessionStep{
		{expr: "ans + 1", wantErr: true}, // no previous result yet
		{expr: "2 + 3", want: "5"},
		{expr: "ans * 2", want: "10"},
		{expr: "ans", want: "10"},
		// A failed evaluation leaves ans unchanged.
		{expr: "1 / 0", wantErr: true},
		{expr: "ans + 1", want: "11"},
	})
}

func TestSessionsAreIndependent(t *testing.T) {
	a, b := NewSession(), NewSession()
	runSession(t, a, []sessionStep{{expr: "7", want: "7"}})
	runSession(t, b, []sessionStep{{expr: "ans", wantErr: true}})
	runSession(t, a, []sessionStep{{expr: "ans", want: "7"}})
}

func TestEvaluateAnsUnavailable(t *testing.T) {
	_, err := Evaluate("ans + 1")
	if err == nil {
		t.Fatal("expected an error when ans is not yet available, got nil")
	}
}

func TestEvaluateUnknownIdentifierOrFunction(t *testing.T) {
	exprs := []string{"foo", "bar(1)"}
	for _, expr := range exprs {
		_, err := Evaluate(expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", expr)
		}
	}
}

func TestEvaluateTrigDomainErrors(t *testing.T) {
	exprs := []string{"arcsin(2)", "arccos(2)", "arcsin(-2)"}
	for _, expr := range exprs {
		_, err := Evaluate(expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected a domain error, got nil", expr)
		}
	}
}

func assertEvaluates(t *testing.T, cases []struct{ expr, want string }) {
	t.Helper()
	for _, c := range cases {
		got, err := Evaluate(c.expr)
		if err != nil {
			t.Errorf("Evaluate(%q) returned error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("Evaluate(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestEvaluateNoNegativeZero(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		{"0 * -1", "0"},
		{"-0", "0"},
		{"-0.000", "0"},
		{"-1 + 1", "0"},
	})
}

func TestEvaluateScientificNotationOutput(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		// Up to the context precision (50 digits) integers stay plain.
		{"10^49", "10000000000000000000000000000000000000000000000000"},
		{"10^50", "1e50"},
		{"10^100", "1e100"},
		{"-10^100", "-1e100"},
		{"2^200", "1.6069380442589902755419620923411626025222029937828e60"},
		{"1e-10", "0.0000000001"},
		{"1e-11", "1e-11"},
		{"1.5e-12", "1.5e-12"},
	})

	// Output must stay short even for very large results.
	got, err := Evaluate("10^99999")
	if err != nil {
		t.Fatalf("Evaluate(10^99999) returned error: %v", err)
	}
	if got != "1e99999" {
		t.Errorf("Evaluate(10^99999) = %q, want %q", got, "1e99999")
	}
}

func TestEvaluateScientificResultParsesAsInput(t *testing.T) {
	// Results are shown in scientific notation; the user must be able to
	// copy one back into the input.
	first, err := Evaluate("10^100")
	if err != nil {
		t.Fatalf("Evaluate(10^100) returned error: %v", err)
	}
	expr := first + " / 10^99"
	got, err := Evaluate(expr)
	if err != nil {
		t.Fatalf("Evaluate(%q) returned error: %v", expr, err)
	}
	if got != "10" {
		t.Errorf("Evaluate(%q) = %q, want %q", expr, got, "10")
	}
}

func TestEvaluateScientificNotationInput(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		{"1e5", "100000"},
		{"1E5", "100000"},
		{"2.5e+3", "2500"},
		{"1e-3", "0.001"},
		{".5e1", "5"},
		{"2e3 * 2", "4000"},
	})

	// An 'e' with no exponent digits isn't part of the number.
	for _, expr := range []string{"1e", "1e+", "2e x"} {
		if _, err := Evaluate(expr); err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", expr)
		}
	}
}

func TestEvaluateFloat64FunctionsPrecision(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		// Results within the argument's float64 conversion error are 0.
		{"sin(pi)", "0"},
		{"sin(-pi)", "0"},
		{"sin(100 * pi)", "0"},
		{"cos(pi / 2)", "0"},
		{"tan(pi)", "0"},
		// ...but genuinely tiny results are kept.
		{"sin(1e-20)", "1e-20"},
		{"arctan(1e-30)", "1e-30"},
		// Results are rounded to the 15 digits float64 can vouch for.
		{"sin(1)", "0.841470984807897"},
		{"sin(pi / 6)", "0.5"},
		{"cos(2 * pi)", "1"},
		{"arccos(-1)", "3.14159265358979"},
		{"arctan(10^100)", "1.5707963267949"},
		// Large arguments that float64 represents exactly still work.
		{"sin(10^22)", "-0.852200849767189"},
	})
}

func TestEvaluateTrigArgumentTooImprecise(t *testing.T) {
	// float64(10^23) and float64(10^100) are off by millions or more, so
	// any sin/cos/tan result would be arbitrary.
	for _, expr := range []string{"sin(10^23)", "cos(10^100)", "tan(10^100)", "sin(1e400)"} {
		_, err := Evaluate(expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", expr)
			continue
		}
		if !strings.Contains(err.Error(), "argument out of range") {
			t.Errorf("Evaluate(%q) error = %q, want an argument-out-of-range error", expr, err)
		}
	}
}

func TestEvaluateErrorMessages(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"(1 + 2", "missing closing parenthesis"},
		{"sin(1 + 2", "missing closing parenthesis"},
		{"(1 + 2 3", `expected ')' but found "3" at position 7`},
		{"sin(1 2)", `expected ')' but found "2" at position 6`},
		{"5٣", `unexpected character '٣' at position 1`},
		{"1e", `unexpected token "e" at position 1`},
	}
	for _, c := range cases {
		_, err := Evaluate(c.expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", c.expr)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("Evaluate(%q) error = %q, want %q", c.expr, err.Error(), c.want)
		}
	}
}

func TestEvaluateLibraryErrorMessages(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"2^1000000", "exponent out of range"},
		{"exp(-100000)", "underflow"},
	}
	for _, c := range cases {
		_, err := Evaluate(c.expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", c.expr)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("Evaluate(%q) error = %q, want %q", c.expr, err.Error(), c.want)
		}
	}
}

func TestEvaluateZeroToNegativePower(t *testing.T) {
	for _, expr := range []string{"0^-1", "0^-0.5", "0 ** -2", "(-0)^-1", "1 / (0^-1)", "-(0^-1)"} {
		_, err := Evaluate(expr)
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", expr)
			continue
		}
		if err.Error() != "division by zero" {
			t.Errorf("Evaluate(%q) error = %q, want %q", expr, err.Error(), "division by zero")
		}
	}

	// Zero to a positive power is still fine.
	assertEvaluates(t, []struct{ expr, want string }{
		{"0^2", "0"},
		{"0^0.5", "0"},
	})
}

func TestEvalRejectsUnknownUnaryOperator(t *testing.T) {
	node := &UnaryNode{Op: TokenStar, Operand: &NumberNode{Value: "2"}}
	if _, err := eval(defaultContext(), node, nil); err == nil {
		t.Error("eval of a unary '*' expected an error, got nil")
	}
}

func TestEvalPiReturnsCopy(t *testing.T) {
	want := pi.Text('f')

	got, err := eval(defaultContext(), &IdentNode{Name: "pi"}, nil)
	if err != nil {
		t.Fatalf("eval(pi) returned error: %v", err)
	}
	got.SetInt64(0)

	if pi.Text('f') != want {
		t.Errorf("mutating eval's result changed the cached pi to %s", pi.Text('f'))
	}
}

func TestSessionAnsNotAliased(t *testing.T) {
	s := NewSession()
	runSession(t, s, []sessionStep{{expr: "5", want: "5"}})

	got, err := eval(defaultContext(), &IdentNode{Name: "ans"}, s.ans)
	if err != nil {
		t.Fatalf("eval(ans) returned error: %v", err)
	}
	got.SetInt64(0)

	runSession(t, s, []sessionStep{{expr: "ans", want: "5"}})
}

func TestEvaluateModuloSign(t *testing.T) {
	// % is a truncated remainder: the result takes the sign of the left
	// operand (as in Go and C, unlike Python's floored modulo).
	assertEvaluates(t, []struct{ expr, want string }{
		{"5 % 3", "2"},
		{"-5 % 3", "-2"},
		{"5 % -3", "2"},
		{"-5 % -3", "-2"},
		{"-6 % 3", "0"},
		{"5.5 % 2", "1.5"},
		{"-5.5 % 2", "-1.5"},
		{"1e50 % 7", "2"},
	})
}

func TestEvaluateModuloLargeQuotient(t *testing.T) {
	// The integer quotient can have far more digits than the 50-digit
	// precision; only the remainder has to fit, and it's computed exactly.
	assertEvaluates(t, []struct{ expr, want string }{
		{"10^60 % 7", "1"},
		{"-10^60 % 7", "-1"},
		{"10^60 % -7", "1"},
		{"1e99999 % 7", "6"},
		{"1e99999 % 1e-99999", "0"},
		{"1e-99999 % 3", "1e-99999"},
		{"123456789e40 % 0.001", "0"},
	})
}

func TestSessionReset(t *testing.T) {
	s := NewSession()
	runSession(t, s, []sessionStep{{expr: "5", want: "5"}})
	s.Reset()
	runSession(t, s, []sessionStep{
		{expr: "ans", wantErr: true},
		{expr: "2", want: "2"},
		{expr: "ans + 1", want: "3"},
	})
}

func TestEvaluateZeroToZero(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		{"0^0", "1"},
		{"0 ** 0", "1"},
		{"0.0 ^ 0.00", "1"},
		{"(-0)^0", "1"},
		{"-0^0", "-1"}, // unary minus binds looser than ^: -(0^0)
		{"(1 - 1)^(2 - 2)", "1"},
	})
}
