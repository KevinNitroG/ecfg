// Package parser implements the EditorConfig file parser with precise position tracking.
//
// The parser uses a two-stage approach:
//  1. Lexer (tokenizer) - scans source text into tokens with positions
//  2. Parser - builds an AST from the token stream
//
// Every token and AST node includes precise position information (line, column, byte offset)
// to support LSP features like hover, diagnostics, and completion.
//
// Token Vocabulary:
//   - TokenEOF: End of file marker
//   - TokenComment: Comment line (# or ; prefix)
//   - TokenSectionStart: Opening bracket [
//   - TokenSectionEnd: Closing bracket ]
//   - TokenIdentifier: Property key or glob pattern content
//   - TokenEquals: Assignment operator =
//   - TokenValue: Property value after =
//   - TokenNewline: Line break (\n or \r\n)
package parser

import "fmt"

// Position represents a location in the source text.
// It tracks byte offset, line number, and column number following LSP conventions:
//   - Offset: 0-indexed byte position in source
//   - Line: 1-indexed line number (LSP standard)
//   - Column: 0-indexed column number (LSP standard)
type Position struct {
	Offset int // 0-indexed byte offset in source
	Line   int // 1-indexed line number
	Column int // 0-indexed column number
}

// String returns a human-readable representation of the position.
func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d (offset %d)", p.Line, p.Column, p.Offset)
}

// Range represents a span of text in the source.
// It is defined by start and end positions.
type Range struct {
	Start Position
	End   Position
}

// String returns a human-readable representation of the range.
func (r Range) String() string {
	if r.Start.Line == r.End.Line {
		return fmt.Sprintf("line %d, columns %d-%d", r.Start.Line, r.Start.Column, r.End.Column)
	}
	return fmt.Sprintf("lines %d-%d", r.Start.Line, r.End.Line)
}

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF          TokenType = iota // End of file
	TokenComment                       // Comment line (# or ; prefix)
	TokenSectionStart                  // Opening bracket [
	TokenSectionEnd                    // Closing bracket ]
	TokenIdentifier                    // Property key or glob pattern content
	TokenEquals                        // Assignment operator =
	TokenValue                         // Property value after =
	TokenNewline                       // Line break
)

// String returns the name of the token type.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenComment:
		return "COMMENT"
	case TokenSectionStart:
		return "SECTION_START"
	case TokenSectionEnd:
		return "SECTION_END"
	case TokenIdentifier:
		return "IDENTIFIER"
	case TokenEquals:
		return "EQUALS"
	case TokenValue:
		return "VALUE"
	case TokenNewline:
		return "NEWLINE"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", int(t))
	}
}

// Token represents a lexical token with its type, raw text, and position.
type Token struct {
	Type  TokenType // Type of the token
	Value string    // Raw text (including whitespace)
	Range Range     // Position in source
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	if len(t.Value) > 20 {
		return fmt.Sprintf("%s(%q...) at %s", t.Type, t.Value[:20], t.Range)
	}
	return fmt.Sprintf("%s(%q) at %s", t.Type, t.Value, t.Range)
}
