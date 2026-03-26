---
phase: 03-lsp-intelligence-features
researched: 2026-03-26
status: complete
---

# Research: LSP Intelligence & Hover

## Executive Summary

This research covers implementation patterns for LSP hover and completion features in Go language servers. The phase builds on existing parser and validator infrastructure to provide intelligent EditorConfig property documentation and context-aware suggestions.

**Key Finding:** LSP hover and completion are synchronous request-response operations that map cursor positions to AST nodes, then return documentation or suggestions based on node context and schema metadata.

**Recommended Approach:**
1. Position-to-node resolution using existing AST Range data
2. Schema-driven documentation from validator.Schema
3. Context-aware filtering (preamble vs section)
4. Markdown formatting for rich tooltips

**Estimated Effort:** 2-3 plans, ~40% context each

---

## LSP Protocol Reference

### textDocument/hover

**Method:** `textDocument/hover`

**Request Parameters:**
```typescript
interface HoverParams {
  textDocument: TextDocumentIdentifier;
  position: Position;  // { line: uinteger, character: uinteger }
}
```

**Response:**
```typescript
interface Hover {
  contents: MarkedString | MarkedString[] | MarkupContent;
  range?: Range;  // Optional range to highlight
}

interface MarkupContent {
  kind: MarkupKind;  // "plaintext" | "markdown"
  value: string;
}
```

**When to Respond:**
- Cursor over property key → show property documentation
- Cursor over property value → show valid values and description
- Cursor over unknown property → return null (no hover)
- Outside identifiable node → return null

**Protocol Spec:** LSP 3.17 § 3.11 - Hover Request

### textDocument/completion

**Method:** `textDocument/completion`

**Request Parameters:**
```typescript
interface CompletionParams {
  textDocument: TextDocumentIdentifier;
  position: Position;
  context?: CompletionContext;
}

interface CompletionContext {
  triggerKind: CompletionTriggerKind;  // 1=Invoked, 2=TriggerCharacter, 3=TriggerForIncompleteCompletions
  triggerCharacter?: string;
}
```

**Response:**
```typescript
interface CompletionList {
  isIncomplete: boolean;
  items: CompletionItem[];
}

interface CompletionItem {
  label: string;
  kind?: CompletionItemKind;  // 10=Property, 12=Value, etc.
  detail?: string;
  documentation?: string | MarkupContent;
  insertText?: string;
  insertTextFormat?: InsertTextFormat;  // 1=PlainText, 2=Snippet
}
```

**Completion Triggers:**
- Typing before `=` → suggest property keys
- Typing after `=` → suggest valid values for detected property
- Explicit invocation (Ctrl+Space) → context-aware suggestions

**Protocol Spec:** LSP 3.17 § 3.12 - Completion Request

---

## Position-to-Node Resolution Strategy

### Algorithm: Find Node at Position

Given a `Position{line, character}` and a `Document` AST:

**Step 1: Locate containing section or preamble**
```go
func FindContext(doc *Document, pos Position) Context {
  // Check if in preamble range
  if doc.Preamble != nil && ContainsPosition(doc.Preamble.Range, pos) {
    return Context{InPreamble: true, Preamble: doc.Preamble}
  }
  
  // Find containing section
  for _, section := range doc.Sections {
    if ContainsPosition(section.Range, pos) {
      return Context{InSection: true, Section: section}
    }
  }
  
  return Context{} // Outside any identifiable context
}
```

**Step 2: Find specific KeyValue node**
```go
func FindKeyValue(pairs []*KeyValue, pos Position) (*KeyValue, NodePart) {
  for _, kv := range pairs {
    if ContainsPosition(kv.KeyRange, pos) {
      return kv, PartKey
    }
    if ContainsPosition(kv.ValueRange, pos) {
      return kv, PartValue
    }
  }
  return nil, PartNone
}
```

