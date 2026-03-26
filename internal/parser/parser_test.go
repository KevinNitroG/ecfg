package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestParser validates parser behavior with table-driven tests covering:
// - Valid syntax (preamble, sections, key-value pairs, comments)
// - Malformed input (error recovery without panicking)
// - Position accuracy (Range tracking for LSP)
func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, doc *Document)
		wantErrs int
	}{
		{
			name:  "preamble only",
			input: "root = true\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil {
					t.Fatal("Expected preamble, got nil")
				}
				if len(doc.Preamble.Pairs) != 1 {
					t.Fatalf("Expected 1 preamble pair, got %d", len(doc.Preamble.Pairs))
				}
				kv := doc.Preamble.Pairs[0]
				if kv.Key != "root" {
					t.Errorf("Expected key 'root', got %q", kv.Key)
				}
				if kv.Value != "true" {
					t.Errorf("Expected value 'true', got %q", kv.Value)
				}
				// Verify range exists
				if kv.Range.Start.Line != 1 {
					t.Errorf("Expected KeyValue on line 1, got line %d", kv.Range.Start.Line)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "section with properties",
			input: "[*.go]\nindent_style = tab\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
				}
				sec := doc.Sections[0]
				if sec.Header != "*.go" {
					t.Errorf("Expected header '*.go', got %q", sec.Header)
				}
				if len(sec.Pairs) != 1 {
					t.Fatalf("Expected 1 pair in section, got %d", len(sec.Pairs))
				}
				kv := sec.Pairs[0]
				if kv.Key != "indent_style" {
					t.Errorf("Expected key 'indent_style', got %q", kv.Key)
				}
				if kv.Value != "tab" {
					t.Errorf("Expected value 'tab', got %q", kv.Value)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "preamble and section",
			input: "root = true\n\n[*.go]\nindent_style = tab\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil {
					t.Fatal("Expected preamble")
				}
				if len(doc.Preamble.Pairs) != 1 {
					t.Errorf("Expected 1 preamble pair, got %d", len(doc.Preamble.Pairs))
				}
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
				}
			},
			wantErrs: 0,
		},
		{
			name:  "multiple sections",
			input: "[*.go]\nindent_style = tab\n\n[*.js]\nindent_style = space\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 2 {
					t.Fatalf("Expected 2 sections, got %d", len(doc.Sections))
				}
				if doc.Sections[0].Header != "*.go" {
					t.Errorf("Expected first section '*.go', got %q", doc.Sections[0].Header)
				}
				if doc.Sections[1].Header != "*.js" {
					t.Errorf("Expected second section '*.js', got %q", doc.Sections[1].Header)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "comments preserved",
			input: "# This is a comment\nroot = true\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Preamble.Comments) == 0 && len(doc.Comments) == 0 {
					t.Error("Expected comment to be preserved")
				}
			},
			wantErrs: 0,
		},
		{
			name:  "section with comment",
			input: "[*.go]\n# Indent with tabs\nindent_style = tab\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
				}
				sec := doc.Sections[0]
				if len(sec.Comments) == 0 {
					t.Error("Expected comment in section")
				}
			},
			wantErrs: 0,
		},
		{
			name:  "malformed unclosed section",
			input: "[*.go\nindent_style = tab\n",
			validate: func(t *testing.T, doc *Document) {
				// Parser should recover and continue
				if len(doc.Sections) == 0 {
					t.Error("Expected parser to create section despite missing ]")
				}
			},
			wantErrs: 1, // Missing ] error
		},
		{
			name:  "malformed missing equals",
			input: "indent_style tab\n",
			validate: func(t *testing.T, doc *Document) {
				// Should collect error but continue
			},
			wantErrs: 1,
		},
		{
			name:  "malformed no key",
			input: "= value\n",
			validate: func(t *testing.T, doc *Document) {
				// Should collect error
			},
			wantErrs: 1,
		},
		{
			name:  "empty value is valid",
			input: "key =\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil || len(doc.Preamble.Pairs) != 1 {
					t.Fatal("Expected one preamble pair")
				}
				kv := doc.Preamble.Pairs[0]
				if kv.Key != "key" {
					t.Errorf("Expected key 'key', got %q", kv.Key)
				}
				if kv.Value != "" {
					t.Errorf("Expected empty value, got %q", kv.Value)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "whitespace trimming",
			input: "  key  =  value  \n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil || len(doc.Preamble.Pairs) != 1 {
					t.Fatal("Expected one preamble pair")
				}
				kv := doc.Preamble.Pairs[0]
				if kv.Key != "key" {
					t.Errorf("Expected trimmed key 'key', got %q", kv.Key)
				}
				if kv.Value != "value" {
					t.Errorf("Expected trimmed value 'value', got %q", kv.Value)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "value with internal whitespace",
			input: "key = value with spaces\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil || len(doc.Preamble.Pairs) != 1 {
					t.Fatal("Expected one preamble pair")
				}
				kv := doc.Preamble.Pairs[0]
				if kv.Value != "value with spaces" {
					t.Errorf("Expected value with internal spaces, got %q", kv.Value)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "complex glob pattern",
			input: "[*.{js,jsx,ts,tsx}]\nindent_size = 2\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
				}
				if doc.Sections[0].Header != "*.{js,jsx,ts,tsx}" {
					t.Errorf("Expected complex glob, got %q", doc.Sections[0].Header)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "empty lines ignored",
			input: "root = true\n\n\n[*.go]\nindent_style = tab\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil {
					t.Error("Expected preamble")
				}
				if len(doc.Sections) != 1 {
					t.Errorf("Expected 1 section, got %d", len(doc.Sections))
				}
			},
			wantErrs: 0,
		},
		{
			name:  "both comment styles",
			input: "# Hash comment\n; Semicolon comment\nroot = true\n",
			validate: func(t *testing.T, doc *Document) {
				commentCount := len(doc.Comments)
				if doc.Preamble != nil {
					commentCount += len(doc.Preamble.Comments)
				}
				if commentCount != 2 {
					t.Errorf("Expected 2 comments, got %d", commentCount)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "section with multiple properties",
			input: "[*.go]\nindent_style = tab\nindent_size = 4\ntab_width = 4\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section, got %d", len(doc.Sections))
				}
				sec := doc.Sections[0]
				if len(sec.Pairs) != 3 {
					t.Errorf("Expected 3 pairs, got %d", len(sec.Pairs))
				}
			},
			wantErrs: 0,
		},
		{
			name:  "range tracking for keys",
			input: "root = true\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil || len(doc.Preamble.Pairs) != 1 {
					t.Fatal("Expected one preamble pair")
				}
				kv := doc.Preamble.Pairs[0]
				// KeyRange should point to "root"
				if kv.KeyRange.Start.Column >= kv.KeyRange.End.Column {
					t.Error("Expected valid KeyRange")
				}
				// ValueRange should point to "true"
				if kv.ValueRange.Start.Column >= kv.ValueRange.End.Column {
					t.Error("Expected valid ValueRange")
				}
			},
			wantErrs: 0,
		},
		{
			name:  "section header range",
			input: "[*.go]\n",
			validate: func(t *testing.T, doc *Document) {
				if len(doc.Sections) != 1 {
					t.Fatalf("Expected 1 section")
				}
				sec := doc.Sections[0]
				// HeaderRange should point to [*.go]
				if sec.HeaderRange.Start.Line != 1 {
					t.Errorf("Expected header on line 1, got line %d", sec.HeaderRange.Start.Line)
				}
			},
			wantErrs: 0,
		},
		{
			name:  "node type identification - preamble vs section",
			input: "root = true\n[section]\nother = value\n",
			validate: func(t *testing.T, doc *Document) {
				// root=true should be in preamble
				if doc.Preamble == nil || len(doc.Preamble.Pairs) != 1 {
					t.Fatal("Expected preamble with 1 pair")
				}
				if doc.Preamble.Pairs[0].Key != "root" {
					t.Error("Expected root in preamble")
				}
				// other=value should be in section
				if len(doc.Sections) != 1 || len(doc.Sections[0].Pairs) != 1 {
					t.Fatal("Expected 1 section with 1 pair")
				}
				if doc.Sections[0].Pairs[0].Key != "other" {
					t.Error("Expected other in section")
				}
			},
			wantErrs: 0,
		},
		{
			name:  "unclosed section at EOF",
			input: "[*.go",
			validate: func(t *testing.T, doc *Document) {
				// Should have section with error
				if len(doc.Sections) == 0 {
					t.Error("Expected section despite unclosed bracket")
				}
			},
			wantErrs: 1,
		},
		{
			name:  "CRLF line endings",
			input: "root = true\r\n[*.go]\r\nindent_style = tab\r\n",
			validate: func(t *testing.T, doc *Document) {
				if doc.Preamble == nil {
					t.Error("Expected preamble")
				}
				if len(doc.Sections) != 1 {
					t.Errorf("Expected 1 section, got %d", len(doc.Sections))
				}
			},
			wantErrs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := Parse([]byte(tt.input))
			if err != nil {
				t.Fatalf("Parse returned error: %v", err)
			}
			if doc == nil {
				t.Fatal("Parse returned nil document")
			}

			// Verify error count
			if len(doc.Errors) != tt.wantErrs {
				t.Errorf("Expected %d errors, got %d", tt.wantErrs, len(doc.Errors))
				for i, e := range doc.Errors {
					t.Logf("  Error %d: %s (code: %s) at %s", i+1, e.Message, e.Code, e.Range)
				}
			}

			// Run custom validation
			if tt.validate != nil {
				tt.validate(t, doc)
			}

			// Verify document has valid range
			if doc.Range.Start.Line < 1 {
				t.Error("Document range has invalid start line")
			}
		})
	}
}

