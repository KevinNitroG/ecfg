package lsp

import (
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.lsp.dev/protocol"
)

func TestDetectCompletionContext(t *testing.T) {
	tests := []struct {
		name                string
		source              string
		position            protocol.Position
		wantCompletingKey   bool
		wantCompletingValue bool
		wantPropertyKey     string
		wantInPreamble      bool
		wantInSection       bool
	}{
		{
			name:              "incomplete property key in preamble",
			source:            "ind",
			position:          protocol.Position{Line: 0, Character: 2},
			wantCompletingKey: true,
			wantInPreamble:    true,
		},
		{
			name:                "value after equals",
			source:              "indent_style = t",
			position:            protocol.Position{Line: 0, Character: 15}, // On the 't'
			wantCompletingValue: true,
			wantPropertyKey:     "indent_style",
			wantInPreamble:      true, // No section, so in preamble
		},
		{
			name:              "new property in section",
			source:            "[*.go]\ntab",
			position:          protocol.Position{Line: 1, Character: 2},
			wantCompletingKey: true,
			wantInSection:     true,
		},
		{
			name:              "line without equals - completing new property",
			source:            "indent",
			position:          protocol.Position{Line: 0, Character: 5},
			wantCompletingKey: true,
			wantInPreamble:    true, // No section, so in preamble
		},
		{
			name:              "middle of key",
			source:            "indent_style = tab",
			position:          protocol.Position{Line: 0, Character: 6},
			wantCompletingKey: true,
			wantPropertyKey:   "indent_style",
			wantInPreamble:    true, // No section, so in preamble
		},
		{
			name:                "in value area",
			source:              "indent_style = tab",
			position:            protocol.Position{Line: 0, Character: 17},
			wantCompletingValue: true,
			wantPropertyKey:     "indent_style",
			wantInPreamble:      true, // No section, so in preamble
		},
		{
			name:                "value with spaces",
			source:              "charset = utf-8",
			position:            protocol.Position{Line: 0, Character: 13},
			wantCompletingValue: true,
			wantPropertyKey:     "charset",
			wantInPreamble:      true, // No section, so in preamble
		},
		{
			name:              "empty line in section",
			source:            "[*.go]\n",
			position:          protocol.Position{Line: 1, Character: 0},
			wantCompletingKey: true,
			wantInSection:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parser.Parse([]byte(tt.source))
			assert.NoError(t, err)
			ctx := detectCompletionContext(doc, tt.position)

			assert.Equal(t, tt.wantCompletingKey, ctx.CompletingKey, "CompletingKey mismatch")
			assert.Equal(t, tt.wantCompletingValue, ctx.CompletingValue, "CompletingValue mismatch")
			assert.Equal(t, tt.wantPropertyKey, ctx.PropertyKey, "PropertyKey mismatch")
			assert.Equal(t, tt.wantInPreamble, ctx.InPreamble, "InPreamble mismatch")
			assert.Equal(t, tt.wantInSection, ctx.InSection, "InSection mismatch")
		})
	}
}

func TestCompletePropertyKeysInPreamble(t *testing.T) {
	source := "ro"
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 2})

	require.NotNil(t, completion)
	assert.False(t, completion.IsIncomplete)

	// Should include all 9 properties including root
	assert.Len(t, completion.Items, 9)

	// Find root property
	var rootItem *protocol.CompletionItem
	for i := range completion.Items {
		if completion.Items[i].Label == "root" {
			rootItem = &completion.Items[i]
			break
		}
	}

	require.NotNil(t, rootItem, "root property should be in preamble completions")
	assert.Equal(t, protocol.CompletionItemKindProperty, rootItem.Kind)
	assert.NotNil(t, rootItem.Documentation)
}

func TestCompletePropertyKeysInSection(t *testing.T) {
	source := "[*.go]\nind"
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 1, Character: 2})

	require.NotNil(t, completion)

	// Should include 8 properties (excluding root)
	assert.Len(t, completion.Items, 8)

	// Verify root is NOT present
	for _, item := range completion.Items {
		assert.NotEqual(t, "root", item.Label, "root should not appear in section completions")
	}
}

func TestCompletePropertyValues_EnumProperty(t *testing.T) {
	source := "indent_style = "
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 15})

	require.NotNil(t, completion)
	assert.Len(t, completion.Items, 2) // tab, space

	labels := []string{}
	for _, item := range completion.Items {
		labels = append(labels, item.Label)
		assert.Equal(t, protocol.CompletionItemKindValue, item.Kind)
	}

	assert.ElementsMatch(t, []string{"tab", "space"}, labels)
}

func TestCompletePropertyValues_SpecialValues(t *testing.T) {
	source := "indent_size = "
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 14})

	require.NotNil(t, completion)

	// Should only include special value "tab" (integers are typed by user)
	assert.Len(t, completion.Items, 1)
	assert.Equal(t, "tab", completion.Items[0].Label)
}

func TestCompletePropertyValues_IntegerProperty(t *testing.T) {
	source := "tab_width = "
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 12})

	require.NotNil(t, completion)

	// Integer property with no special values - no suggestions
	assert.Len(t, completion.Items, 0)
}

func TestCompletePropertyValues_BooleanProperty(t *testing.T) {
	source := "insert_final_newline = "
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 23})

	require.NotNil(t, completion)
	assert.Len(t, completion.Items, 2) // true, false

	labels := []string{}
	for _, item := range completion.Items {
		labels = append(labels, item.Label)
	}

	assert.ElementsMatch(t, []string{"true", "false"}, labels)
}

func TestCompletePropertyValues_UnknownProperty(t *testing.T) {
	source := "unknown_property = "
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 19})

	require.NotNil(t, completion)
	assert.Len(t, completion.Items, 0)
}

func TestCompletionItemsHaveDocumentation(t *testing.T) {
	source := "ind"
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 2})

	require.NotNil(t, completion)
	require.Greater(t, len(completion.Items), 0)

	for _, item := range completion.Items {
		assert.NotNil(t, item.Documentation, "item %s should have documentation", item.Label)
		if mc, ok := item.Documentation.(*protocol.MarkupContent); ok {
			assert.NotEmpty(t, mc.Value, "item %s documentation should not be empty", item.Label)
		}
	}
}

func TestCompletionItemInsertText(t *testing.T) {
	source := "ind"
	doc, err := parser.Parse([]byte(source))
	require.NoError(t, err)

	completion := ComputeCompletion(doc, protocol.Position{Line: 0, Character: 2})

	// Find indent_style item
	var indentStyleItem *protocol.CompletionItem
	for i := range completion.Items {
		if completion.Items[i].Label == "indent_style" {
			indentStyleItem = &completion.Items[i]
			break
		}
	}

	require.NotNil(t, indentStyleItem)
	assert.Equal(t, "indent_style = ", indentStyleItem.InsertText)
}
