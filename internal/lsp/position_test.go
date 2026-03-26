package lsp

import (
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"go.lsp.dev/protocol"
)

func TestFindNodeAtPosition(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		position       protocol.Position
		wantPart       NodePart
		wantKey        string
		wantValue      string
		wantInPreamble bool
		wantInSection  bool
		wantNil        bool
	}{
		{
			name:           "key in preamble",
			source:         "root = true\n",
			position:       protocol.Position{Line: 0, Character: 2},
			wantPart:       PartKey,
			wantKey:        "root",
			wantValue:      "true",
			wantInPreamble: true,
		},
		{
			name:           "value in preamble",
			source:         "root = true\n",
			position:       protocol.Position{Line: 0, Character: 9},
			wantPart:       PartValue,
			wantKey:        "root",
			wantValue:      "true",
			wantInPreamble: true,
		},
		{
			name:          "key in section",
			source:        "[*.go]\nindent_style = tab\n",
			position:      protocol.Position{Line: 1, Character: 5},
			wantPart:      PartKey,
			wantKey:       "indent_style",
			wantValue:     "tab",
			wantInSection: true,
		},
		{
			name:          "value in section",
			source:        "[*.go]\nindent_style = tab\n",
			position:      protocol.Position{Line: 1, Character: 17},
			wantPart:      PartValue,
			wantKey:       "indent_style",
			wantValue:     "tab",
			wantInSection: true,
		},
		{
			name:     "whitespace between key and equals",
			source:   "indent_style = tab\n",
			position: protocol.Position{Line: 0, Character: 13},
			wantNil:  true,
		},
		{
			name:     "on comment",
			source:   "# This is a comment\n",
			position: protocol.Position{Line: 0, Character: 5},
			wantNil:  true,
		},
		{
			name:     "on section header",
			source:   "[*.go]\n",
			position: protocol.Position{Line: 0, Character: 2},
			wantNil:  true,
		},
		{
			name:           "LSP line 0 maps to parser line 1",
			source:         "root = true\n",
			position:       protocol.Position{Line: 0, Character: 0},
			wantPart:       PartKey,
			wantKey:        "root",
			wantInPreamble: true,
		},
		{
			name:          "multi-line document cursor on line 3",
			source:        "root = true\n\n[*.go]\nindent_size = 4\n",
			position:      protocol.Position{Line: 3, Character: 5},
			wantPart:      PartKey,
			wantKey:       "indent_size",
			wantValue:     "4",
			wantInSection: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := parser.Parse([]byte(tt.source))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			node := FindNodeAtPosition(doc, tt.position)

			if tt.wantNil {
				if node != nil {
					t.Errorf("expected nil, got %+v", node)
				}
				return
			}

			if node == nil {
				t.Fatalf("expected node, got nil")
			}

			if node.Part != tt.wantPart {
				t.Errorf("Part = %v, want %v", node.Part, tt.wantPart)
			}

			if node.KeyValue.Key != tt.wantKey {
				t.Errorf("Key = %q, want %q", node.KeyValue.Key, tt.wantKey)
			}

			if tt.wantValue != "" && node.KeyValue.Value != tt.wantValue {
				t.Errorf("Value = %q, want %q", node.KeyValue.Value, tt.wantValue)
			}

			if node.InPreamble != tt.wantInPreamble {
				t.Errorf("InPreamble = %v, want %v", node.InPreamble, tt.wantInPreamble)
			}

			if node.InSection != tt.wantInSection {
				t.Errorf("InSection = %v, want %v", node.InSection, tt.wantInSection)
			}
		})
	}
}
