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
