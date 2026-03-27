package parser

import (
	"strings"
)

// Parser consumes tokens from a lexer and builds an AST.
// It handles error recovery to ensure parsing never panics.
type Parser struct {
	lexer   *Lexer
	current Token
	errors  []ParseError
}

// NewParser creates a new parser for the given lexer.
func NewParser(lexer *Lexer) *Parser {
	p := &Parser{
		lexer:  lexer,
		errors: []ParseError{},
	}
	// Prime the parser with the first token
	p.advance()
	return p
}

// Parse is the main entry point for parsing EditorConfig source.
// It never panics and always returns a Document, even if parse errors occurred.
func Parse(source []byte) (*Document, error) {
	lexer := NewLexer(source)
	parser := NewParser(lexer)
	doc := parser.parseDocument()
	return doc, nil
}

// advance moves to the next token.
func (p *Parser) advance() {
	p.current = p.lexer.Scan()
}

// peek returns the current token without consuming it.
func (p *Parser) peek() Token {
	return p.current
}

// expect checks if the current token matches the expected type.
// If not, it adds a parse error and returns false.
func (p *Parser) expect(tokenType TokenType) bool {
	return p.current.Type == tokenType
}

// consume advances if the current token matches the expected type.
// Returns true if consumed, false otherwise.
func (p *Parser) consume(tokenType TokenType) bool {
	if p.expect(tokenType) {
		p.advance()
		return true
	}
	return false
}

// addError records a parse error.
func (p *Parser) addError(r Range, message string, code string) {
	p.errors = append(p.errors, ParseError{
		Range:   r,
		Message: message,
		Code:    code,
	})
}

// skipToSync skips tokens until a synchronization point.
func (p *Parser) skipToSync() {
	for p.current.Type != TokenEOF {
		if p.current.Type == TokenNewline || p.current.Type == TokenSectionStart {
			return
		}
		p.advance()
	}
}

// parseDocument parses the entire document.
func (p *Parser) parseDocument() *Document {
	startPos := Position{Offset: 0, Line: 1, Column: 0}

	doc := &Document{
		Range:    Range{Start: startPos, End: startPos},
		Sections: []*Section{},
		Comments: []*Comment{},
		Errors:   []ParseError{},
	}

	// Parse comments and preamble (before first section)
	preamblePairs := []*KeyValue{}
	preambleComments := []*Comment{}
	topLevelComments := []*Comment{}

	inPreamble := true

	for p.current.Type != TokenEOF {
		switch p.current.Type {
		case TokenEOF, TokenSectionEnd, TokenEquals, TokenValue:
			// Unexpected at top-level - skip and add error
			p.addError(p.current.Range, "unexpected token", "unexpected-token")
			p.advance()
		case TokenComment:
			comment := p.parseComment()
			if inPreamble {
				preambleComments = append(preambleComments, comment)
			} else {
				topLevelComments = append(topLevelComments, comment)
			}

		case TokenSectionStart:
			// First section starts - preamble ends
			if inPreamble {
				inPreamble = false
				// Create preamble if we have pairs or comments
				if len(preamblePairs) > 0 || len(preambleComments) > 0 {
					preambleStart := startPos
					preambleEnd := p.current.Range.Start
					if len(preamblePairs) > 0 {
						preambleEnd = preamblePairs[len(preamblePairs)-1].Range.End
					} else if len(preambleComments) > 0 {
						preambleEnd = preambleComments[len(preambleComments)-1].Range.End
					}
					doc.Preamble = &Preamble{
						Range:    Range{Start: preambleStart, End: preambleEnd},
						Pairs:    preamblePairs,
						Comments: preambleComments,
					}
				}
			}
			// Parse section
			section := p.parseSection()
			doc.Sections = append(doc.Sections, section)

		case TokenIdentifier:
			// Key-value pair
			kv := p.parseKeyValue()
			if kv != nil {
				if inPreamble {
					preamblePairs = append(preamblePairs, kv)
				} else {
					// Orphan key-value after section? Add to last section
					if len(doc.Sections) > 0 {
						lastSec := doc.Sections[len(doc.Sections)-1]
						lastSec.Pairs = append(lastSec.Pairs, kv)
					} else {
						// Treat as preamble
						preamblePairs = append(preamblePairs, kv)
					}
				}
			}

		case TokenNewline:
			// Skip empty lines
			p.advance()

		default:
			// Unexpected token - add error and skip
			p.addError(p.current.Range, "unexpected token", "unexpected-token")
			p.skipToSync()
		}
	}

	// Create preamble if needed and not already created
	if doc.Preamble == nil && (len(preamblePairs) > 0 || len(preambleComments) > 0) {
		preambleStart := startPos
		preambleEnd := p.current.Range.Start
		if len(preamblePairs) > 0 {
			preambleEnd = preamblePairs[len(preamblePairs)-1].Range.End
		} else if len(preambleComments) > 0 {
			preambleEnd = preambleComments[len(preambleComments)-1].Range.End
		}
		doc.Preamble = &Preamble{
			Range:    Range{Start: preambleStart, End: preambleEnd},
			Pairs:    preamblePairs,
			Comments: preambleComments,
		}
	}

	doc.Comments = topLevelComments
	doc.Errors = p.errors
	doc.Range.End = p.current.Range.End

	return doc
}

