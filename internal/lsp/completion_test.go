package lsp

import (
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/stretchr/testify/assert"
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
