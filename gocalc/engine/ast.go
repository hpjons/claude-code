package engine

type Node interface{}

type NumberNode struct {
	Value string
}

type BinaryNode struct {
	Op    TokenType
	Left  Node
	Right Node
}

// UnaryNode is a prefix operator applied to one operand. The parser only
// produces TokenMinus (unary plus is dropped while parsing); eval rejects
// any other Op rather than silently treating it as negation.
type UnaryNode struct {
	Op      TokenType
	Operand Node
}

// IdentNode is a bare identifier reference, e.g. a constant or variable
// such as "pi" or "ans".
type IdentNode struct {
	Name string
}

// CallNode is a single-argument function call, e.g. "sin(x)".
type CallNode struct {
	Name string
	Arg  Node
}
