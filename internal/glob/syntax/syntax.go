package syntax

import (
	"github.com/lavaorg/telex/internal/glob/syntax/ast"
	"github.com/lavaorg/telex/internal/glob/syntax/lexer"
)

func Parse(s string) (*ast.Node, error) {
	return ast.Parse(lexer.NewLexer(s))
}

func Special(b byte) bool {
	return lexer.Special(b)
}