// parseSection parses a section: [glob] followed by key-value pairs.
func (p *Parser) parseSection() *Section {
	startPos := p.current.Range.Start
	headerStart := p.current.Range.Start

	// Consume [
	p.consume(TokenSectionStart)

	// Read section header content - collect all tokens until newline
	// Then find the last ] to determine the actual header end
	headerTokens := []Token{}
	for p.current.Type != TokenNewline && p.current.Type != TokenEOF {
		headerTokens = append(headerTokens, p.current)
		p.advance()
	}

	// Find the last TokenSectionEnd in the collected tokens
	headerContent := ""
	headerEnd := headerStart
	lastBracketIdx := -1

	for i, tok := range headerTokens {
		if tok.Type == TokenSectionEnd {
			lastBracketIdx = i
		}
	}

	// Build header content from tokens up to last bracket
	if lastBracketIdx >= 0 {
		for i := 0; i < lastBracketIdx; i++ {
			headerContent += headerTokens[i].Value
		}
		headerEnd = headerTokens[lastBracketIdx].Range.End
	} else {
		// No closing bracket found
		for _, tok := range headerTokens {
			headerContent += tok.Value
			headerEnd = tok.Range.End
		}
		p.addError(Range{Start: headerStart, End: headerEnd}, "missing closing bracket in section header", "missing-section-close")
	}

	// Skip newline after header
	p.consume(TokenNewline)

	headerRange := Range{Start: headerStart, End: headerEnd}

	// Parse key-value pairs and comments in this section
	pairs := []*KeyValue{}
	comments := []*Comment{}

	for p.current.Type != TokenEOF && p.current.Type != TokenSectionStart {
		switch p.current.Type {
		case TokenComment:
			comment := p.parseComment()
			comments = append(comments, comment)

		case TokenIdentifier:
			kv := p.parseKeyValue()
			if kv != nil {
				pairs = append(pairs, kv)
			}

		case TokenEOF, TokenSectionStart, TokenSectionEnd, TokenEquals, TokenValue, TokenNewline:
			// End of section or skip
			goto endSection
		}
	}

endSection:
	var endPos Position
	if len(pairs) > 0 {
		endPos = pairs[len(pairs)-1].Range.End
	} else if len(comments) > 0 {
		endPos = comments[len(comments)-1].Range.End
	} else {
		endPos = headerEnd
	}

	return &Section{
		Range:       Range{Start: startPos, End: endPos},
		Header:      strings.TrimSpace(headerContent),
		HeaderRange: headerRange,
		Pairs:       pairs,
		Comments:    comments,
	}
}

// parseKeyValue parses a key-value pair: key = value
func (p *Parser) parseKeyValue() *KeyValue {
	if p.current.Type != TokenIdentifier {
		return nil
	}

	startPos := p.current.Range.Start
	keyStart := p.current.Range.Start
	key := strings.TrimSpace(p.current.Value)
	keyEnd := p.current.Range.End
	p.advance()

	// Expect =
	if p.current.Type != TokenEquals {
		p.addError(Range{Start: keyStart, End: keyEnd}, "missing '=' after key", "missing-equals")
		// Skip to next line
		p.skipToSync()
		if p.current.Type == TokenNewline {
			p.advance()
		}
		return nil
	}

	p.consume(TokenEquals)

	// Read value (may be empty)
	value := ""
	valueStart := p.current.Range.Start
	valueEnd := p.current.Range.End

	switch p.current.Type {
	case TokenValue:
		value = strings.TrimSpace(p.current.Value)
		valueEnd = p.current.Range.End
		p.advance()
	case TokenIdentifier:
		// Sometimes the lexer might emit identifier instead of value
		value = strings.TrimSpace(p.current.Value)
		valueEnd = p.current.Range.End
		p.advance()
	case TokenEOF, TokenComment, TokenSectionStart, TokenSectionEnd, TokenEquals, TokenNewline:
		// Invalid tokens after equals - empty value is valid
	}
	// else: empty value is valid

	// Skip newline
	p.consume(TokenNewline)

	endPos := valueEnd
	if endPos.Offset < keyEnd.Offset {
		endPos = keyEnd
	}

	return &KeyValue{
		Range:      Range{Start: startPos, End: endPos},
		Key:        key,
		KeyRange:   Range{Start: keyStart, End: keyEnd},
		Value:      value,
		ValueRange: Range{Start: valueStart, End: valueEnd},
	}
}

// parseComment parses a comment line.
func (p *Parser) parseComment() *Comment {
	comment := &Comment{
		Range: p.current.Range,
		Text:  p.current.Value,
	}
	p.advance()
	// Skip newline after comment
	p.consume(TokenNewline)
	return comment
}
