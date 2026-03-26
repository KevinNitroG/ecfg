package lsp

import (
	"fmt"
	"strings"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/validator"
	"go.lsp.dev/protocol"
)

// formatPropertyHover generates Markdown documentation for a property.
func formatPropertyHover(schema validator.PropertySchema) string {
	var buf strings.Builder

	// Header: property name and type
	fmt.Fprintf(&buf, "**%s** _(%s)_\n\n", schema.Name, schema.Type.String())

	// Description
	fmt.Fprintf(&buf, "%s\n", schema.Description)

	// Valid values section for enums
	if len(schema.ValidValues) > 0 {
		buf.WriteString("\n**Valid values:**\n")
		for _, v := range schema.ValidValues {
			fmt.Fprintf(&buf, "- `%s`\n", v)
		}
	}

	// Special values section
	if len(schema.SpecialValues) > 0 {
		buf.WriteString("\n**Special values:**\n")
		for _, v := range schema.SpecialValues {
			fmt.Fprintf(&buf, "- `%s`\n", v)
		}
	}

	// Integer range
	if schema.MinValue != nil || schema.MaxValue != nil {
		buf.WriteString("\n**Range:** ")
		if schema.MinValue != nil {
			fmt.Fprintf(&buf, "min=%d ", *schema.MinValue)
		}
		if schema.MaxValue != nil {
			fmt.Fprintf(&buf, "max=%d", *schema.MaxValue)
		}
		buf.WriteString("\n")
	}

	return buf.String()
}

// ComputeHover computes hover information for a cursor position.
// Returns nil if the position is not on a property or the property is unknown.
func ComputeHover(doc *parser.Document, pos protocol.Position) *protocol.Hover {
	// Step 1: Resolve position to node
	node := FindNodeAtPosition(doc, pos)
	if node == nil {
		return nil // Cursor not on identifiable node
	}

	// Step 2: Ensure we're on a key or value
	if node.Part != PartKey && node.Part != PartValue {
		return nil
	}

	// Step 3: Lookup property in schema (case-insensitive)
	propertyKey := strings.ToLower(node.KeyValue.Key)
	schema, exists := validator.Schema[propertyKey]
	if !exists {
		return nil // Unknown property, no hover
	}

	// Step 4: Format hover content
	content := formatPropertyHover(schema)

	// Step 5: Determine range to highlight
	// Use KeyRange for both key and value hover (highlights the property name)
	highlightRange := node.KeyValue.KeyRange

	// Step 6: Return hover
	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: content,
		},
		Range: &protocol.Range{
			Start: protocol.Position{
				Line:      uint32(highlightRange.Start.Line - 1), // Parser 1-indexed → LSP 0-indexed
				Character: uint32(highlightRange.Start.Column),
			},
			End: protocol.Position{
				Line:      uint32(highlightRange.End.Line - 1),
				Character: uint32(highlightRange.End.Column),
			},
		},
	}
}
