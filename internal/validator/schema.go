// Package validator provides schema-based validation for EditorConfig files.
//
// The validator checks AST nodes against the EditorConfig specification,
// detecting invalid property values, misplaced properties, and type violations.
// Validation errors include precise position information for LSP diagnostics.
package validator

// PropertyType identifies the expected type of a property value.
type PropertyType int

const (
	PropertyTypeString PropertyType = iota
	PropertyTypeInteger
	PropertyTypeBoolean
	PropertyTypeEnum
)

// String returns the name of the property type.
func (pt PropertyType) String() string {
	switch pt {
	case PropertyTypeString:
		return "string"
	case PropertyTypeInteger:
		return "integer"
	case PropertyTypeBoolean:
		return "boolean"
	case PropertyTypeEnum:
		return "enum"
	default:
		return "unknown"
	}
}

// PropertySchema defines validation rules for an EditorConfig property.
type PropertySchema struct {
	// Name is the property identifier (e.g., "indent_style")
	Name string

	// Type indicates the expected value type
	Type PropertyType

	// ValidValues lists allowed values for enum types
	// (e.g., ["tab", "space"] for indent_style)
	ValidValues []string

	// SpecialValues lists additional accepted values beyond ValidValues
	// (e.g., "tab" for indent_size which normally accepts integers)
	SpecialValues []string

	// MinValue is the minimum allowed integer value (nil for non-integers)
	MinValue *int

	// MaxValue is the maximum allowed integer value (nil for non-integers)
	MaxValue *int

	// PreambleOnly indicates the property can only appear in the preamble
	// (true for "root" property)
	PreambleOnly bool

	// Description is a human-readable explanation of the property
	Description string
}

// Schema maps property names to their validation rules.
// Includes all 9 official EditorConfig properties from the specification.
var Schema = map[string]PropertySchema{
	"root": {
		Name:         "root",
		Type:         PropertyTypeBoolean,
		ValidValues:  []string{"true", "false"},
		PreambleOnly: true,
		Description:  "Marks the end of configuration hierarchy search; must be in preamble",
	},
	"indent_style": {
		Name:        "indent_style",
		Type:        PropertyTypeEnum,
		ValidValues: []string{"tab", "space"},
		Description: "Indentation style: tab or space",
	},
	"indent_size": {
		Name:          "indent_size",
		Type:          PropertyTypeInteger,
		SpecialValues: []string{"tab"},
		MinValue:      ptrInt(1),
		MaxValue:      ptrInt(8),
		Description:   "Number of spaces per indentation level, or 'tab' for tab indentation",
	},
	"tab_width": {
		Name:        "tab_width",
		Type:        PropertyTypeInteger,
		MinValue:    ptrInt(1),
		Description: "Number of spaces per tab character",
	},
	"end_of_line": {
		Name:        "end_of_line",
		Type:        PropertyTypeEnum,
		ValidValues: []string{"lf", "crlf", "cr"},
		Description: "Line ending style: lf, crlf, or cr",
	},
	"charset": {
		Name:        "charset",
		Type:        PropertyTypeEnum,
		ValidValues: []string{"utf-8", "utf-8-bom", "utf-16be", "utf-16le", "latin1"},
		Description: "File character set encoding",
	},
	"trim_trailing_whitespace": {
		Name:        "trim_trailing_whitespace",
		Type:        PropertyTypeBoolean,
		ValidValues: []string{"true", "false"},
		Description: "Remove trailing whitespace from lines",
	},
	"insert_final_newline": {
		Name:        "insert_final_newline",
		Type:        PropertyTypeBoolean,
		ValidValues: []string{"true", "false"},
		Description: "Ensure file ends with a newline",
	},
	"max_line_length": {
		Name:          "max_line_length",
		Type:          PropertyTypeInteger,
		SpecialValues: []string{"off"},
		MinValue:      ptrInt(1),
		Description:   "Maximum line length in characters, or 'off' to disable",
	},
	"spelling_language": {
		Name: "spelling_language",
		Type: PropertyTypeEnum,
		ValidValues: []string{
			"en", "en-US", "en-GB", "en-AU", "en-CA", "en-NZ",
			"es", "es-ES", "es-MX", "fr", "fr-FR", "fr-CA",
			"de", "de-DE", "de-AT", "it", "it-IT", "pt", "pt-BR", "pt-PT",
			"nl", "nl-BE", "sv", "sv-SE", "da", "da-DK", "fi", "fi-FI",
			"no", "no-NO", "nb", "nb-NO", "ru", "ru-RU", "uk", "uk-UA",
			"pl", "pl-PL", "cs", "cs-CZ", "hu", "hu-HU", "ro", "ro-RO",
			"tr", "tr-TR", "el", "el-GR", "zh", "zh-CN", "zh-TW", "ja", "ja-JP",
			"ko", "ko-KR", "ar", "ar-SA", "he", "he-IL",
		},
		Description: "Natural language for spell checking (ISO 639 language code, optionally with ISO 3166 territory)",
	},
}

// ptrInt is a helper to create a pointer to an int.
func ptrInt(v int) *int {
	return &v
}
