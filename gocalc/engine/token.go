package engine

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenNumber
	TokenIdent
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenPercent
	TokenCaret
	TokenLParen
	TokenRParen
)

type Token struct {
	Type  TokenType
	Value string
	Pos   int
}
