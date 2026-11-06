package engine

import (
	"strings"
	"testing"

	"github.com/cockroachdb/apd/v3"
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

func TestEvaluateTrigValues(t *testing.T) {
	// Reference values from mpmath at 1,200 digits, rounded to 50.
	assertEvaluates(t, []struct{ expr, want string }{
		{"sin(1)", "0.84147098480789650665250232163029899962256306079837"},
		{"cos(1)", "0.54030230586813971740093660744297660373231042061792"},
		{"sin(1e-20)", "9.9999999999999999999999999999999999999998333333333e-21"},
		{"sin(10^22)", "-0.85220084976718880177270589375302936826176215041004"},
		{"sin(10^100)", "-0.37237612366127668826208669555316429571966788356743"},
		{"sin(10^1000)", "0.65335979821036985694809946803976857426591654081541"},
		{"sin(2^199)", "-0.96898675455215408154330141770095624584619203465282"},
		{"arctan(1)", "0.78539816339744830961566084581987572104929234984378"},
		{"arcsin(0.5)", "0.52359877559829887307710723054658381403286156656252"},
		{"arccos(0.99999999999999999999)", "0.00000000014142135623730950488028672355116756577770092676309"},
		{"arccos(-1)", "3.1415926535897932384626433832795028841971693993751"},
		{"arctan(10^100)", "1.5707963267948966192313216916397514420985846996876"},
		{"arcsin(1)", "1.5707963267948966192313216916397514420985846996876"},
		{"arccos(1)", "0"},
		{"arcsin(-1) * 2 + pi", "0"},
	})
}

func TestEvaluateTrigExactResults(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		// Results within the rounding error of a rounded argument are 0.
		{"sin(pi)", "0"},
		{"sin(-pi)", "0"},
		{"sin(pi / 7 * 7)", "0"},
		{"sin(100 * pi)", "0"},
		{"cos(pi / 2)", "0"},
		{"tan(pi)", "0"},
		// ...but tiny results of exact arguments are kept.
		{"arctan(1e-30)", "1e-30"},
		// Exact values come out exact.
		{"sin(pi / 6)", "0.5"},
		{"cos(2 * pi)", "1"},
		{"tan(pi / 4)", "1"},
		{"arctan(1) * 4", "3.1415926535897932384626433832795028841971693993751"},
		{"sin(arcsin(0.3))", "0.3"},
	})
}

func TestEvaluateTrigIdentities(t *testing.T) {
	// The working precision's guard digits absorb the rounding in each
	// step, so identities hold exactly at display precision.
	for _, x := range []string{"2", "1", "0.5", "-3.7", "123.456", "1e10", "1e-5", "pi / 3"} {
		expr := "sin(" + x + ")^2 + cos(" + x + ")^2"
		assertEvaluates(t, []struct{ expr, want string }{{expr, "1"}})
	}
}

func TestEvaluateGuardDigits(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		{"1/3*3", "1"},
		{"2/3*3", "2"},
		{"sqrt(2)^2", "2"},
		{"1/7*7", "1"},
	})
}

func TestEvaluateTrigRangeAndDomain(t *testing.T) {
	cases := []struct {
		expr string
		want string
	}{
		{"sin(1e1001)", "sin: argument out of range"},
		{"cos(-1e1001)", "cos: argument out of range"},
		{"tan(pi / 2)", "tan: argument out of domain"},
		// 2^200 has 61 digits, so it's rounded and uncertain by ~100.
		{"sin(2^200)", "sin: argument out of range"},
		{"cos(10^70 + 1)", "cos: argument out of range"},
		{"arcsin(1.0000001)", "arcsin: argument out of domain"},
		{"arccos(-2)", "arccos: argument out of domain"},
	}
	for _, c := range cases {
		_, err := Evaluate(c.expr)
		if err == nil || err.Error() != c.want {
			t.Errorf("Evaluate(%q) error = %v, want %q", c.expr, err, c.want)
		}
	}
}

func TestComputePi(t *testing.T) {
	const want = "3.141592653589793238462643383279502884197169399375105820974944592307816406286208998628034825342117068"
	if got := computePi(100).Text('f'); got != want {
		t.Errorf("computePi(100) = %s, want %s", got, want)
	}
	// Larger requests grow the cache and agree on the shared digits.
	if got := piTo(500).Text('f'); got[:len(want)-1] != want[:len(want)-1] {
		t.Errorf("piTo(500) = %s..., disagrees with the first 100 digits", got[:len(want)])
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
	got.d.SetInt64(0)

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
	got.d.SetInt64(0)

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

func TestFormatDecimalRounding(t *testing.T) {
	cases := []struct{ in, want string }{
		{"0.1234", "0.1234"},
		{"-5.00", "-5"},
		{"1234500", "1234500"},
		{"0.00001", "0.00001"},
		// Half up at the 50th digit, with carries.
		{"1.00000000000000000000000000000000000000000000000005", "1.0000000000000000000000000000000000000000000000001"},
		{"1.00000000000000000000000000000000000000000000000004", "1"},
		{"9.99999999999999999999999999999999999999999999999995", "10"},
		{"-9.99999999999999999999999999999999999999999999999995e-30", "-1e-29"},
		{"99999999999999999999999999999999999999999999999999.5", "1e50"},
	}
	for _, c := range cases {
		d, _, err := apd.NewFromString(c.in)
		if err != nil {
			t.Fatalf("NewFromString(%q): %v", c.in, err)
		}
		if got := formatDecimal(d); got != c.want {
			t.Errorf("formatDecimal(%s) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatDecimalBelowApdExponentLimit(t *testing.T) {
	// A 60-digit coefficient for a value near 1e-99999 has an exponent
	// below apd's limit (calculations produce these, though
	// NewFromString refuses them); it must still format correctly.
	cases := []struct {
		coeff string
		exp   int32
		want  string
	}{
		{"1" + strings.Repeat("0", 59), -100058, "1e-99999"},
		{strings.Repeat("3", 60), -100058, "3.3333333333333333333333333333333333333333333333333e-99999"},
	}
	for _, c := range cases {
		coeff, ok := new(apd.BigInt).SetString(c.coeff, 10)
		if !ok {
			t.Fatalf("bad coefficient %q", c.coeff)
		}
		if got := formatDecimal(apd.NewWithBigInt(coeff, c.exp)); got != c.want {
			t.Errorf("formatDecimal(%se%d) = %q, want %q", c.coeff, c.exp, got, c.want)
		}
	}
}

func TestEvaluateNearMinimumExponent(t *testing.T) {
	assertEvaluates(t, []struct{ expr, want string }{
		{"tan(1e-99999)", "1e-99999"},
		{"sin(1e-99999)", "1e-99999"},
		{"arcsin(1e-99999)", "1e-99999"},
		{"arctan(1e99999) * 2 - pi", "0"},
	})
}