**Step 3: Return node + context**
```go
type ResolvedNode struct {
  KeyValue  *KeyValue
  Part      NodePart  // PartKey or PartValue
  InSection bool
  InPreamble bool
}
```

**Why this works:**
- AST nodes already have precise Range data (from Phase 01)
- KeyValue nodes track separate KeyRange and ValueRange
- O(n) scan acceptable (EditorConfig files are small, <100 lines typically)

---

## Hover Implementation Pattern

### Hover on Property Key

**Input:** Cursor on `indent_style` in `indent_style = tab`

**Resolution:**
1. FindKeyValue → returns KeyValue node, PartKey
2. Lookup schema: `validator.Schema["indent_style"]`
3. Build MarkupContent:

```markdown
**indent_style** _(enum)_

Indentation style: tab or space

**Valid values:**
- `tab`
- `space`
```

**Implementation:**
```go
func Hover(doc *Document, pos Position) *protocol.Hover {
  node := FindNodeAtPosition(doc, pos)
  if node == nil || node.Part != PartKey {
    return nil
  }
  
  schema, exists := validator.Schema[node.KeyValue.Key]
  if !exists {
    return nil // Unknown property
  }
  
  content := formatPropertyHover(schema)
  return &protocol.Hover{
    Contents: protocol.MarkupContent{
      Kind: protocol.Markdown,
      Value: content,
    },
    Range: node.KeyValue.KeyRange,
  }
}
```

### Hover on Property Value

**Input:** Cursor on `tab` in `indent_style = tab`

**Option 1 - Show same documentation as key:**
- Repeats key hover content
- User sees valid values while editing value

**Option 2 - No hover on value:**
- Simpler implementation
- Completion is more useful for values

**Recommendation:** Implement Option 1 (show property docs on value too) for consistency. Completion will provide interactive suggestions.

---

## Completion Implementation Pattern

### Completion: Property Keys (before `=`)

**Trigger:** User typing `ind<cursor>` before any `=`

**Context Detection:**
```go
func isBeforeEquals(doc *Document, pos Position) bool {
  // Find current line content
  line := getLineContent(doc, pos.Line)
  
  // Check if position is before = on this line
  equalsPos := strings.Index(line, "=")
  if equalsPos == -1 || pos.Character <= uint32(equalsPos) {
    return true
  }
  return false
}
```

**Completion Items:**
```go
func CompletePropertyKeys(inPreamble bool) []protocol.CompletionItem {
  items := []protocol.CompletionItem{}
  
  for name, schema := range validator.Schema {
    // Skip root if not in preamble
    if schema.PreambleOnly && !inPreamble {
      continue
    }
    
    items = append(items, protocol.CompletionItem{
      Label: name,
      Kind: protocol.CompletionItemKindProperty,
      Detail: schema.Type.String(),
      Documentation: protocol.MarkupContent{
        Kind: protocol.Markdown,
        Value: schema.Description,
      },
      InsertText: name + " = ",
      InsertTextFormat: protocol.InsertTextFormatPlainText,
    })
  }
  
  return items
}
```

**Context-Aware Filtering:**
- In preamble: Include `root` property
- In section: Exclude `root` (triggers validation error if placed in section)

### Completion: Property Values (after `=`)

**Trigger:** User typing `indent_style = <cursor>`

**Step 1: Identify property key**
```go
func findPropertyForValue(doc *Document, pos Position) string {
  // Get current line
  line := getLineContent(doc, pos.Line)
  
  // Extract key before =
  parts := strings.SplitN(line, "=", 2)
  if len(parts) < 2 {
    return ""
  }
  
  return strings.TrimSpace(parts[0])
}
```

