package engine

import "github.com/cockroachdb/apd/v3"

// Session evaluates a sequence of expressions, making each successful
// result available to the next one as "ans". The previous result is kept
// as a decimal, so it's never formatted to text and parsed back.
//
// A Session is not safe for concurrent use.
type Session struct {
	ans *apd.Decimal // nil until the first successful evaluation
}

// NewSession returns a Session with no previous result.
func NewSession() *Session {
	return &Session{}
}

// Reset forgets the previous result, so "ans" is unavailable until the
// next successful evaluation.
func (s *Session) Reset() {
	s.ans = nil
}

// Evaluate parses and evaluates one expression, returning a clean,
// human-readable decimal string. On success the result becomes the
// session's "ans"; on error ans is left unchanged.
func (s *Session) Evaluate(expression string) (string, error) {
	node, err := Parse(expression)
	if err != nil {
		return "", err
	}

	ctx := defaultContext()
	result, err := eval(ctx, node, s.ans)
	if err != nil {
		return "", err
	}

	s.ans = result
	return formatDecimal(ctx, result), nil
}

// Evaluate evaluates a single expression on its own, with no previous
// result: "ans" is reported as not available yet. Use a Session to chain
// evaluations.
func Evaluate(expression string) (string, error) {
	return NewSession().Evaluate(expression)
}
