package diagnostic

import (
	"fmt"
	"strings"

	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/resolver"
	"github.com/KevinNitroG/ecfg/internal/validator"
)

// Severity represents the severity level of a diagnostic message.
// Severity levels follow LSP protocol definitions.
type Severity int

const (
	// SeverityError indicates an error (invalid syntax or values).
	// Severity level: 1
	SeverityError Severity = 1

	// SeverityWarning indicates a warning (duplicates, conflicts).
	// Severity level: 2
	SeverityWarning Severity = 2

	// SeverityInfo indicates informational diagnostics (redundant properties).
	// Severity level: 3
	SeverityInfo Severity = 3

	// SeverityHint indicates hints/suggestions (currently unused).
	// Severity level: 4
	SeverityHint Severity = 4
)

// String returns the string representation of the severity level.
func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "unknown"
	}
}

// Diagnostic represents an LSP diagnostic message.
// Diagnostics are used to report errors, warnings, and information about the document.
type Diagnostic struct {
	// Range is the precise location of the diagnostic in the document.
	// Used by editors to underline or highlight the affected text.
	Range parser.Range

	// Severity is the severity level of the diagnostic.
	// Determines the icon and color used in the editor UI.
	Severity Severity

	// Message is a human-readable description of the issue.
	// Should be clear and actionable, including suggestions when possible.
	Message string

	// Source identifies the tool that produced the diagnostic.
	// Set to "ecfg" to identify diagnostics from this language server.
	Source string
}

// String returns a human-readable representation of the diagnostic.
func (d Diagnostic) String() string {
	return fmt.Sprintf("[%s] %s: %s", d.Severity, d.Range, d.Message)
}

// ToDiagnostics converts validation errors to LSP diagnostics.
// Maps each ValidationError to a Diagnostic with appropriate severity,
// message formatting, and precise range information for editor display.
func ToDiagnostics(errors []validator.ValidationError) []Diagnostic {
	var diagnostics []Diagnostic

	for _, err := range errors {
		severity := determineSeverity(err.Reason, err.IsWarning)
		message := formatMessage(err)

		diag := Diagnostic{
			Range:    err.Range,
			Severity: severity,
			Message:  message,
			Source:   "ecfg",
		}

		diagnostics = append(diagnostics, diag)
	}

	return diagnostics
}

// determineSeverity maps a validation error to the appropriate severity level.
// Rules:
// - Invalid values, misplaced properties, unknown properties → Error
// - Duplicate keys, logical conflicts → Warning
// - Redundant properties → Info
func determineSeverity(reason string, isWarning bool) Severity {
	if isWarning {
		return SeverityWarning
	}

	// Check for specific error types
	reasonLower := strings.ToLower(reason)

	// Info level for redundant properties (for future use)
	if strings.Contains(reasonLower, "redundant") {
		return SeverityInfo
	}

	// Default to Error for validation failures
	return SeverityError
}

// formatMessage creates a human-readable and actionable diagnostic message.
// Includes the property name, invalid value, and suggestion when applicable.
func formatMessage(err validator.ValidationError) string {
	// For invalid enum values, include valid options if available
	if strings.Contains(err.Reason, "invalid value") {
		return err.Reason
	}

	// For unknown properties
	if strings.Contains(err.Reason, "unknown property") {
		return err.Reason
	}

	// For duplicate keys, mention the first occurrence
	if strings.Contains(err.Reason, "duplicate key") {
		return fmt.Sprintf("Duplicate property %q. %s", err.Property, err.Reason)
	}

	// For logical conflicts
	if strings.Contains(err.Reason, "logical conflict") {
		return fmt.Sprintf("Logical conflict in property %q: %s", err.Property, err.Reason)
	}

	// For preamble-only constraint violations
	if strings.Contains(err.Reason, "preamble only") {
		return fmt.Sprintf("Property %q can only appear in preamble (before first section header)", err.Property)
	}

	// For value type errors (not integer, not boolean, etc.)
	if strings.Contains(err.Reason, "not a valid") {
		return err.Reason
	}

	// For out-of-bounds integers
	if strings.Contains(err.Reason, "less than minimum") || strings.Contains(err.Reason, "exceeds maximum") {
		return err.Reason
	}

	// Fallback to the raw reason
	return err.Reason
}

// AddRedundantPropertyDiagnostics adds info-level diagnostics for properties
// that are redundant (inherited from parent .editorconfig files with same value).
// The resolver is used to determine inheritance.
func AddRedundantPropertyDiagnostics(diagnostics []Diagnostic, doc *parser.Document, filePath string, res *resolver.Resolver) []Diagnostic {
	if doc == nil || res == nil {
		return diagnostics
	}

	// Find redundant properties
	redundant, err := res.FindRedundantProperties(filePath)
	if err != nil {
		return diagnostics
	}

	// Create info diagnostics for each redundant property
	for key, parentValue := range redundant {
		// Find the KeyValue in the document for this property
		kv := findKeyValue(doc, key)
		if kv != nil {
			diagnostics = append(diagnostics, Diagnostic{
				Range:    kv.KeyRange,
				Severity: SeverityInfo,
				Message:  fmt.Sprintf("Property %q is redundant; inherits value %q from parent .editorconfig", key, parentValue),
				Source:   "ecfg",
			})
		}
	}

	return diagnostics
}

// findKeyValue finds a KeyValue by property name in the document.
func findKeyValue(doc *parser.Document, propName string) *parser.KeyValue {
	propNameLower := strings.ToLower(propName)

	// Check preamble
	if doc.Preamble != nil {
		for _, kv := range doc.Preamble.Pairs {
			if strings.ToLower(kv.Key) == propNameLower {
				return kv
			}
		}
	}

	// Check sections
	for _, section := range doc.Sections {
		for _, kv := range section.Pairs {
			if strings.ToLower(kv.Key) == propNameLower {
				return kv
			}
		}
	}

	return nil
}
