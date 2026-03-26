// Package lsp provides LSP feature implementations for EditorConfig files.
//
// This package implements position-to-node resolution, hover tooltips,
// and completion suggestions using the parser's AST and validator's schema.
package lsp

import (
	"github.com/KevinNitroG/ecfg/internal/parser"
	"go.lsp.dev/protocol"
)

// NodePart indicates which part of a KeyValue node the cursor is on.
type NodePart int

const (
	PartNone  NodePart = iota // Not on a recognizable node part
	PartKey                   // Cursor on property key
	PartValue                 // Cursor on property value
)

// String returns the name of the node part.
func (np NodePart) String() string {
	switch np {
	case PartNone:
		return "None"
	case PartKey:
		return "Key"
	case PartValue:
		return "Value"
	default:
		return "Unknown"
	}
}

// ResolvedNode represents a KeyValue node with context information.
type ResolvedNode struct {
	KeyValue   *parser.KeyValue // The key-value pair at this position
	Part       NodePart         // Which part (key or value) the cursor is on
	InPreamble bool             // True if in the document preamble
	InSection  bool             // True if in a section
}

// lspPositionToParser converts LSP position (0-indexed line) to parser position (1-indexed line).
func lspPositionToParser(lspPos protocol.Position) parser.Position {
	return parser.Position{
		Line:   int(lspPos.Line) + 1,  // LSP 0-indexed → parser 1-indexed
		Column: int(lspPos.Character), // Both 0-indexed
		Offset: 0,                     // Not needed for comparison
	}
}

// containsPosition checks if a position is within a range.
func containsPosition(r parser.Range, pos parser.Position) bool {
	// Position must be:
	// - On or after start position
	// - Before end position
	if pos.Line < r.Start.Line || pos.Line > r.End.Line {
		return false
	}
	if pos.Line == r.Start.Line && pos.Column < r.Start.Column {
		return false
	}
	if pos.Line == r.End.Line && pos.Column >= r.End.Column {
		return false
	}
	return true
}

// FindNodeAtPosition resolves a cursor position to an AST node with context.
// Returns nil if the position is not on a recognizable node (whitespace, comment, etc.).
func FindNodeAtPosition(doc *parser.Document, lspPos protocol.Position) *ResolvedNode {
	pos := lspPositionToParser(lspPos)

	// Check preamble
	if doc.Preamble != nil && containsPosition(doc.Preamble.Range, pos) {
		for _, kv := range doc.Preamble.Pairs {
			// Check key range first (key takes precedence)
			if containsPosition(kv.KeyRange, pos) {
				return &ResolvedNode{
					KeyValue:   kv,
					Part:       PartKey,
					InPreamble: true,
					InSection:  false,
				}
			}
			// Check value range
			if containsPosition(kv.ValueRange, pos) {
				return &ResolvedNode{
					KeyValue:   kv,
					Part:       PartValue,
					InPreamble: true,
					InSection:  false,
				}
			}
		}
		// Position in preamble but not on a key-value node
		return nil
	}

	// Check sections
	for _, section := range doc.Sections {
		if !containsPosition(section.Range, pos) {
			continue
		}

		// Check key-value pairs in this section
		for _, kv := range section.Pairs {
			// Check key range first (key takes precedence)
			if containsPosition(kv.KeyRange, pos) {
				return &ResolvedNode{
					KeyValue:   kv,
					Part:       PartKey,
					InPreamble: false,
					InSection:  true,
				}
			}
			// Check value range
			if containsPosition(kv.ValueRange, pos) {
				return &ResolvedNode{
					KeyValue:   kv,
					Part:       PartValue,
					InPreamble: false,
					InSection:  true,
				}
			}
		}

		// Position in section but not on a key-value node
		return nil
	}

	// Position not in any identifiable node
	return nil
}
