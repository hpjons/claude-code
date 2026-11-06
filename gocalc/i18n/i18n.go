// Package i18n provides translation for every user-facing string in the
// app, backed by a single PO catalog (../locales/is.po). The PO file's
// msgid is always the English source text, used verbatim as the fallback
// when no translation is loaded or a string isn't in the catalog — so
// English never needs its own catalog file.
package i18n

import (
	_ "embed"
	"errors"
	"os"
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
// one of apd's small fixed set of decimal-condition strings — passing it
// through unchanged if it isn't one of those known, recognized messages.
// The condition strings are matched against constant literals (rather
// than looking up err.Error() directly) so each msgid stays a compile-time
// constant for go vet's printf check.
func WrapErr(err error) error {
	if err == nil {
		return nil
	}
	switch err.Error() {
	case "division by zero":
		return errors.New(T("division by zero"))
	case "invalid operation":
		return errors.New(T("invalid operation"))
	case "division undefined":
		return errors.New(T("division undefined"))
	case "division impossible":
		return errors.New(T("division impossible"))
	case "overflow":
		return errors.New(T("overflow"))
	case "underflow":
		return errors.New(T("underflow"))
	case "subnormal":
		return errors.New(T("subnormal"))
	default:
		return err
	}
}
