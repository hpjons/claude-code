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
