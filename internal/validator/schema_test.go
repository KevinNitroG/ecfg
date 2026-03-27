package validator

import (
	"testing"
)

// TestPropertySchemaCompletion verifies all properties are defined.
func TestPropertySchemaCompletion(t *testing.T) {
	requiredProperties := []string{
		"root",
		"indent_style",
		"indent_size",
		"tab_width",
		"end_of_line",
		"charset",
		"trim_trailing_whitespace",
		"insert_final_newline",
		"max_line_length",
		"spelling_language",
	}

	for _, prop := range requiredProperties {
		if _, ok := Schema[prop]; !ok {
			t.Errorf("Property %q not found in Schema", prop)
		}
	}

	if len(Schema) != len(requiredProperties) {
		t.Errorf("Expected %d properties, got %d", len(requiredProperties), len(Schema))
	}
}

// TestPropertyTypes verifies each property has the correct type.
func TestPropertyTypes(t *testing.T) {
	tests := []struct {
		property     string
		expectedType PropertyType
	}{
		{"root", PropertyTypeBoolean},
		{"indent_style", PropertyTypeEnum},
		{"indent_size", PropertyTypeInteger},
		{"tab_width", PropertyTypeInteger},
		{"end_of_line", PropertyTypeEnum},
		{"charset", PropertyTypeEnum},
		{"trim_trailing_whitespace", PropertyTypeBoolean},
		{"insert_final_newline", PropertyTypeBoolean},
		{"max_line_length", PropertyTypeInteger},
		{"spelling_language", PropertyTypeEnum},
	}

	for _, tt := range tests {
		schema := Schema[tt.property]
		if schema.Type != tt.expectedType {
			t.Errorf("Property %q: expected type %v, got %v", tt.property, tt.expectedType, schema.Type)
		}
	}
}

// TestEnumProperties verifies enum properties have correct valid values.
func TestEnumProperties(t *testing.T) {
	tests := []struct {
		property      string
		expectedVals  []string
		specialValues []string
	}{
		{"indent_style", []string{"tab", "space"}, nil},
		{"end_of_line", []string{"lf", "crlf", "cr"}, nil},
		{"charset", []string{"utf-8", "utf-8-bom", "utf-16be", "utf-16le", "latin1"}, nil},
		{"trim_trailing_whitespace", []string{"true", "false"}, nil},
		{"insert_final_newline", []string{"true", "false"}, nil},
		{"indent_size", nil, []string{"tab"}},
		{"max_line_length", nil, []string{"off"}},
	}

	for _, tt := range tests {
		schema := Schema[tt.property]

		// Check valid values
		if tt.expectedVals != nil {
			if len(schema.ValidValues) != len(tt.expectedVals) {
				t.Errorf("Property %q: expected %d valid values, got %d", tt.property, len(tt.expectedVals), len(schema.ValidValues))
			}
			for _, val := range tt.expectedVals {
				found := false
				for _, v := range schema.ValidValues {
					if v == val {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Property %q: missing valid value %q", tt.property, val)
				}
			}
		}

		// Check special values
		if tt.specialValues != nil {
			if len(schema.SpecialValues) != len(tt.specialValues) {
				t.Errorf("Property %q: expected %d special values, got %d", tt.property, len(tt.specialValues), len(schema.SpecialValues))
			}
			for _, val := range tt.specialValues {
				found := false
				for _, v := range schema.SpecialValues {
					if v == val {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Property %q: missing special value %q", tt.property, val)
				}
			}
		}
	}
}

// TestIntegerBounds verifies integer properties have correct min/max values.
func TestIntegerBounds(t *testing.T) {
	tests := []struct {
		property string
		minValue *int
		maxValue *int
	}{
		{"indent_size", ptrInt(1), ptrInt(8)},
		{"tab_width", ptrInt(1), nil},
		{"max_line_length", ptrInt(1), nil},
	}

	for _, tt := range tests {
		schema := Schema[tt.property]

		// Check min
		if tt.minValue == nil {
			if schema.MinValue != nil {
				t.Errorf("Property %q: expected no min value, got %v", tt.property, *schema.MinValue)
			}
		} else {
			if schema.MinValue == nil {
				t.Errorf("Property %q: expected min value %v, got nil", tt.property, *tt.minValue)
			} else if *schema.MinValue != *tt.minValue {
				t.Errorf("Property %q: expected min %v, got %v", tt.property, *tt.minValue, *schema.MinValue)
			}
		}

		// Check max
		if tt.maxValue == nil {
			if schema.MaxValue != nil {
				t.Errorf("Property %q: expected no max value, got %v", tt.property, *schema.MaxValue)
			}
		} else {
			if schema.MaxValue == nil {
				t.Errorf("Property %q: expected max value %v, got nil", tt.property, *tt.maxValue)
			} else if *schema.MaxValue != *tt.maxValue {
				t.Errorf("Property %q: expected max %v, got %v", tt.property, *tt.maxValue, *schema.MaxValue)
			}
		}
	}
}

// TestPreambleOnlyConstraint verifies root property is marked preamble-only.
func TestPreambleOnlyConstraint(t *testing.T) {
	schema := Schema["root"]
	if !schema.PreambleOnly {
		t.Errorf("Property 'root' should be preamble-only")
	}

	// All other properties should not be preamble-only
	for prop, schema := range Schema {
		if prop != "root" && schema.PreambleOnly {
			t.Errorf("Property %q should not be preamble-only", prop)
		}
	}
}

// TestSpecialValuesHandling verifies special values are tracked separately.
func TestSpecialValuesHandling(t *testing.T) {
	tests := []struct {
		property      string
		hasSpecial    bool
		specialValues []string
	}{
		{"indent_size", true, []string{"tab"}},
		{"max_line_length", true, []string{"off"}},
		{"indent_style", false, nil},
		{"charset", false, nil},
	}

	for _, tt := range tests {
		schema := Schema[tt.property]
		if tt.hasSpecial {
			if len(schema.SpecialValues) != len(tt.specialValues) {
				t.Errorf("Property %q: expected %d special values, got %d", tt.property, len(tt.specialValues), len(schema.SpecialValues))
			}
		} else {
			if len(schema.SpecialValues) > 0 {
				t.Errorf("Property %q: should have no special values", tt.property)
			}
		}
	}
}

// TestSchemaHasDescriptions verifies all properties have descriptions.
func TestSchemaHasDescriptions(t *testing.T) {
	for prop, schema := range Schema {
		if schema.Description == "" {
			t.Errorf("Property %q is missing description", prop)
		}
	}
}