// TestParserWithFixtures loads all test fixtures and verifies parser doesn't panic.
func TestParserWithFixtures(t *testing.T) {
	fixturePatterns := []string{
		"testdata/valid/*.editorconfig",
		"testdata/malformed/*.editorconfig",
		"testdata/positions/*.editorconfig",
	}

	for _, pattern := range fixturePatterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatalf("Failed to glob %s: %v", pattern, err)
		}

		for _, path := range matches {
			t.Run(filepath.Base(path), func(t *testing.T) {
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read %s: %v", path, err)
				}

				doc, err := Parse(source)
				if err != nil {
					t.Fatalf("Parse returned error: %v", err)
				}
				if doc == nil {
					t.Fatal("Parse returned nil document")
				}

				// Verify no panic occurred (test would fail if panic)
				t.Logf("Parsed %s: %d sections, %d errors", filepath.Base(path), len(doc.Sections), len(doc.Errors))

				// For valid fixtures, expect no errors
				if strings.Contains(path, "/valid/") && len(doc.Errors) > 0 {
					t.Errorf("Valid fixture %s produced errors:", filepath.Base(path))
					for _, e := range doc.Errors {
						t.Logf("  %s", e.Message)
					}
				}

				// For malformed fixtures, expect errors
				if strings.Contains(path, "/malformed/") && len(doc.Errors) == 0 {
					t.Errorf("Malformed fixture %s produced no errors", filepath.Base(path))
				}
			})
		}
	}
}

// TestParserPositionAccuracy validates Range tracking for all node types.
func TestParserPositionAccuracy(t *testing.T) {
	input := "root = true\n[*.go]\nindent_style = tab\n"
	doc, err := Parse([]byte(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify document range covers entire input
	if doc.Range.Start.Offset != 0 {
		t.Errorf("Document start offset should be 0, got %d", doc.Range.Start.Offset)
	}
	if doc.Range.End.Offset != len(input) {
		t.Errorf("Document end offset should be %d, got %d", len(input), doc.Range.End.Offset)
	}

	// Verify preamble range
	if doc.Preamble != nil {
		if doc.Preamble.Range.Start.Line != 1 {
			t.Errorf("Preamble should start at line 1, got %d", doc.Preamble.Range.Start.Line)
		}
	}

	// Verify section range
	if len(doc.Sections) > 0 {
		sec := doc.Sections[0]
		if sec.Range.Start.Line != 2 {
			t.Errorf("Section should start at line 2, got %d", sec.Range.Start.Line)
		}
		if sec.HeaderRange.Start.Line != 2 {
			t.Errorf("Section header should be on line 2, got %d", sec.HeaderRange.Start.Line)
		}
	}
}
