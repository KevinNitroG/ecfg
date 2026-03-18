package parser

import (
	"unicode/utf8"
)

// Lexer scans EditorConfig source text into a stream of tokens.
// It tracks precise position information (byte offset, line, column) for each token
// to support LSP features.
//
// The lexer handles:
//   - Comments (# or ; prefix)
//   - Section headers ([glob])
//   - Key-value pairs (key = value)
//   - Newlines (LF and CRLF)
//   - Malformed input without panicking
//
// Position tracking follows LSP conventions:
//   - Line numbers are 1-indexed
//   - Column numbers are 0-indexed
//   - Offset is 0-indexed byte position
type Lexer struct {
	source []byte // Input source text
	pos    int    // Current byte offset
	line   int    // Current line (1-indexed)
	col    int    // Current column (0-indexed, in runes)
}

// NewLexer creates a new lexer for the given source text.
func NewLexer(source []byte) *Lexer {
	return &Lexer{
		source: source,
		pos:    0,
		line:   1,
		col:    0,
	}
}

// Scan returns the next token from the input.
// It advances the lexer position and returns TokenEOF when the end is reached.
func (l *Lexer) Scan() Token {
	// Skip leading whitespace (but not newlines)
	l.skipWhitespace()

	// Check for EOF
	if l.pos >= len(l.source) {
		return Token{
			Type:  TokenEOF,
			Value: "",
			Range: Range{
				Start: Position{Offset: l.pos, Line: l.line, Column: l.col},
				End:   Position{Offset: l.pos, Line: l.line, Column: l.col},
			},
		}
	}

	start := Position{Offset: l.pos, Line: l.line, Column: l.col}

	ch := l.source[l.pos]

	// Handle single-character tokens
	switch ch {
	case '[':
		l.advance()
		return Token{
			Type:  TokenSectionStart,
			Value: "[",
			Range: Range{Start: start, End: l.currentPos()},
		}

	case ']':
		l.advance()
		return Token{
			Type:  TokenSectionEnd,
			Value: "]",
			Range: Range{Start: start, End: l.currentPos()},
		}

	case '=':
		l.advance()
		return Token{
			Type:  TokenEquals,
			Value: "=",
			Range: Range{Start: start, End: l.currentPos()},
		}

	case '\n':
		l.advance()
		l.line++
		l.col = 0
		return Token{
			Type:  TokenNewline,
			Value: "\n",
			Range: Range{Start: start, End: l.currentPos()},
		}

	case '\r':
		// Handle CRLF
		l.advance()
		if l.pos < len(l.source) && l.source[l.pos] == '\n' {
			l.advance()
			l.line++
			l.col = 0
			return Token{
				Type:  TokenNewline,
				Value: "\r\n",
				Range: Range{Start: start, End: l.currentPos()},
			}
		}
		// Treat standalone \r as newline
		l.line++
		l.col = 0
		return Token{
			Type:  TokenNewline,
			Value: "\r",
			Range: Range{Start: start, End: l.currentPos()},
		}

	case '#', ';':
		// Comment - read to end of line
		return l.scanComment()
	}

	// After '=' we expect a value
	// After '[' we expect section content (identifier)
	// At line start or after whitespace, we expect identifier

	// Check context to determine if this is an identifier or value
	// We need to look back at the previous non-whitespace token type
	// For simplicity in this lexer, we'll use a heuristic:
	// - If we just saw '=', this is a value
	// - Otherwise, it's an identifier

	// Since we don't maintain state, we'll scan as identifier or value based on context
	// Actually, let's check what comes before: scan backwards for '=' on same line
	if l.hasEqualsBeforeOnLine() {
		return l.scanValue()
	}

	// Check if we're inside a section header (after '[')
	if l.hasUnclosedSectionOnLine() {
		return l.scanSectionContent()
	}

	// Default: scan identifier (key or section content)
	return l.scanIdentifier()
}

