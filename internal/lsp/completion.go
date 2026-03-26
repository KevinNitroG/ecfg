package lsp

import (
	"github.com/KevinNitroG/ecfg/internal/parser"
	"go.lsp.dev/protocol"
)

// CompletionContext holds information about what the user is completing.
type CompletionContext struct {
	CompletingKey   bool   // True if typing property key
	CompletingValue bool   // True if typing property value
	PropertyKey     string // For value completion, which property?
	InPreamble      bool   // True if in document preamble
	InSection       bool   // True if in a section
}

// detectCompletionContext determines what the user is completing based on cursor position.
//
// Strategy:
//  1. Try to resolve position to existing KeyValue node
//  2. If found and on key part → completing key
//  3. If found and on value part → completing value for that property
//  4. If not found → assume completing new property key
func detectCompletionContext(doc *parser.Document, pos protocol.Position) CompletionContext {
	ctx := CompletionContext{}

	// Try to resolve to existing node
	node := FindNodeAtPosition(doc, pos)

	if node == nil {
		// Not on existing node - assume typing new property key
		ctx.CompletingKey = true

		// Determine if in preamble or section
		// Strategy: Check if any section appears before or at this line
		// If no section exists before the cursor, we're in preamble
		parserPos := lspPositionToParser(pos)

		inAnySection := false
		for _, section := range doc.Sections {
			// Check if section starts at or before cursor line
			if section.HeaderRange.Start.Line <= parserPos.Line {
				// We're after at least one section header, so we're in a section
				inAnySection = true
				break
			}
		}

		if inAnySection {
			ctx.InSection = true
		} else {
			// No section before cursor = preamble
			ctx.InPreamble = true
		}

		return ctx
	}

	// On existing KeyValue node
	if node.Part == PartKey {
		ctx.CompletingKey = true
		ctx.PropertyKey = node.KeyValue.Key // User might be editing existing key
	} else if node.Part == PartValue {
		ctx.CompletingValue = true
		ctx.PropertyKey = node.KeyValue.Key // Complete values for this property
	}

	ctx.InPreamble = node.InPreamble
	ctx.InSection = node.InSection

	return ctx
}
