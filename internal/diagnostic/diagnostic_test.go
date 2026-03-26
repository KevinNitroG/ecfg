package diagnostic

import (
	"strings"
	"testing"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/validator"
)

// TestSeverityString tests the string representation of severity levels.
func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityError, "error"},
		{SeverityWarning, "warning"},
		{SeverityInfo, "info"},
		{SeverityHint, "hint"},
	}

	for _, test := range tests {
		if result := test.severity.String(); result != test.expected {
			t.Errorf("Severity(%d).String() = %q, want %q", test.severity, result, test.expected)
		}
	}
}

// TestDetermineSeverityError tests that errors are mapped correctly.
func TestDetermineSeverityError(t *testing.T) {
	tests := []struct {
		reason    string
		isWarning bool
		expected  Severity
	}{
		// Errors (isWarning=false)
		{"invalid value", false, SeverityError},
		{"unknown property", false, SeverityError},
		{"not a valid integer", false, SeverityError},

		// Warnings (isWarning=true)
		{"duplicate key", true, SeverityWarning},
		{"logical conflict", true, SeverityWarning},

		// Info (with "redundant" keyword)
		{"redundant property", false, SeverityInfo},
	}

	for _, test := range tests {
		result := determineSeverity(test.reason, test.isWarning)
		if result != test.expected {
			t.Errorf("determineSeverity(%q, %v) = %v, want %v", test.reason, test.isWarning, result, test.expected)
		}
	}
}

// TestFormatMessage tests that diagnostic messages are human-readable.
func TestFormatMessage(t *testing.T) {
	tests := []struct {
		err           validator.ValidationError
		expectedMatch string // Substring that should appear in message
	}{
		{
			err: validator.ValidationError{
				Property:  "indent_style",
				Value:     "invalid",
				Reason:    "invalid value \"invalid\" for indent_style; valid values are: tab, space",
				Range:     parser.Range{},
				IsWarning: false,
			},
			expectedMatch: "invalid value",
		},
		{
			err: validator.ValidationError{
				Property:  "indent_size",
				Value:     "indent_size",
				Reason:    "duplicate key (first defined at line 5)",
				Range:     parser.Range{},
				IsWarning: true,
			},
			expectedMatch: "Duplicate property",
		},
		{
			err: validator.ValidationError{
				Property:  "indent_size",
				Value:     "4",
				Reason:    "logical conflict: indent_style=tab with numeric indent_size (use 'tab' or remove indent_size)",
				Range:     parser.Range{},
				IsWarning: true,
			},
			expectedMatch: "Logical conflict",
		},
		{
			err: validator.ValidationError{
				Property:  "root",
				Value:     "true",
				Reason:    "property \"root\" can only appear in preamble, not in sections",
				Range:     parser.Range{},
				IsWarning: false,
			},
			expectedMatch: "preamble",
		},
	}

	for i, test := range tests {
		result := formatMessage(test.err)
		if !strings.Contains(result, test.expectedMatch) {
			t.Errorf("Test %d: formatMessage() = %q, expected to contain %q", i, result, test.expectedMatch)
		}
	}
}

// TestToDiagnosticsEmpty tests that empty validation errors produce empty diagnostics.
func TestToDiagnosticsEmpty(t *testing.T) {
	errors := []validator.ValidationError{}
	diagnostics := ToDiagnostics(errors)

	if len(diagnostics) != 0 {
		t.Errorf("ToDiagnostics([]) = %d diagnostics, want 0", len(diagnostics))
	}
}

