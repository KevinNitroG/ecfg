package lsp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/validator"
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
//  4. If not found → check if there's a KeyValue on the same line with cursor after the key
//     (this handles "property = |" where cursor is in empty value position)
//  5. Otherwise → assume completing new property key
func detectCompletionContext(doc *parser.Document, pos protocol.Position) CompletionContext {
	ctx := CompletionContext{}
	parserPos := lspPositionToParser(pos)

	// Try to resolve to existing node
	node := FindNodeAtPosition(doc, pos)

	if node == nil {
		// Not on existing node - but check if we're after a KeyValue's equals sign
		// This handles completion for empty values like "indent_style = |"

		// Check preamble first
		if doc.Preamble != nil {
			for _, kv := range doc.Preamble.Pairs {
				// If cursor is on the same line as the KeyValue and after the key
				if kv.Range.Start.Line == parserPos.Line && parserPos.Column > kv.KeyRange.End.Column {
					// We're completing the value for this property
					ctx.CompletingValue = true
					ctx.PropertyKey = kv.Key
					ctx.InPreamble = true
					ctx.InSection = false
					return ctx
				}
			}
		}

		// Check sections
		for _, section := range doc.Sections {
			for _, kv := range section.Pairs {
				// If cursor is on the same line as the KeyValue and after the key
				if kv.Range.Start.Line == parserPos.Line && parserPos.Column > kv.KeyRange.End.Column {
					// We're completing the value for this property
					ctx.CompletingValue = true
					ctx.PropertyKey = kv.Key
					ctx.InPreamble = false
					ctx.InSection = true
					return ctx
				}
			}
		}

		// Not on or after existing KeyValue - assume typing new property key
		ctx.CompletingKey = true

		// Determine if in preamble or section
		// Strategy: Check if any section appears before or at this line
		// If no section exists before the cursor, we're in preamble
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
	switch node.Part {
	case PartKey:
		ctx.CompletingKey = true
		ctx.PropertyKey = node.KeyValue.Key // User might be editing existing key
	case PartValue:
		ctx.CompletingValue = true
		ctx.PropertyKey = node.KeyValue.Key // Complete values for this property
	case PartNone:
		// Position resolved to KeyValue but not specifically on key or value
		// This shouldn't happen in practice, but handle it gracefully
		ctx.CompletingKey = true
	}

	ctx.InPreamble = node.InPreamble
	ctx.InSection = node.InSection

	return ctx
}

// completePropertyKeys returns completion items for property keys.
// Filters out preamble-only properties (like "root") when not in preamble.
func completePropertyKeys(inPreamble bool) []protocol.CompletionItem {
	items := []protocol.CompletionItem{}

	for name, schema := range validator.Schema {
		// Filter out root if not in preamble
		if schema.PreambleOnly && !inPreamble {
			continue
		}

		detail := schema.Type.String()

		// Add insert text with " = " suffix for convenience
		insertText := name + " = "

		item := protocol.CompletionItem{
			Label:  name,
			Kind:   protocol.CompletionItemKindProperty,
			Detail: detail,
			Documentation: &protocol.MarkupContent{
				Kind:  protocol.Markdown,
				Value: schema.Description,
			},
			InsertText:       insertText,
			InsertTextFormat: protocol.InsertTextFormatPlainText,
		}

		items = append(items, item)
	}

	// Sort alphabetically for consistent ordering
	sort.Slice(items, func(i, j int) bool {
		return items[i].Label < items[j].Label
	})

	return items
}

// completePropertyValues returns completion items for property values.
// Returns enum values, special values, or empty list based on property type.
func completePropertyValues(propertyKey string) []protocol.CompletionItem {
	// Case-insensitive lookup
	propertyKey = strings.ToLower(strings.TrimSpace(propertyKey))

	schema, exists := validator.Schema[propertyKey]
	if !exists {
		return []protocol.CompletionItem{} // Unknown property
	}

	items := []protocol.CompletionItem{}

	// For integer properties, don't suggest completions (user types number)
	// But do suggest special values if any exist
	if schema.Type == validator.PropertyTypeInteger {
		if len(schema.SpecialValues) == 0 {
			return []protocol.CompletionItem{}
		}
	}

	// Add enum values
	for _, value := range schema.ValidValues {
		detail := fmt.Sprintf("Valid value for %s", propertyKey)
		item := protocol.CompletionItem{
			Label:      value,
			Kind:       protocol.CompletionItemKindValue,
			Detail:     detail,
			InsertText: value,
		}
		items = append(items, item)
	}

	// Add special values (e.g., "tab" for indent_size, "off" for max_line_length)
	for _, value := range schema.SpecialValues {
		detail := fmt.Sprintf("Special value for %s", propertyKey)
		item := protocol.CompletionItem{
			Label:      value,
			Kind:       protocol.CompletionItemKindValue,
			Detail:     detail,
			InsertText: value,
		}
		items = append(items, item)
	}

	return items
}

// ComputeCompletion computes completion items for a given position.
// Returns context-aware property keys or property values based on cursor position.
func ComputeCompletion(doc *parser.Document, pos protocol.Position) *protocol.CompletionList {
	ctx := detectCompletionContext(doc, pos)

	var items []protocol.CompletionItem

	if ctx.CompletingKey {
		items = completePropertyKeys(ctx.InPreamble)
	} else if ctx.CompletingValue {
		items = completePropertyValues(ctx.PropertyKey)
	} else {
		// Unknown context, return empty
		items = []protocol.CompletionItem{}
	}

	return &protocol.CompletionList{
		IsIncomplete: false, // EditorConfig is small, always return all
		Items:        items,
	}
}
