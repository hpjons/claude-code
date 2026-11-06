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

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

// Parse parses a full expression and ensures no trailing tokens remain.
func Parse(input string) (Node, error) {
	tokens, err := Tokenize(input)
	if err != nil {
		return nil, err
	}
	p := NewParser(tokens)
	node, err := p.parseExpression(precAdd)
	if err != nil {
		return nil, err
	}
	if p.peek().Type != TokenEOF {
		return nil, i18n.Errorf("unexpected token %q at position %d", p.peek().Value, p.peek().Pos)
	}
	return node, nil
}

func (p *Parser) peek() Token {
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.tokens[p.pos]
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return tok
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

func (p *Parser) parseExpression(minPrec int) (Node, error) {
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

func (p *Parser) parseUnary() (Node, error) {
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

func (p *Parser) parsePrimary() (Node, error) {
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
		if p.peek().Type != TokenRParen {
			return nil, i18n.Errorf("missing closing parenthesis")
		}
		p.advance()
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
		if p.peek().Type != TokenRParen {
			return nil, i18n.Errorf("missing closing parenthesis")
		}
		p.advance()
		return &CallNode{Name: tok.Value, Arg: arg}, nil
	case TokenEOF:
		return nil, i18n.Errorf("unexpected end of expression")
	default:
		return nil, i18n.Errorf("unexpected token %q at position %d", tok.Value, tok.Pos)
	}
}
