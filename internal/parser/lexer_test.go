package parser

import (
	"testing"
)

// TestLexer tests the lexer's ability to tokenize EditorConfig source into a token stream.
func TestLexer(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		tokens []Token
	}{
		{
			name:  "comment with hash",
			input: "# hello",
			tokens: []Token{
				{Type: TokenComment, Value: "# hello", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 7, Line: 1, Column: 7},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 7, Line: 1, Column: 7},
					End:   Position{Offset: 7, Line: 1, Column: 7},
				}},
			},
		},
		{
			name:  "comment with semicolon",
			input: "; hello",
			tokens: []Token{
				{Type: TokenComment, Value: "; hello", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 7, Line: 1, Column: 7},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 7, Line: 1, Column: 7},
					End:   Position{Offset: 7, Line: 1, Column: 7},
				}},
			},
		},
		{
			name:  "section header",
			input: "[*.go]",
			tokens: []Token{
				{Type: TokenSectionStart, Value: "[", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 1, Line: 1, Column: 1},
				}},
				{Type: TokenIdentifier, Value: "*.go", Range: Range{
					Start: Position{Offset: 1, Line: 1, Column: 1},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenSectionEnd, Value: "]", Range: Range{
					Start: Position{Offset: 5, Line: 1, Column: 5},
					End:   Position{Offset: 6, Line: 1, Column: 6},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 6, Line: 1, Column: 6},
					End:   Position{Offset: 6, Line: 1, Column: 6},
				}},
			},
		},
		{
			name:  "key-value pair",
			input: "indent_style = tab",
			tokens: []Token{
				{Type: TokenIdentifier, Value: "indent_style", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 12, Line: 1, Column: 12},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 13, Line: 1, Column: 13},
					End:   Position{Offset: 14, Line: 1, Column: 14},
				}},
				{Type: TokenValue, Value: "tab", Range: Range{
					Start: Position{Offset: 15, Line: 1, Column: 15},
					End:   Position{Offset: 18, Line: 1, Column: 18},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 18, Line: 1, Column: 18},
					End:   Position{Offset: 18, Line: 1, Column: 18},
				}},
			},
		},
		{
			name:  "newline LF",
			input: "key = value\n",
			tokens: []Token{
				{Type: TokenIdentifier, Value: "key", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 3, Line: 1, Column: 3},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 4, Line: 1, Column: 4},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenValue, Value: "value", Range: Range{
					Start: Position{Offset: 6, Line: 1, Column: 6},
					End:   Position{Offset: 11, Line: 1, Column: 11},
				}},
				{Type: TokenNewline, Value: "\n", Range: Range{
					Start: Position{Offset: 11, Line: 1, Column: 11},
					End:   Position{Offset: 12, Line: 2, Column: 0},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 12, Line: 2, Column: 0},
					End:   Position{Offset: 12, Line: 2, Column: 0},
				}},
			},
		},
		{
			name:  "newline CRLF",
			input: "key = value\r\n",
			tokens: []Token{
				{Type: TokenIdentifier, Value: "key", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 3, Line: 1, Column: 3},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 4, Line: 1, Column: 4},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenValue, Value: "value", Range: Range{
					Start: Position{Offset: 6, Line: 1, Column: 6},
					End:   Position{Offset: 11, Line: 1, Column: 11},
				}},
				{Type: TokenNewline, Value: "\r\n", Range: Range{
					Start: Position{Offset: 11, Line: 1, Column: 11},
					End:   Position{Offset: 13, Line: 2, Column: 0},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 13, Line: 2, Column: 0},
					End:   Position{Offset: 13, Line: 2, Column: 0},
				}},
			},
		},
		{
			name:  "unclosed section header",
			input: "[*.go",
			tokens: []Token{
				{Type: TokenSectionStart, Value: "[", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 1, Line: 1, Column: 1},
				}},
				{Type: TokenIdentifier, Value: "*.go", Range: Range{
					Start: Position{Offset: 1, Line: 1, Column: 1},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 5, Line: 1, Column: 5},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
			},
		},
		{
			name:  "UTF-8 multi-byte characters",
			input: "# 你好",
			tokens: []Token{
				{Type: TokenComment, Value: "# 你好", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 8, Line: 1, Column: 3},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 8, Line: 1, Column: 3},
					End:   Position{Offset: 8, Line: 1, Column: 3},
				}},
			},
		},
		{
			name:  "empty input",
			input: "",
			tokens: []Token{
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 0, Line: 1, Column: 0},
				}},
			},
		},
		{
			name:  "whitespace only",
			input: "   ",
			tokens: []Token{
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 3, Line: 1, Column: 3},
					End:   Position{Offset: 3, Line: 1, Column: 3},
				}},
			},
		},
		{
			name:  "key without value",
			input: "key =",
			tokens: []Token{
				{Type: TokenIdentifier, Value: "key", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 3, Line: 1, Column: 3},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 4, Line: 1, Column: 4},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 5, Line: 1, Column: 5},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
			},
		},
		{
			name:  "multiline document",
			input: "[*.go]\nindent_style = tab\n# comment",
			tokens: []Token{
				{Type: TokenSectionStart, Value: "[", Range: Range{
					Start: Position{Offset: 0, Line: 1, Column: 0},
					End:   Position{Offset: 1, Line: 1, Column: 1},
				}},
				{Type: TokenIdentifier, Value: "*.go", Range: Range{
					Start: Position{Offset: 1, Line: 1, Column: 1},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenSectionEnd, Value: "]", Range: Range{
					Start: Position{Offset: 5, Line: 1, Column: 5},
					End:   Position{Offset: 6, Line: 1, Column: 6},
				}},
				{Type: TokenNewline, Value: "\n", Range: Range{
					Start: Position{Offset: 6, Line: 1, Column: 6},
					End:   Position{Offset: 7, Line: 2, Column: 0},
				}},
				{Type: TokenIdentifier, Value: "indent_style", Range: Range{
					Start: Position{Offset: 7, Line: 2, Column: 0},
					End:   Position{Offset: 19, Line: 2, Column: 12},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 20, Line: 2, Column: 13},
					End:   Position{Offset: 21, Line: 2, Column: 14},
				}},
				{Type: TokenValue, Value: "tab", Range: Range{
					Start: Position{Offset: 22, Line: 2, Column: 15},
					End:   Position{Offset: 25, Line: 2, Column: 18},
				}},
				{Type: TokenNewline, Value: "\n", Range: Range{
					Start: Position{Offset: 25, Line: 2, Column: 18},
					End:   Position{Offset: 26, Line: 3, Column: 0},
				}},
				{Type: TokenComment, Value: "# comment", Range: Range{
					Start: Position{Offset: 26, Line: 3, Column: 0},
					End:   Position{Offset: 35, Line: 3, Column: 9},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 35, Line: 3, Column: 9},
					End:   Position{Offset: 35, Line: 3, Column: 9},
				}},
			},
		},
		{
			name:  "leading whitespace before key",
			input: "  key = value",
			tokens: []Token{
				{Type: TokenIdentifier, Value: "key", Range: Range{
					Start: Position{Offset: 2, Line: 1, Column: 2},
					End:   Position{Offset: 5, Line: 1, Column: 5},
				}},
				{Type: TokenEquals, Value: "=", Range: Range{
					Start: Position{Offset: 6, Line: 1, Column: 6},
					End:   Position{Offset: 7, Line: 1, Column: 7},
				}},
				{Type: TokenValue, Value: "value", Range: Range{
					Start: Position{Offset: 8, Line: 1, Column: 8},
					End:   Position{Offset: 13, Line: 1, Column: 13},
				}},
				{Type: TokenEOF, Value: "", Range: Range{
					Start: Position{Offset: 13, Line: 1, Column: 13},
					End:   Position{Offset: 13, Line: 1, Column: 13},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer([]byte(tt.input))

			for i, expected := range tt.tokens {
				got := lexer.Scan()

				if got.Type != expected.Type {
					t.Errorf("token %d: Type = %v, want %v", i, got.Type, expected.Type)
				}
				if got.Value != expected.Value {
					t.Errorf("token %d: Value = %q, want %q", i, got.Value, expected.Value)
				}
				if got.Range.Start != expected.Range.Start {
					t.Errorf("token %d: Start = %v, want %v", i, got.Range.Start, expected.Range.Start)
				}
				if got.Range.End != expected.Range.End {
					t.Errorf("token %d: End = %v, want %v", i, got.Range.End, expected.Range.End)
				}
			}
		})
	}
}