**Step 2: Get valid values from schema**
```go
func CompletePropertyValues(propertyKey string) []protocol.CompletionItem {
  schema, exists := validator.Schema[propertyKey]
  if !exists {
    return []protocol.CompletionItem{}
  }
  
  items := []protocol.CompletionItem{}
  
  // Add enum values
  for _, value := range schema.ValidValues {
    items = append(items, protocol.CompletionItem{
      Label: value,
      Kind: protocol.CompletionItemKindValue,
      Detail: fmt.Sprintf("Valid value for %s", propertyKey),
      InsertText: value,
    })
  }
  
  // Add special values (e.g., "tab" for indent_size)
  for _, value := range schema.SpecialValues {
    items = append(items, protocol.CompletionItem{
      Label: value,
      Kind: protocol.CompletionItemKindValue,
      Detail: fmt.Sprintf("Special value for %s", propertyKey),
      InsertText: value,
    })
  }
  
  return items
}
```

**Edge Cases:**
- Integer properties: No completion (user types number)
- Boolean properties: Complete `true` and `false`
- Unknown property: No completion

---

## Schema Documentation Enrichment

The existing `validator.Schema` includes `Description` fields suitable for hover tooltips. No additional documentation is needed.

**Current Schema Example:**
```go
"indent_style": {
  Name:        "indent_style",
  Type:        PropertyTypeEnum,
  ValidValues: []string{"tab", "space"},
  Description: "Indentation style: tab or space",
}
```

**Hover Output:**
```markdown
**indent_style** _(enum)_

Indentation style: tab or space

**Valid values:**
- `tab`
- `space`
```

**Enhancement Opportunity (Future):**
Add URL reference to EditorConfig spec for each property. Not required for v1.

---

## Implementation Patterns from LSP Ecosystem

### Pattern 1: Two-Phase Resolution (gopls, rust-analyzer)

**Phase 1:** Position → AST node (synchronous, cached AST)
**Phase 2:** Node + Schema → Response (synchronous, pure function)

**Benefit:** Testable in isolation without LSP infrastructure

### Pattern 2: Handler Functions per Method

```go
type Server struct {
  documents map[string]*Document  // URI → parsed AST
}

func (s *Server) Hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
  doc := s.documents[params.TextDocument.URI]
  if doc == nil {
    return nil, nil
  }
  
  return computeHover(doc, params.Position)
}

func (s *Server) Completion(ctx context.Context, params *protocol.CompletionParams) (*protocol.CompletionList, error) {
  doc := s.documents[params.TextDocument.URI]
  if doc == nil {
    return &protocol.CompletionList{Items: []protocol.CompletionItem{}}, nil
  }
  
  return computeCompletion(doc, params.Position)
}
```

**Benefit:** Clear separation between LSP protocol handling and language logic

### Pattern 3: Pure Functions for Language Logic

```go
// Pure function - no side effects, easily testable
func computeHover(doc *parser.Document, pos protocol.Position) *protocol.Hover {
  node := FindNodeAtPosition(doc, pos)
  if node == nil {
    return nil
  }
  
  return buildHover(node)
}
```

**Benefit:** Unit tests don't need LSP server or JSON-RPC infrastructure

---

## Testing Strategy

### Unit Tests for Position Resolution

```go
func TestFindNodeAtPosition(t *testing.T) {
  source := "indent_style = tab\n"
  doc := parser.Parse(source)
  
  // Position on key
  node := FindNodeAtPosition(doc, Position{Line: 1, Character: 5})
  assert.Equal(t, PartKey, node.Part)
  assert.Equal(t, "indent_style", node.KeyValue.Key)
  
  // Position on value
  node = FindNodeAtPosition(doc, Position{Line: 1, Character: 16})
  assert.Equal(t, PartValue, node.Part)
  assert.Equal(t, "tab", node.KeyValue.Value)
}
```

### Unit Tests for Hover Content

```go
func TestBuildHoverForProperty(t *testing.T) {
  schema := validator.Schema["indent_style"]
  hover := buildPropertyHover(schema)
  
  assert.Contains(t, hover, "indent_style")
  assert.Contains(t, hover, "tab")
  assert.Contains(t, hover, "space")
}
```

### Integration Tests with LSP Server

