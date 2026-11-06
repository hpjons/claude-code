package engine

import (
	"unicode"

	"gocalc/i18n"
)

type lexer struct {
	input []rune
	pos   int
}

func newLexer(input string) *lexer {
	return &lexer{input: []rune(input)}
}

func (l *lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *lexer) next() (Token, error) {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}

	start := l.pos
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Pos: start}, nil
	}

	c := l.input[l.pos]

	switch c {
	case '+':
		l.pos++
		return Token{Type: TokenPlus, Value: "+", Pos: start}, nil
	case '-':
		l.pos++
		return Token{Type: TokenMinus, Value: "-", Pos: start}, nil
	case '*':
		l.pos++
		if l.peek() == '*' {
			l.pos++
			return Token{Type: TokenCaret, Value: "**", Pos: start}, nil
		}
		return Token{Type: TokenStar, Value: "*", Pos: start}, nil
	case '/':
		l.pos++
		return Token{Type: TokenSlash, Value: "/", Pos: start}, nil
	case '%':
		l.pos++
		return Token{Type: TokenPercent, Value: "%", Pos: start}, nil
	case '^':
		l.pos++
		return Token{Type: TokenCaret, Value: "^", Pos: start}, nil
	case '(':
		l.pos++
		return Token{Type: TokenLParen, Value: "(", Pos: start}, nil
	case ')':
		l.pos++
		return Token{Type: TokenRParen, Value: ")", Pos: start}, nil
	}

	if isDigit(c) || c == '.' {
		return l.lexNumber(start)
	}

	if unicode.IsLetter(c) {
		return l.lexIdent(start)
	}

	return Token{}, i18n.Errorf("unexpected character %q at position %d", c, start)
}

func (l *lexer) lexNumber(start int) (Token, error) {
	sawDigit := false
	sawDot := false

	for l.pos < len(l.input) {
		c := l.input[l.pos]
		if isDigit(c) {
			sawDigit = true
			l.pos++
			continue
		}
		if c == '.' {
			if sawDot {
				return Token{}, i18n.Errorf("malformed number %q at position %d: multiple decimal points", string(l.input[start:l.pos+1]), start)
			}
			sawDot = true
			l.pos++
			continue
		}
		break
	}

	if !sawDigit {
		return Token{}, i18n.Errorf("malformed number %q at position %d", string(l.input[start:l.pos]), start)
	}

	// Optional exponent: e/E, optional sign, then at least one digit. If no
	// digit follows, the 'e' is left alone so it lexes as an identifier.
	if c := l.peek(); c == 'e' || c == 'E' {
		i := l.pos + 1
		if i < len(l.input) && (l.input[i] == '+' || l.input[i] == '-') {
			i++
		}
		if i < len(l.input) && isDigit(l.input[i]) {
			for i < len(l.input) && isDigit(l.input[i]) {
				i++
			}
			l.pos = i
		}
	}

	return Token{Type: TokenNumber, Value: string(l.input[start:l.pos]), Pos: start}, nil
}

// isDigit reports whether c is an ASCII decimal digit. unicode.IsDigit is
// deliberately avoided: it accepts digits from other scripts (e.g. '٣'),
// which the decimal library can't parse.
func isDigit(c rune) bool {
	return c >= '0' && c <= '9'
}

func (l *lexer) lexIdent(start int) (Token, error) {
	for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || isDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
		l.pos++
	}
	return Token{Type: TokenIdent, Value: string(l.input[start:l.pos]), Pos: start}, nil
}

// tokenize converts an input expression into a flat token stream, always
// terminated by a TokenEOF.
func tokenize(input string) ([]Token, error) {
	l := newLexer(input)
	var tokens []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			return tokens, nil
		}
	}
}
