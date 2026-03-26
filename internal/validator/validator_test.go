package validator

import (
	"github.com/KevinNitroG/ecfg/internal/parser"
	"testing"
)

// TestValidateInvalidEnumValue tests that invalid enum values are detected.
func TestValidateInvalidEnumValue(t *testing.T) {
	source := `indent_style = invalid`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Property != "indent_style" {
		t.Errorf("Expected property 'indent_style', got %q", err.Property)
	}
	if err.Value != "invalid" {
		t.Errorf("Expected value 'invalid', got %q", err.Value)
	}
	if err.Reason == "" {
		t.Error("Expected reason to be non-empty")
	}
}

// TestValidateValidEnumValue tests that valid enum values pass.
func TestValidateValidEnumValue(t *testing.T) {
	source := `indent_style = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors, got %d: %v", len(errors), errors)
	}
}

// TestValidateRootInPreambleValid tests that root=true in preamble passes.
func TestValidateRootInPreambleValid(t *testing.T) {
	source := `root = true
[*.go]
indent_style = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors, got %d: %v", len(errors), errors)
	}
}

// TestValidateRootInSection tests that root in section is detected.
func TestValidateRootInSection(t *testing.T) {
	source := `[*.go]
root = true
indent_style = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for root in section, got %d", len(errors))
	}

	err := errors[0]
	if err.Property != "root" {
		t.Errorf("Expected property 'root', got %q", err.Property)
	}
	if err.Reason == "" {
		t.Error("Expected reason to describe preamble-only constraint")
	}
}

// TestValidateIntegerOutOfRange tests that out-of-range integers are detected.
func TestValidateIntegerOutOfRange(t *testing.T) {
	source := `indent_size = 100`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Property != "indent_size" {
		t.Errorf("Expected property 'indent_size', got %q", err.Property)
	}
}

// TestValidateIntegerInRange tests that in-range integers pass.
func TestValidateIntegerInRange(t *testing.T) {
	source := `indent_size = 4`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors, got %d: %v", len(errors), errors)
	}
}

// TestValidateSpecialValueAccepted tests that special values like "tab" are accepted.
func TestValidateSpecialValueAccepted(t *testing.T) {
	source := `indent_size = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors for special value 'tab', got %d: %v", len(errors), errors)
	}
}

// TestValidateBooleanInvalid tests that invalid boolean values are detected.
func TestValidateBooleanInvalid(t *testing.T) {
	source := `insert_final_newline = yes`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Property != "insert_final_newline" {
		t.Errorf("Expected property 'insert_final_newline', got %q", err.Property)
	}
}

// TestValidateBooleanValid tests that valid boolean values pass.
func TestValidateBooleanValid(t *testing.T) {
	tests := []string{
		`insert_final_newline = true`,
		`insert_final_newline = false`,
	}

	for _, source := range tests {
		doc, _ := parser.Parse([]byte(source))
		errors := Validate(doc)
		if len(errors) != 0 {
			t.Fatalf("Expected 0 errors for %q, got %d: %v", source, len(errors), errors)
		}
	}
}

// TestValidateUnknownProperty tests that unknown properties are detected.
func TestValidateUnknownProperty(t *testing.T) {
	source := `unknown_prop = value`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for unknown property, got %d", len(errors))
	}

	err := errors[0]
	if err.Property != "unknown_prop" {
		t.Errorf("Expected property 'unknown_prop', got %q", err.Property)
	}
}

// TestValidateEmptyDocument tests that empty documents pass.
func TestValidateEmptyDocument(t *testing.T) {
	source := ""
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors for empty document, got %d: %v", len(errors), errors)
	}
}

// TestValidateErrorIncludesRange tests that ValidationError includes Range from AST.
func TestValidateErrorIncludesRange(t *testing.T) {
	source := `indent_style = invalid`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(errors))
	}

	err := errors[0]
	if err.Range.Start.Line == 0 && err.Range.End.Line == 0 {
		t.Error("Expected Range to be set from AST")
	}
}

// TestValidateMultipleErrors tests that multiple errors are collected.
func TestValidateMultipleErrors(t *testing.T) {
	source := `indent_style = invalid
[*.go]
root = true
indent_size = 100`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	// Should have at least 3 errors: indent_style invalid, root in section, indent_size out of range
	if len(errors) < 3 {
		t.Fatalf("Expected at least 3 errors, got %d", len(errors))
	}
}

// TestValidateCharsetEnum tests charset enum validation.
func TestValidateCharsetEnum(t *testing.T) {
	validCharsets := []string{"utf-8", "utf-8-bom", "utf-16be", "utf-16le", "latin1"}
	for _, cs := range validCharsets {
		source := `charset = ` + cs
		doc, _ := parser.Parse([]byte(source))
		errors := Validate(doc)
		if len(errors) != 0 {
			t.Errorf("charset = %q should be valid, got error: %v", cs, errors)
		}
	}

	// Test invalid charset
	source := `charset = invalid`
	doc, _ := parser.Parse([]byte(source))
	errors := Validate(doc)
	if len(errors) != 1 {
		t.Errorf("charset = invalid should produce 1 error, got %d", len(errors))
	}
}

// TestValidateEndOfLineEnum tests end_of_line enum validation.
func TestValidateEndOfLineEnum(t *testing.T) {
	validEol := []string{"lf", "crlf", "cr"}
	for _, eol := range validEol {
		source := `end_of_line = ` + eol
		doc, _ := parser.Parse([]byte(source))
		errors := Validate(doc)
		if len(errors) != 0 {
			t.Errorf("end_of_line = %q should be valid, got error: %v", eol, errors)
		}
	}

	// Test invalid end_of_line
	source := `end_of_line = auto`
	doc, _ := parser.Parse([]byte(source))
	errors := Validate(doc)
	if len(errors) != 1 {
		t.Errorf("end_of_line = auto should produce 1 error, got %d", len(errors))
	}
}

// TestValidateMaxLineLengthSpecial tests max_line_length special value.
func TestValidateMaxLineLengthSpecial(t *testing.T) {
	source := `max_line_length = off`
	doc, _ := parser.Parse([]byte(source))
	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("max_line_length = off should be valid, got %d errors: %v", len(errors), errors)
	}
}

// TestValidateTabWidth tests tab_width validation.
func TestValidateTabWidth(t *testing.T) {
	source := `tab_width = 4`
	doc, _ := parser.Parse([]byte(source))
	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("tab_width = 4 should be valid, got %d errors", len(errors))
	}

	// Test out of range (no max)
	source = `tab_width = 0`
	doc, _ = parser.Parse([]byte(source))
	errors = Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("tab_width = 0 should be invalid (< min 1), got %d errors", len(errors))
	}
}

// TestValidateSectionWithValidProperties tests full section validation.
func TestValidateSectionWithValidProperties(t *testing.T) {
	source := `[*.go]
indent_style = tab
indent_size = tab
tab_width = 4
end_of_line = lf`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors for valid section, got %d: %v", len(errors), errors)
	}
}