// skipWhitespace skips spaces and tabs but not newlines.
func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' {
			l.advance()
		} else {
			break
		}
	}
}

// advance moves the position forward by one rune, updating line, column, and offset.
func (l *Lexer) advance() {
	if l.pos >= len(l.source) {
		return
	}

	// Decode the rune to handle UTF-8
	r, size := utf8.DecodeRune(l.source[l.pos:])
	if r == utf8.RuneError && size == 1 {
		// Invalid UTF-8, skip the byte
		l.pos++
		l.col++
		return
	}

	l.pos += size
	l.col++
}

// currentPos returns the current position.
func (l *Lexer) currentPos() Position {
	return Position{Offset: l.pos, Line: l.line, Column: l.col}
}

// scanComment scans a comment (# or ; to end of line).
func (l *Lexer) scanComment() Token {
	start := Position{Offset: l.pos, Line: l.line, Column: l.col}

	// Read to end of line or EOF
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == '\n' || ch == '\r' {
			break
		}
		l.advance()
	}

	value := string(l.source[start.Offset:l.pos])
	return Token{
		Type:  TokenComment,
		Value: value,
		Range: Range{Start: start, End: l.currentPos()},
	}
}

// scanIdentifier scans an identifier (key name).
func (l *Lexer) scanIdentifier() Token {
	start := Position{Offset: l.pos, Line: l.line, Column: l.col}

	// Read until whitespace, '=', newline, or EOF
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ' ' || ch == '\t' || ch == '=' || ch == '\n' || ch == '\r' {
			break
		}
		l.advance()
	}

	value := string(l.source[start.Offset:l.pos])
	return Token{
		Type:  TokenIdentifier,
		Value: value,
		Range: Range{Start: start, End: l.currentPos()},
	}
}

// scanValue scans a value (after '=').
func (l *Lexer) scanValue() Token {
	start := Position{Offset: l.pos, Line: l.line, Column: l.col}

	// Read until newline or EOF
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == '\n' || ch == '\r' {
			break
		}
		l.advance()
	}

	value := string(l.source[start.Offset:l.pos])
	return Token{
		Type:  TokenValue,
		Value: value,
		Range: Range{Start: start, End: l.currentPos()},
	}
}

// scanSectionContent scans the content inside a section header (between [ and ]).
func (l *Lexer) scanSectionContent() Token {
	start := Position{Offset: l.pos, Line: l.line, Column: l.col}

	// Read until ']', newline, or EOF
	for l.pos < len(l.source) {
		ch := l.source[l.pos]
		if ch == ']' || ch == '\n' || ch == '\r' {
			break
		}
		l.advance()
	}

	value := string(l.source[start.Offset:l.pos])
	return Token{
		Type:  TokenIdentifier,
		Value: value,
		Range: Range{Start: start, End: l.currentPos()},
	}
}

// hasEqualsBeforeOnLine checks if there's an '=' earlier on the current line.
func (l *Lexer) hasEqualsBeforeOnLine() bool {
	// Scan backwards from current position to start of line
	pos := l.pos - 1
	for pos >= 0 {
		ch := l.source[pos]
		if ch == '\n' || ch == '\r' {
			// Reached start of line
			return false
		}
		if ch == '=' {
			return true
		}
		pos--
	}
	return false
}

// hasUnclosedSectionOnLine checks if there's an unclosed '[' on the current line.
func (l *Lexer) hasUnclosedSectionOnLine() bool {
	// Scan backwards from current position to start of line
	pos := l.pos - 1
	bracketCount := 0
	for pos >= 0 {
		ch := l.source[pos]
		if ch == '\n' || ch == '\r' {
			// Reached start of line
			break
		}
		if ch == ']' {
			bracketCount++
		}
		if ch == '[' {
			bracketCount--
		}
		pos--
	}
	return bracketCount < 0 // More '[' than ']'
}
