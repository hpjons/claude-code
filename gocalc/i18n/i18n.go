// Package i18n provides translation for every user-facing string in the
// app, backed by a single PO catalog (locales/is.po, embedded below). The PO file's
// msgid is always the English source text, used verbatim as the fallback
// when no translation is loaded or a string isn't in the catalog — so
// English never needs its own catalog file.
package i18n

import (
	_ "embed"
	"errors"
	"os"
	"slices"
	"strings"

	"github.com/leonelquinteros/gotext"
)

//go:embed locales/is.po
var isPO []byte

var catalog *gotext.Po

func init() {
	if detectLocale() == "is" {
		po := gotext.NewPo()
		po.Parse(isPO)
		catalog = po
	}
}

// detectLocale resolves the active locale from the standard POSIX
// precedence (LC_ALL > LC_MESSAGES > LANG), returning "is" for Icelandic
// and "en" for everything else, including unset/unparseable values.
func detectLocale() string {
	for _, env := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := os.Getenv(env); v != "" {
			if strings.HasPrefix(strings.ToLower(v), "is") {
				return "is"
			}
			return "en"
		}
	}
	return "en"
}

// Language returns the BCP 47 tag of the language T currently produces
// text in: "is" when the Icelandic catalog is loaded, otherwise "en".
// The frontend uses it to set <html lang>.
func Language() string {
	if catalog != nil {
		return "is"
	}
	return "en"
}

// T translates msgid (English source text, using fmt verbs like %q/%s/%d
// for any args) into the active locale, falling back to the formatted
// English text if no translation is loaded or found.
func T(msgid string, args ...interface{}) string {
	if catalog == nil {
		return gotext.NewPo().Get(msgid, args...)
	}
	return catalog.Get(msgid, args...)
}

// Errorf is a drop-in replacement for fmt.Errorf that translates msgid
// before wrapping it in an error.
func Errorf(msgid string, args ...interface{}) error {
	return errors.New(T(msgid, args...))
}

// WrapErr translates a third-party error's message text — in practice,
// one of apd's decimal-condition strings — passing the error through
// unchanged if its message isn't made up entirely of known conditions.
//
// apd messages come in a few shapes, all handled here:
//   - a single condition: "division by zero"
//   - several conditions joined by ", ": "underflow, subnormal"
//   - a condition prefixed with an internal operation name:
//     "Rem: exponent out of range"
//
// The operation prefix is dropped, since it means nothing to a user, and
// "subnormal" is dropped when "underflow" is present, because apd always
// raises the two together and the second adds nothing.
func WrapErr(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()
	if i := strings.LastIndex(msg, ": "); i >= 0 {
		msg = msg[i+2:]
	}

	conditions := strings.Split(msg, ", ")
	if slices.Contains(conditions, "underflow") {
		conditions = slices.DeleteFunc(conditions, func(c string) bool { return c == "subnormal" })
	}

	translated := make([]string, len(conditions))
	for i, c := range conditions {
		t, ok := translateCondition(c)
		if !ok {
			return err
		}
		translated[i] = t
	}
	return errors.New(strings.Join(translated, ", "))
}

// translateCondition translates one apd condition string. Each case calls
// T with a constant literal (rather than T(c)) so every msgid stays a
// compile-time constant for go vet's printf check.
func translateCondition(c string) (string, bool) {
	switch c {
	case "division by zero":
		return T("division by zero"), true
	case "invalid operation":
		return T("invalid operation"), true
	case "division undefined":
		return T("division undefined"), true
	case "division impossible":
		return T("division impossible"), true
	case "overflow":
		return T("overflow"), true
	case "underflow":
		return T("underflow"), true
	case "subnormal":
		return T("subnormal"), true
	case "exponent out of range":
		return T("exponent out of range"), true
	default:
		return "", false
	}
}