```go
func TestHoverRequest(t *testing.T) {
  server := NewServer()
  server.documents["file:///test.editorconfig"] = parser.Parse("indent_style = tab")
  
  hover, err := server.Hover(context.Background(), &protocol.HoverParams{
    TextDocument: protocol.TextDocumentIdentifier{URI: "file:///test.editorconfig"},
    Position: protocol.Position{Line: 0, Character: 5},
  })
  
  require.NoError(t, err)
  assert.NotNil(t, hover)
  assert.Contains(t, hover.Contents.Value, "indent_style")
}
```

**Test Fixtures:**
- Use existing testdata from Phase 01
- Add new fixtures for cursor positions (similar to `positions/` directory)

---

## Dependencies

### Internal Packages (Already Exist)
- `internal/parser` — AST with Range tracking ✅
- `internal/validator` — Schema with descriptions ✅
- `internal/diagnostic` — Error types ✅

### External Dependencies (To Add)
- `go.lsp.dev/protocol` — LSP types and constants
  - Provides: Position, Range, Hover, CompletionItem, etc.
  - Why: Type-safe LSP protocol without manual JSON handling
  - Version: v0.12.0+ (LSP 3.17 support)

**Installation:**
```bash
go get go.lsp.dev/protocol@v0.12.0
```

**Alternative:** `github.com/tliron/glsp` (mentioned in PROJECT.md)
- More batteries-included (includes server framework)
- Heavier dependency
- Not necessary for just types

**Recommendation:** Use `go.lsp.dev/protocol` for types only. Server setup deferred to Phase 5.

---

## Architecture Decision: Package Structure

### Option 1: Single `internal/lsp` package
```
internal/lsp/
  hover.go         // Hover logic
  completion.go    // Completion logic
  position.go      // Position resolution utilities
```

**Pros:** Simple, all LSP logic together
**Cons:** Will grow in Phase 5 with server lifecycle

### Option 2: Separate `internal/intelligence` package
```
internal/intelligence/
  hover.go
  completion.go
  resolver.go      // Position → node resolution

internal/lsp/      // Phase 5
  server.go        // Server lifecycle
  handlers.go      // Protocol handlers
```

**Pros:** Clear separation between language intelligence and protocol handling
**Cons:** Extra package layer

**Recommendation:** Option 1 (`internal/lsp`) for now. Simple and sufficient. Refactor in Phase 5 if package grows large.

---

## Standard Stack

### LSP Protocol Library
**Choice:** `go.lsp.dev/protocol`

**Why:**
- Official Go implementation of LSP 3.17 types
- No server framework (we control initialization)
- Used by gopls and other production servers
- Active maintenance

**Alternative considered:** `github.com/tliron/glsp`
- Includes server framework (too opinionated for Phase 3)
- Better for Phase 5 if we need server lifecycle helpers

### Markdown Formatting
**Choice:** Standard library `fmt` + string templates

**Why:**
- LSP MarkupContent is just a string
- No complex rendering needed
- Standard library sufficient for property docs

