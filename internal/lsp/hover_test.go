package lsp

import (
	"strings"
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/validator"
	"go.lsp.dev/protocol"
)

func TestFormatPropertyHover(t *testing.T) {
	schema := validator.Schema["indent_style"]
	content := formatPropertyHover(schema)

	// Check basic structure
	if !strings.Contains(content, "indent_style") {
		t.Errorf("content missing property name 'indent_style'")
	}
	if !strings.Contains(content, "enum") {
		t.Errorf("content missing type 'enum'")
	}
	if !strings.Contains(content, "tab") {
		t.Errorf("content missing valid value 'tab'")
	}
	if !strings.Contains(content, "space") {
		t.Errorf("content missing valid value 'space'")
	}
	if !strings.Contains(content, schema.Description) {
		t.Errorf("content missing description")
	}
}

func TestComputeHoverOnPropertyKey(t *testing.T) {
	source := "indent_style = tab\n"
	doc, err := parser.Parse([]byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Position on "indent_style" (character 5 = middle of key)
	hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 5})

	if hover == nil {
		t.Fatal("expected hover, got nil")
	}
	if hover.Contents.Kind != protocol.Markdown {
		t.Errorf("Contents.Kind = %v, want Markdown", hover.Contents.Kind)
	}
	if !strings.Contains(hover.Contents.Value, "indent_style") {
		t.Errorf("hover missing property name")
	}
	if !strings.Contains(hover.Contents.Value, "tab") {
		t.Errorf("hover missing valid value 'tab'")
	}
	if hover.Range == nil {
		t.Error("expected Range to be set")
	}
}

func TestComputeHoverOnPropertyValue(t *testing.T) {
	source := "indent_style = tab\n"
	doc, err := parser.Parse([]byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Position on "tab" (character 16 = on value)
	hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 16})

	if hover == nil {
		t.Fatal("expected hover, got nil")
	}
	if !strings.Contains(hover.Contents.Value, "indent_style") {
		t.Errorf("hover on value should show property docs")
	}
}

func TestComputeHoverUnknownProperty(t *testing.T) {
	source := "unknown_prop = value\n"
	doc, err := parser.Parse([]byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 5})
	if hover != nil {
		t.Errorf("expected nil for unknown property, got %+v", hover)
	}
}

func TestComputeHoverOnWhitespace(t *testing.T) {
	source := "indent_style = tab\n"
	doc, err := parser.Parse([]byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	// Position between key and = (whitespace)
	hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 13})
	if hover != nil {
		t.Errorf("expected nil for whitespace, got %+v", hover)
	}
}

func TestComputeHoverAllProperties(t *testing.T) {
	tests := []struct {
		property     string
		value        string
		wantContains []string
	}{
		{
			property:     "root",
			value:        "true",
			wantContains: []string{"root", "boolean", "preamble"},
		},
		{
			property:     "indent_size",
			value:        "4",
			wantContains: []string{"indent_size", "integer", "tab"},
		},
		{
			property:     "end_of_line",
			value:        "lf",
			wantContains: []string{"end_of_line", "lf", "crlf", "cr"},
		},
		{
			property:     "charset",
			value:        "utf-8",
			wantContains: []string{"charset", "utf-8"},
		},
		{
			property:     "trim_trailing_whitespace",
			value:        "true",
			wantContains: []string{"trim_trailing_whitespace", "boolean"},
		},
		{
			property:     "insert_final_newline",
			value:        "false",
			wantContains: []string{"insert_final_newline", "boolean"},
		},
		{
			property:     "tab_width",
			value:        "4",
			wantContains: []string{"tab_width", "integer"},
		},
		{
			property:     "max_line_length",
			value:        "80",
			wantContains: []string{"max_line_length", "integer", "off"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.property, func(t *testing.T) {
			source := tt.property + " = " + tt.value + "\n"
			doc, err := parser.Parse([]byte(source))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 2})

			if hover == nil {
				t.Fatal("expected hover, got nil")
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(hover.Contents.Value, want) {
					t.Errorf("hover missing %q in content:\n%s", want, hover.Contents.Value)
				}
			}
		})
	}
}

func TestComputeHoverRangeConversion(t *testing.T) {
	source := "root = true\n"
	doc, err := parser.Parse([]byte(source))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	hover := ComputeHover(doc, protocol.Position{Line: 0, Character: 2})

	if hover == nil {
		t.Fatal("expected hover, got nil")
	}
	if hover.Range == nil {
		t.Fatal("expected Range to be set")
	}

	// Parser uses 1-indexed lines, LSP uses 0-indexed
	if hover.Range.Start.Line != 0 {
		t.Errorf("Range.Start.Line = %d, want 0", hover.Range.Start.Line)
	}
	if hover.Range.Start.Character != 0 {
		t.Errorf("Range.Start.Character = %d, want 0", hover.Range.Start.Character)
	}
}
