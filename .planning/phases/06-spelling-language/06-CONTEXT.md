---
phase: 06-spelling-language
status: Ready for planning
---

<domain>
## Phase Boundary

Phase 6 adds the `spelling_language` EditorConfig property to the LSP server's schema, validation, hover, and completion features. This property allows users to specify a natural language for spell checking in editors that support it.
</domain>

<decisions>
## Implementation Decisions

### Property Schema
- Add spelling_language to Schema map in internal/validator/schema.go
- Type: enum (string with specific valid values per ISO 639/3166)
- Valid format: 2-letter (ss) or 5-letter (ss-TT) codes like "en", "en-US", "fr-FR"

### Autocomplete Filtering
- If a property is already set in the file, exclude it from completion suggestions
- This matches existing behavior for trim_trailing_whitespace (which is in schema already)
- Implement in completion.go by checking if key already exists in current section

### Validation
- Accept valid ISO 639 language codes
- Accept valid ISO 639 + ISO 3166 territory codes
- Reject invalid codes with appropriate diagnostic

### the agent's Discretion
Implementation choices are at the agent's discretion — use codebase conventions from existing properties.

</decisions>

(code_context>
## Existing Code Insights

### Reusable Assets
- internal/validator/schema.go: Schema map with 9 properties - add new entry here
- internal/lsp/hover.go: Property hover documentation
- internal/lsp/completion.go: Key and value completion

### Established Patterns
- PropertySchema struct defines validation rules
- Schema is map[string]PropertySchema
- Completion filters by PreambleOnly flag for root property

### Integration Points
- Schema changes propagate to validation, hover, completion automatically
- Add to Schema map in schema.go

</code_context>

<specifics>
## Specific Ideas

The completion should exclude properties that are already defined in the file (like trim_trailing_whitespace). If that's not possible, skip it.
</specifics>

<deferred>
## Deferred Ideas

None — focused scope on spelling_language property
</deferred>