// TestLexerPositionTracking tests that position tracking is accurate across various inputs.
func TestLexerPositionTracking(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, *Lexer)
	}{
		{
			name:  "position after first char",
			input: "abc",
			check: func(t *testing.T, lexer *Lexer) {
				tok := lexer.Scan()
				if tok.Range.Start.Offset != 0 {
					t.Errorf("Start.Offset = %d, want 0", tok.Range.Start.Offset)
				}
				if tok.Range.Start.Line != 1 {
					t.Errorf("Start.Line = %d, want 1", tok.Range.Start.Line)
				}
				if tok.Range.Start.Column != 0 {
					t.Errorf("Start.Column = %d, want 0", tok.Range.Start.Column)
				}
			},
		},
		{
			name:  "position after newline",
			input: "a\nb",
			check: func(t *testing.T, lexer *Lexer) {
				lexer.Scan()        // 'a'
				lexer.Scan()        // '\n'
				tok := lexer.Scan() // 'b'
				if tok.Range.Start.Line != 2 {
					t.Errorf("After newline: Line = %d, want 2", tok.Range.Start.Line)
				}
				if tok.Range.Start.Column != 0 {
					t.Errorf("After newline: Column = %d, want 0", tok.Range.Start.Column)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer([]byte(tt.input))
			tt.check(t, lexer)
		})
	}
}
