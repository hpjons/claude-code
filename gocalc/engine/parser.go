package engine

import "gocalc/i18n"

// Precedence levels, higher binds tighter.
// Parentheses are handled structurally and don't need a level here.
const (
	precAdd   = 1 // + -
	precMul   = 2 // * / %
	precUnary = 3 // unary minus
	precPow   = 4 // ^ (right-associative)
)

type parser struct {
	tokens []Token
	pos    int
}

func newParser(tokens []Token) *parser {
	return &parser{tokens: tokens}
}

// Parse parses a full expression and ensures no trailing tokens remain.
func Parse(input string) (Node, error) {
	tokens, err := tokenize(input)
	if err != nil {
		return nil, err
	}
	p := newParser(tokens)
	node, err := p.parseExpression(precAdd)
	if err != nil {
		return nil, err
	}
	if p.peek().Type != TokenEOF {
		return nil, i18n.Errorf("unexpected token %q at position %d", p.peek().Value, p.peek().Pos)
	}
	return node, nil
}

func (p *parser) peek() Token {
	return p.tokens[p.pos]
}

func (p *parser) advance() Token {
	tok := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return tok
}

// expectRParen consumes a closing parenthesis, or reports what was found
// instead so the user can see which token broke the expression.
func (p *parser) expectRParen() error {
	tok := p.peek()
	switch tok.Type {
	case TokenRParen:
		p.advance()
		return nil
	case TokenEOF:
		return i18n.Errorf("missing closing parenthesis")
	default:
		return i18n.Errorf("expected ')' but found %q at position %d", tok.Value, tok.Pos)
	}
}

func binaryInfo(t TokenType) (prec int, rightAssoc bool, ok bool) {
	switch t {
	case TokenPlus, TokenMinus:
		return precAdd, false, true
	case TokenStar, TokenSlash, TokenPercent:
		return precMul, false, true
	case TokenCaret:
		return precPow, true, true
	default:
		return 0, false, false
	}
}

func (p *parser) parseExpression(minPrec int) (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for {
		tok := p.peek()
		prec, rightAssoc, ok := binaryInfo(tok.Type)
		if !ok || prec < minPrec {
			break
		}
		p.advance()

		nextMin := prec
		if !rightAssoc {
			nextMin = prec + 1
		}

		right, err := p.parseExpression(nextMin)
		if err != nil {
			return nil, err
		}
		left = &BinaryNode{Op: tok.Type, Left: left, Right: right}
	}

	return left, nil
}

func (p *parser) parseUnary() (Node, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenMinus:
		p.advance()
		operand, err := p.parseExpression(precUnary)
		if err != nil {
			return nil, err
		}
		return &UnaryNode{Op: TokenMinus, Operand: operand}, nil
	case TokenPlus:
		p.advance()
		return p.parseExpression(precUnary)
	default:
		return p.parsePrimary()
	}
}

func (p *parser) parsePrimary() (Node, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenNumber:
		p.advance()
		return &NumberNode{Value: tok.Value}, nil
	case TokenLParen:
		p.advance()
		inner, err := p.parseExpression(precAdd)
		if err != nil {
			return nil, err
		}
		if err := p.expectRParen(); err != nil {
			return nil, err
		}
		return inner, nil
	case TokenIdent:
		p.advance()
		if p.peek().Type != TokenLParen {
			return &IdentNode{Name: tok.Value}, nil
		}
		p.advance()
		arg, err := p.parseExpression(precAdd)
		if err != nil {
			return nil, err
		}
		if err := p.expectRParen(); err != nil {
			return nil, err
		}
		return &CallNode{Name: tok.Value, Arg: arg}, nil
	case TokenEOF:
		return nil, i18n.Errorf("unexpected end of expression")
	default:
		return nil, i18n.Errorf("unexpected token %q at position %d", tok.Value, tok.Pos)
	}
}
