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
		got, err := Evaluate(c.expr, "")
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
	_, err := Evaluate("5 / 0", "")
	if err == nil {
		t.Fatal("expected an error for division by zero, got nil")
	}
	if !strings.Contains(err.Error(), "division by zero") {
		t.Errorf("expected an informative division-by-zero message, got: %v", err)
	}
}

func TestEvaluateModuloByZero(t *testing.T) {
	_, err := Evaluate("5 % 0", "")
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
		_, err := Evaluate(expr, "")
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
		got, err := Evaluate(c.expr, "")
		if err != nil {
			t.Errorf("Evaluate(%q) returned error: %v", c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("Evaluate(%q) = %q, want %q", c.expr, got, c.want)
		}
	}
}

func TestEvaluateAns(t *testing.T) {
	first, err := Evaluate("2 + 3", "")
	if err != nil {
		t.Fatalf("Evaluate(%q) returned error: %v", "2 + 3", err)
	}
	if first != "5" {
		t.Fatalf("Evaluate(%q) = %q, want %q", "2 + 3", first, "5")
	}

	second, err := Evaluate("ans * 2", first)
	if err != nil {
		t.Fatalf("Evaluate(%q) returned error: %v", "ans * 2", err)
	}
	if second != "10" {
		t.Errorf("Evaluate(%q) = %q, want %q", "ans * 2", second, "10")
	}
}

func TestEvaluateAnsUnavailable(t *testing.T) {
	_, err := Evaluate("ans + 1", "")
	if err == nil {
		t.Fatal("expected an error when ans is not yet available, got nil")
	}
}

func TestEvaluateUnknownIdentifierOrFunction(t *testing.T) {
	exprs := []string{"foo", "bar(1)"}
	for _, expr := range exprs {
		_, err := Evaluate(expr, "")
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", expr)
		}
	}
}

func TestEvaluateTrigDomainErrors(t *testing.T) {
	exprs := []string{"arcsin(2)", "arccos(2)", "arcsin(-2)"}
	for _, expr := range exprs {
		_, err := Evaluate(expr, "")
		if err == nil {
			t.Errorf("Evaluate(%q) expected a domain error, got nil", expr)
		}
	}
}

func assertEvaluates(t *testing.T, cases []struct{ expr, want string }) {
	t.Helper()
	for _, c := range cases {
		got, err := Evaluate(c.expr, "")
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
	got, err := Evaluate("10^99999", "")
	if err != nil {
		t.Fatalf("Evaluate(10^99999) returned error: %v", err)
	}
	if got != "1e99999" {
		t.Errorf("Evaluate(10^99999) = %q, want %q", got, "1e99999")
	}
}

func TestEvaluateScientificResultRoundTripsAsAns(t *testing.T) {
	first, err := Evaluate("10^100", "")
	if err != nil {
		t.Fatalf("Evaluate(10^100) returned error: %v", err)
	}
	got, err := Evaluate("ans / 10^99", first)
	if err != nil {
		t.Fatalf("Evaluate(ans / 10^99) returned error: %v", err)
	}
	if got != "10" {
		t.Errorf("Evaluate(ans / 10^99) = %q, want %q", got, "10")
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
		if _, err := Evaluate(expr, ""); err == nil {
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
		_, err := Evaluate(expr, "")
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
		_, err := Evaluate(c.expr, "")
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
		{"1e99999 % 1e-99999", "exponent out of range"},
	}
	for _, c := range cases {
		_, err := Evaluate(c.expr, "")
		if err == nil {
			t.Errorf("Evaluate(%q) expected an error, got nil", c.expr)
			continue
		}
		if err.Error() != c.want {
			t.Errorf("Evaluate(%q) error = %q, want %q", c.expr, err.Error(), c.want)
		}
	}
}
