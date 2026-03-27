package validator

import (
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
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

// TestValidateDuplicateKeyInSection tests detection of duplicate keys within a section.
func TestValidateDuplicateKeyInSection(t *testing.T) {
	source := `[*.go]
indent_size = 4
indent_size = 2`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for duplicate key, got %d: %v", len(errors), errors)
	}

	err := errors[0]
	if err.Property != "indent_size" {
		t.Errorf("Expected property 'indent_size', got %q", err.Property)
	}
	if !err.IsWarning {
		t.Error("Expected IsWarning=true for duplicate key")
	}
	if err.Value != "2" {
		t.Errorf("Expected value '2' (second occurrence), got %q", err.Value)
	}
}

// TestValidateDuplicateKeyInPreamble tests detection of duplicate keys in preamble.
func TestValidateDuplicateKeyInPreamble(t *testing.T) {
	source := `root = true
root = false
[*.go]
indent_style = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for duplicate root in preamble, got %d: %v", len(errors), errors)
	}

	err := errors[0]
	if err.Property != "root" {
		t.Errorf("Expected property 'root', got %q", err.Property)
	}
	if !err.IsWarning {
		t.Error("Expected IsWarning=true for duplicate key")
	}
}

// TestValidateConflictIndentStyleTabWithNumericIndentSize tests logical conflict detection.
func TestValidateConflictIndentStyleTabWithNumericIndentSize(t *testing.T) {
	source := `[*.go]
indent_style = tab
indent_size = 4`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for tab + numeric size conflict, got %d: %v", len(errors), errors)
	}

	err := errors[0]
	if err.Property != "indent_size" {
		t.Errorf("Expected property 'indent_size', got %q", err.Property)
	}
	if !err.IsWarning {
		t.Error("Expected IsWarning=true for conflict")
	}
	if err.Value != "4" {
		t.Errorf("Expected value '4', got %q", err.Value)
	}
}

// TestValidateNoConflictIndentStyleSpaceWithNumericIndentSize tests that space + numeric is valid.
func TestValidateNoConflictIndentStyleSpaceWithNumericIndentSize(t *testing.T) {
	source := `[*.go]
indent_style = space
indent_size = 4`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors for space + numeric size, got %d: %v", len(errors), errors)
	}
}

// TestValidateNoConflictIndentStyleTabWithTabIndentSize tests that tab + tab is valid.
func TestValidateNoConflictIndentStyleTabWithTabIndentSize(t *testing.T) {
	source := `[*.go]
indent_style = tab
indent_size = tab`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	if len(errors) != 0 {
		t.Fatalf("Expected 0 errors for tab + tab indent size, got %d: %v", len(errors), errors)
	}
}

// TestValidateMultipleDuplicates tests detection of multiple duplicate keys.
func TestValidateMultipleDuplicates(t *testing.T) {
	source := `[*.go]
indent_style = tab
indent_style = space
indent_size = 4
indent_size = 2`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	// Should have: 1 duplicate warning for indent_style + 1 duplicate warning for indent_size + 1 conflict warning (tab + 4)
	if len(errors) != 3 {
		t.Fatalf("Expected 3 errors (2 duplicates + 1 conflict), got %d: %v", len(errors), errors)
	}

	// Check all are marked as warnings
	for i, err := range errors {
		if !err.IsWarning {
			t.Errorf("Error %d: expected IsWarning=true, got false", i)
		}
	}
}

// TestValidateCombinedErrorsAndWarnings tests mix of validation errors and conflict warnings.
func TestValidateCombinedErrorsAndWarnings(t *testing.T) {
	source := `[*.go]
indent_style = invalid
indent_style = tab
indent_size = 4`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	// Should have: 1 error for invalid enum, 1 warning for duplicate indent_style
	if len(errors) != 2 {
		t.Fatalf("Expected 2 errors for invalid + duplicate, got %d: %v", len(errors), errors)
	}

	// Find error and warning
	var hasError, hasWarning bool
	for _, err := range errors {
		if err.IsWarning {
			hasWarning = true
		} else {
			hasError = true
		}
	}

	if !hasError {
		t.Error("Expected an error (invalid enum)")
	}
	if !hasWarning {
		t.Error("Expected a warning (duplicate)")
	}
}

// TestValidateSectionHeaderValidPatterns tests valid glob patterns in section headers.
func TestValidateSectionHeaderValidPatterns(t *testing.T) {
	validPatterns := []string{
		"*.go",
		"*.js",
		"*.py",
		"[*.go]",
		"src/*.js",
		"**/*.ts",
		"project/*.{js,ts}",
		"?orld",
		"file[0-9].txt",
	}

	for _, pattern := range validPatterns {
		source := "[" + pattern + "]\nindent_style = tab"
		doc, _ := parser.Parse([]byte(source))

		errors := Validate(doc)
		if len(errors) != 0 {
			t.Errorf("Pattern %q should be valid, got errors: %v", pattern, errors)
		}
	}
}

// TestValidateSectionHeaderInvalidPatterns tests invalid glob patterns in section headers.
func TestValidateSectionHeaderInvalidPatterns(t *testing.T) {
	invalidPatterns := []string{
		"[",  // Incomplete bracket (causes panic in fnmatch)
		"[*", // Incomplete bracket
	}

	for _, pattern := range invalidPatterns {
		source := "[" + pattern + "]\nindent_style = tab"
		doc, _ := parser.Parse([]byte(source))

		errors := Validate(doc)
		if len(errors) == 0 {
			t.Errorf("Pattern %q should be invalid, but no errors found", pattern)
		}
	}
}

// TestValidateMultipleSectionsWithMixedPatterns tests multiple sections with valid and invalid patterns.
func TestValidateMultipleSectionsWithMixedPatterns(t *testing.T) {
	// Use a pattern that triggers validation but not parse errors
	// "[*" triggers validation error but is well-formed at parse time
	source := `[*.go]
indent_style = tab
[*
indent_style = space
[*.js]
indent_style = space`
	doc, _ := parser.Parse([]byte(source))

	errors := Validate(doc)
	// Should have at least one error for the invalid pattern [*
	if len(errors) < 1 {
		t.Fatalf("Expected at least 1 error for invalid pattern, got %d: %v", len(errors), errors)
	}
}