// TestToDiagnosticsErrorSeverity tests that validation errors map to error severity.
func TestToDiagnosticsErrorSeverity(t *testing.T) {
	errors := []validator.ValidationError{
		{
			Property:  "indent_style",
			Value:     "invalid",
			Reason:    "invalid value \"invalid\" for indent_style; valid values are: tab, space",
			Range:     parser.Range{},
			IsWarning: false,
		},
	}

	diagnostics := ToDiagnostics(errors)
	if len(diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Severity != SeverityError {
		t.Errorf("Expected Severity Error, got %v", diag.Severity)
	}
	if diag.Source != "ecfg" {
		t.Errorf("Expected Source 'ecfg', got %q", diag.Source)
	}
	if !strings.Contains(diag.Message, "invalid_value") && !strings.Contains(diag.Message, "invalid value") {
		t.Errorf("Expected message to mention invalid value, got %q", diag.Message)
	}
}

// TestToDiagnosticsWarningSeverity tests that duplicate errors map to warning severity.
func TestToDiagnosticsWarningSeverity(t *testing.T) {
	errors := []validator.ValidationError{
		{
			Property:  "indent_size",
			Value:     "4",
			Reason:    "duplicate key (first defined at line 5)",
			Range:     parser.Range{},
			IsWarning: true,
		},
	}

	diagnostics := ToDiagnostics(errors)
	if len(diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Severity != SeverityWarning {
		t.Errorf("Expected Severity Warning, got %v", diag.Severity)
	}
}

// TestToDiagnosticsPreservesRange tests that ranges are preserved in diagnostics.
func TestToDiagnosticsPreservesRange(t *testing.T) {
	testRange := parser.Range{
		Start: parser.Position{Line: 5, Column: 10, Offset: 42},
		End:   parser.Position{Line: 5, Column: 17, Offset: 49},
	}

	errors := []validator.ValidationError{
		{
			Property:  "indent_size",
			Value:     "invalid",
			Reason:    "not a valid integer",
			Range:     testRange,
			IsWarning: false,
		},
	}

	diagnostics := ToDiagnostics(errors)
	if len(diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Range != testRange {
		t.Errorf("Expected Range %v, got %v", testRange, diag.Range)
	}
}

// TestToDiagnosticsMultipleErrors tests conversion of multiple errors to diagnostics.
func TestToDiagnosticsMultipleErrors(t *testing.T) {
	errors := []validator.ValidationError{
		{
			Property:  "indent_style",
			Value:     "invalid",
			Reason:    "invalid value \"invalid\"",
			Range:     parser.Range{},
			IsWarning: false,
		},
		{
			Property:  "indent_size",
			Value:     "4",
			Reason:    "duplicate key (first defined at line 2)",
			Range:     parser.Range{},
			IsWarning: true,
		},
	}

	diagnostics := ToDiagnostics(errors)
	if len(diagnostics) != 2 {
		t.Fatalf("Expected 2 diagnostics, got %d", len(diagnostics))
	}

	// Check first diagnostic (error)
	if diagnostics[0].Severity != SeverityError {
		t.Errorf("First diagnostic: expected Error severity, got %v", diagnostics[0].Severity)
	}

	// Check second diagnostic (warning)
	if diagnostics[1].Severity != SeverityWarning {
		t.Errorf("Second diagnostic: expected Warning severity, got %v", diagnostics[1].Severity)
	}
}

// TestDiagnosticString tests string representation of diagnostics.
func TestDiagnosticString(t *testing.T) {
	diag := Diagnostic{
		Range:    parser.Range{Start: parser.Position{Line: 1, Column: 0}, End: parser.Position{Line: 1, Column: 5}},
		Severity: SeverityError,
		Message:  "Invalid value",
		Source:   "ecfg",
	}

	result := diag.String()
	if !strings.Contains(result, "error") {
		t.Errorf("String() should contain 'error', got %q", result)
	}
	if !strings.Contains(result, "Invalid value") {
		t.Errorf("String() should contain message, got %q", result)
	}
}

// TestToDiagnosticsConflictDetection tests that conflicts are detected and converted to warnings.
func TestToDiagnosticsConflictDetection(t *testing.T) {
	errors := []validator.ValidationError{
		{
			Property:  "indent_size",
			Value:     "4",
			Reason:    "logical conflict: indent_style=tab with numeric indent_size (use 'tab' or remove indent_size)",
			Range:     parser.Range{},
			IsWarning: true,
		},
	}

	diagnostics := ToDiagnostics(errors)
	if len(diagnostics) != 1 {
		t.Fatalf("Expected 1 diagnostic, got %d", len(diagnostics))
	}

	diag := diagnostics[0]
	if diag.Severity != SeverityWarning {
		t.Errorf("Expected Warning severity for conflict, got %v", diag.Severity)
	}
	if !strings.Contains(diag.Message, "Logical conflict") {
		t.Errorf("Expected 'Logical conflict' in message, got %q", diag.Message)
	}
}