**Not needed:** Markdown parsing library (we're generating, not parsing)

---

## Common Pitfalls

### Pitfall 1: LSP Position is 0-indexed line and character

**Issue:** LSP Position uses 0-indexed line (unlike parser's 1-indexed Line)

**Solution:**
```go
func lspPositionToParser(lspPos protocol.Position) parser.Position {
  return parser.Position{
    Line: lspPos.Line + 1,  // Convert 0-indexed to 1-indexed
    Column: lspPos.Character,  // Already 0-indexed
  }
}
```

**Verification:** Write explicit test for line number conversion

### Pitfall 2: Position between tokens

**Issue:** Cursor might be on whitespace between key and `=`

**Solution:**
- Expand search to nearest token on same line
- Or return nil (no hover) for whitespace positions
- **Recommendation:** Return nil for whitespace (simpler, matches gopls behavior)

### Pitfall 3: Completing in wrong context

**Issue:** User types `root` inside a section, completion suggests it

**Solution:**
- Always check `schema.PreambleOnly` before suggesting
- Filter based on `inPreamble` boolean from context resolution
- **Already validated:** Validator will error if user completes it anyway

### Pitfall 4: UTF-8 handling in character offsets

**Issue:** LSP Position.Character is UTF-16 code units, not bytes

**Solution:**
- Parser already tracks byte offset and UTF-8 runes correctly
- Use parser.Column (rune-based) for matching
- LSP libraries handle UTF-16 ↔ UTF-8 conversion in protocol layer
- **Test with:** Unicode characters in property names/values

---

## Requirements Mapping

Phase 3 requirements from ROADMAP.md:

| Requirement | Covered By | Implementation |
|-------------|-----------|----------------|
| HOVER-01: Provides Markdown hover tooltip for property keys | Plan 01 | `hover.go` with schema lookup |
| HOVER-02: Hover includes official spec description | Plan 01 | Schema.Description field |
| HOVER-03: Hover includes valid values for property | Plan 01 | Schema.ValidValues formatted as list |
| HOVER-04: Hover works when cursor on key name | Plan 01 | Position resolution → KeyRange |
| COMP-01: Completion suggestions for property keys before `=` | Plan 02 | Filter Schema map, add completion items |
| COMP-02: Completion suggestions for enum values after `=` | Plan 02 | Schema.ValidValues → completion items |
| COMP-03: Context-aware completion (no `root` in sections) | Plan 02 | Check PreambleOnly + inPreamble flag |
| COMP-04: Completion items include documentation | Plan 02 | CompletionItem.Documentation = Schema.Description |
| COMP-05: Completion suggests only valid values for property | Plan 02 | Schema lookup by property key |

**Coverage:** All 9 requirements mapped to 2 plans

---

## Recommended Plan Breakdown

### Plan 01: Position Resolution and Hover
**Effort:** ~40% context, 2-3 tasks
**Deliverables:**
- Position → node resolution utilities
- Hover handler implementation
- Schema → Markdown formatting
- Unit tests for resolution and hover content

### Plan 02: Context-Aware Completion
**Effort:** ~40% context, 2-3 tasks
**Deliverables:**
- Property key completion (with preamble filtering)
- Property value completion (enum + special values)
- Completion item documentation
- Unit tests for context-aware filtering

**Dependencies:**
- Plan 02 depends on Plan 01 (uses position resolution utilities)
- Both can share test fixtures

**Parallel Opportunity:** None (Plan 02 needs resolution from Plan 01)

---

## Phase Success Criteria

From ROADMAP.md:

1. ✅ Hovering over `indent_style` shows spec description and valid values (`tab`, `space`)
   - **Verification:** Unit test + manual test in editor

2. ✅ Typing before `=` suggests all valid property keys for context
   - **Verification:** Unit test with preamble vs section completion

3. ✅ Typing after `=` for `end_of_line` suggests `lf`, `crlf`, `cr` only
   - **Verification:** Unit test for enum value completion

4. ✅ Completion does not suggest `root` when cursor inside a section
   - **Verification:** Unit test checking PreambleOnly filtering

5. ✅ All completion items include brief documentation snippets
   - **Verification:** Unit test checking CompletionItem.Documentation field

**Testing Approach:**
- Pure function unit tests (no LSP server needed)
- Use table-driven tests with various cursor positions
- Fixtures from Phase 01 testdata

---

## References

- LSP 3.17 Specification: https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/
- go.lsp.dev/protocol documentation: https://pkg.go.dev/go.lsp.dev/protocol
- EditorConfig Specification: https://spec.editorconfig.org/
- gopls (reference implementation): https://github.com/golang/tools/tree/master/gopls/internal/lsp

---

*Research completed: 2026-03-26*
*Ready for planning: Phase 03 (2 plans recommended)*
