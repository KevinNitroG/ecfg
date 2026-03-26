package validator

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KevinNitroG/ecfg/internal/parser"
)

// ValidationError represents a validation error with precise position information.
type ValidationError struct {
	// Property is the property name that failed validation
	Property string

	// Value is the invalid value provided
	Value string

	// Reason is a human-readable explanation of why the value is invalid
	Reason string

	// Range is the position in the source file for LSP diagnostics
	Range parser.Range
}

// Validate checks an AST document against the EditorConfig schema.
// It returns a slice of ValidationErrors for any violations found.
// An empty slice means the document is valid.
func Validate(doc *parser.Document) []ValidationError {
	var errors []ValidationError

	if doc == nil {
		return errors
	}

	// Validate preamble
	if doc.Preamble != nil {
		for _, kv := range doc.Preamble.Pairs {
			if err := validateProperty(kv, true); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	// Validate sections
	for _, section := range doc.Sections {
		for _, kv := range section.Pairs {
			if err := validateProperty(kv, false); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	return errors
}

// validateProperty validates a single key-value pair against the schema.
// inPreamble indicates whether the property is in the preamble.
func validateProperty(kv *parser.KeyValue, inPreamble bool) *ValidationError {
	if kv == nil {
		return nil
	}

	// Normalize property name to lowercase
	propName := strings.ToLower(kv.Key)

	// Check if property exists in schema
	schema, exists := Schema[propName]
	if !exists {
		return &ValidationError{
			Property: kv.Key,
			Value:    kv.Value,
			Reason:   fmt.Sprintf("unknown property %q", kv.Key),
			Range:    kv.ValueRange,
		}
	}

	// Check preamble-only constraint
	if schema.PreambleOnly && !inPreamble {
		return &ValidationError{
			Property: kv.Key,
			Value:    kv.Value,
			Reason:   fmt.Sprintf("property %q can only appear in preamble, not in sections", kv.Key),
			Range:    kv.KeyRange, // Underline the key name for clarity
		}
	}

	// Validate value based on property type
	switch schema.Type {
	case PropertyTypeEnum:
		return validateEnumValue(kv, schema)
	case PropertyTypeInteger:
		return validateIntegerValue(kv, schema)
	case PropertyTypeBoolean:
		return validateBooleanValue(kv, schema)
	case PropertyTypeString:
		// String properties accept any value
		return nil
	}

	return nil
}

// validateEnumValue validates that the value is in the list of valid values.
func validateEnumValue(kv *parser.KeyValue, schema PropertySchema) *ValidationError {
	value := strings.ToLower(kv.Value)

	// Check valid values
	for _, valid := range schema.ValidValues {
		if value == valid {
			return nil
		}
	}

	// Check special values
	for _, special := range schema.SpecialValues {
		if value == special {
			return nil
		}
	}

	// Value not found in valid or special values
	validList := append([]string{}, schema.ValidValues...)
	validList = append(validList, schema.SpecialValues...)

	return &ValidationError{
		Property: kv.Key,
		Value:    kv.Value,
		Reason:   fmt.Sprintf("invalid value %q for %s; valid values are: %s", kv.Value, kv.Key, strings.Join(validList, ", ")),
		Range:    kv.ValueRange,
	}
}

// validateIntegerValue validates that the value is a valid integer within bounds.
func validateIntegerValue(kv *parser.KeyValue, schema PropertySchema) *ValidationError {
	value := strings.ToLower(kv.Value)

	// Check special values first
	for _, special := range schema.SpecialValues {
		if value == special {
			return nil
		}
	}

	// Try to parse as integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return &ValidationError{
			Property: kv.Key,
			Value:    kv.Value,
			Reason:   fmt.Sprintf("%q is not a valid integer for %s", kv.Value, kv.Key),
			Range:    kv.ValueRange,
		}
	}

	// Check minimum bound
	if schema.MinValue != nil && intValue < *schema.MinValue {
		return &ValidationError{
			Property: kv.Key,
			Value:    kv.Value,
			Reason:   fmt.Sprintf("value %d is less than minimum %d for %s", intValue, *schema.MinValue, kv.Key),
			Range:    kv.ValueRange,
		}
	}

	// Check maximum bound
	if schema.MaxValue != nil && intValue > *schema.MaxValue {
		return &ValidationError{
			Property: kv.Key,
			Value:    kv.Value,
			Reason:   fmt.Sprintf("value %d exceeds maximum %d for %s", intValue, *schema.MaxValue, kv.Key),
			Range:    kv.ValueRange,
		}
	}

	return nil
}

// validateBooleanValue validates that the value is a valid boolean.
func validateBooleanValue(kv *parser.KeyValue, schema PropertySchema) *ValidationError {
	value := strings.ToLower(kv.Value)

	// Check if value is in valid values (true/false)
	for _, valid := range schema.ValidValues {
		if value == valid {
			return nil
		}
	}

	return &ValidationError{
		Property: kv.Key,
		Value:    kv.Value,
		Reason:   fmt.Sprintf("invalid value %q for boolean property %s; must be 'true' or 'false'", kv.Value, kv.Key),
		Range:    kv.ValueRange,
	}
}
