package i18n

import (
	"testing"

	"github.com/leonelquinteros/gotext"
)

type errString string

func (e errString) Error() string { return string(e) }

func loadIcelandicCatalog(t *testing.T) *gotext.Po {
	t.Helper()
	po := gotext.NewPo()
	po.Parse(isPO)
	return po
}

func TestTEnglishPassthrough(t *testing.T) {
	catalog = nil // force English regardless of the test runner's locale
	got := T("unknown function %q", "foo")
	want := `unknown function "foo"`
	if got != want {
		t.Errorf("T() = %q, want %q", got, want)
	}
}

func TestTIcelandic(t *testing.T) {
	po := loadIcelandicCatalog(t)
	catalog = po
	defer func() { catalog = nil }()

	got := T("unknown function %q", "foo")
	want := `óþekkt fall "foo"`
	if got != want {
		t.Errorf("T() = %q, want %q", got, want)
	}
}

func TestWrapErrIcelandic(t *testing.T) {
	po := loadIcelandicCatalog(t)
	catalog = po
	defer func() { catalog = nil }()

	err := WrapErr(errString("division by zero"))
	if err.Error() != "deilt með núlli" {
		t.Errorf("WrapErr() = %q, want %q", err.Error(), "deilt með núlli")
	}
}

func TestWrapErrUnknownPassesThrough(t *testing.T) {
	err := WrapErr(errString("some unrelated library error"))
	if err.Error() != "some unrelated library error" {
		t.Errorf("WrapErr() = %q, want passthrough", err.Error())
	}
}

func TestTIcelandicExpectedRParen(t *testing.T) {
	catalog = loadIcelandicCatalog(t)
	defer func() { catalog = nil }()

	got := T("expected ')' but found %q at position %d", "3", 5)
	want := `bjóst við ')' en fann "3" á stöðu 5`
	if got != want {
		t.Errorf("T() = %q, want %q", got, want)
	}
}

func TestWrapErrShapes(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"exponent out of range", "exponent out of range"},
		// Operation-name prefixes are internal detail and are dropped.
		{"Rem: exponent out of range", "exponent out of range"},
		{"integer power: overflow", "overflow"},
		// apd always pairs subnormal with underflow; only underflow is kept.
		{"underflow, subnormal", "underflow"},
		{"subnormal", "subnormal"},
		{"overflow, division by zero", "overflow, division by zero"},
		// Anything not made up entirely of known conditions passes through
		// unchanged, prefix included.
		{"parse mantissa: abc", "parse mantissa: abc"},
		{"overflow, something new", "overflow, something new"},
	}
	for _, c := range cases {
		if got := WrapErr(errString(c.in)).Error(); got != c.want {
			t.Errorf("WrapErr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestWrapErrIcelandicShapes(t *testing.T) {
	catalog = loadIcelandicCatalog(t)
	defer func() { catalog = nil }()

	cases := []struct {
		in, want string
	}{
		{"exponent out of range", "veldisvísir utan sviðs"},
		{"Rem: exponent out of range", "veldisvísir utan sviðs"},
		{"underflow, subnormal", "undirflæði"},
		{"overflow, division by zero", "yfirflæði, deilt með núlli"},
	}
	for _, c := range cases {
		if got := WrapErr(errString(c.in)).Error(); got != c.want {
			t.Errorf("WrapErr(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestLanguage(t *testing.T) {
	catalog = nil
	if got := Language(); got != "en" {
		t.Errorf("Language() with no catalog = %q, want %q", got, "en")
	}

	catalog = loadIcelandicCatalog(t)
	defer func() { catalog = nil }()
	if got := Language(); got != "is" {
		t.Errorf("Language() with the Icelandic catalog = %q, want %q", got, "is")
	}
}
