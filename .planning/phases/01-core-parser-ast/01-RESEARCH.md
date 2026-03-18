# Phase 1 Research: Core Parser & AST

**Phase:** 1 - Core Parser & AST  
**Researched:** 2026-03-18  
**Status:** Complete

## Research Question

"What do I need to know to PLAN this phase well?"

Phase 1 builds the EditorConfig parser with AST and precise position tracking for LSP features. Every subsequent phase depends on this foundation.

## Standard Stack

### EditorConfig Specification (v0.17.2)

**File Format:** INI-like with specific rules
- Lines processed one at a time (top to bottom)
- Line types: Blank, Comment (`#` or `;`), Section Header (`[glob]`), Key-Value Pair (`key = value`)
- UTF-8 encoded, LF or CRLF line separators
- Leading/trailing whitespace trimmed per line
- Keys and values trimmed of whitespace (but internal whitespace preserved)

**Structure:**
```
# Optional preamble (before first section)
root = true

# Section with glob pattern
[*.go]
indent_style = tab
indent_size = 4

[*.{js,ts}]
indent_style = space
indent_size = 2
```

**Critical Elements:**
1. **Preamble:** Lines before first section (optional, can contain `root` key)
2. **Section Headers:** `[glob_pattern]` — defines file matching rules
3. **Key-Value Pairs:** `property = value` — configuration properties
4. **Comments:** `#` or `;` prefix — ignored by parser
5. **Whitespace:** Leading/trailing trimmed, internal preserved

**Glob Patterns (for section headers):**
- `*` — any characters except `/`
- `**` — any characters including `/`
- `?` — single character except `/`
- `[seq]` — character class
- `[!seq]` — negated character class
- `{s1,s2}` — alternation
- `{1..5}` — numeric range

**Supported Properties (Phase 1 just parses, Phase 2 validates):**
- `root` (preamble only)
- `indent_style`, `indent_size`, `tab_width`
- `end_of_line`, `charset`
- `trim_trailing_whitespace`, `insert_final_newline`
- `spelling_language`

**Limits:**
- Section name: up to 1024 characters
- Key: up to 1024 characters
- Value: up to 4096 characters

### Existing Go Parser (editorconfig-core-go)

**Repository:** https://github.com/editorconfig/editorconfig-core-go

**Current Approach:**
- Uses `gopkg.in/ini.v1` for INI parsing
- **Does NOT track positions** — AST lacks line/column data
- Designed for file resolution, not LSP

**Why We Can't Use It:**
```go
// Their structure (no position data):
type Definition struct {
	Selector       string
	IndentStyle    string
	IndentSize     string
	// ... other properties
	Raw            map[string]string
}
```

**What LSP Needs:**
```go
// Every node needs:
type Position struct {
	Line   int  // 1-indexed
	Column int  // 0-indexed (LSP standard)
}

type Range struct {
	Start Position
	End   Position
}

// Every AST node must have Range
```

**However, we CAN reuse:**
- Glob matching logic (FnmatchCase function)
- Property normalization (lowercasing, trim logic)
- File traversal strategy (finding parent `.editorconfig` files)
- Resolution algorithm (merge logic for overlapping sections)

## Architecture Patterns

### Lexer-Parser Split (Recommended for LSP)

**Lexer (Tokenizer):**
- Scans raw text → stream of tokens
- Each token has: Type, Value, Start Position, End Position
- Token types: COMMENT, SECTION_START, SECTION_END, IDENTIFIER, EQUALS, VALUE, NEWLINE, EOF

**Parser:**
- Consumes tokens → AST nodes
- Each AST node has: Type, Children, Range
- Node types: Document, Preamble, Section, KeyValue, Comment

**Why This Approach:**
1. **Position tracking is natural** — lexer tracks byte offset, line, column as it scans
2. **Error recovery is easier** — parser can skip malformed tokens and continue
3. **LSP-friendly** — can re-lex partial document for incremental updates (Phase 2+)
4. **Separation of concerns** — lexer = "what characters", parser = "what structure"

### AST Design for LSP

**Core Principles:**
1. **Every node has Range** — enables hover, diagnostics, completion
2. **Preserve all source text** — including comments, whitespace (for formatting later)
3. **Parent pointers optional** — simplifies tree walking for analysis
4. **Flat arrays preferred** — easier to serialize, iterate

**Minimal AST Structure:**

```go
package parser

// Position represents a 0-indexed byte offset and 1-indexed line/column
type Position struct {
	Offset int // byte offset in source
	Line   int // 1-indexed line number
	Column int // 0-indexed column (LSP standard)
}

type Range struct {
	Start Position
	End   Position
}

// Token from lexer
type Token struct {
	Type  TokenType
	Value string
	Range Range
}

type TokenType int

const (
	TokenEOF TokenType = iota
	TokenComment
	TokenSectionStart // [
	TokenSectionEnd   // ]
	TokenIdentifier   // key or glob pattern content
	TokenEquals       // =
	TokenValue        // value after =
	TokenNewline
)

// AST Nodes
type Node interface {
	GetRange() Range
	Type() NodeType
}

type NodeType int

const (
	NodeDocument NodeType = iota
	NodeComment
	NodePreamble
	NodeSection
	NodeKeyValue
)

type Document struct {
	Range    Range
	Preamble *Preamble // nil if no preamble
	Sections []*Section
	Comments []*Comment // top-level comments
}

type Preamble struct {
	Range    Range
	Pairs    []*KeyValue
	Comments []*Comment
}

type Section struct {
	Range    Range
	Header   string // glob pattern (without brackets)
	HeaderRange Range // position of [header] for diagnostics
	Pairs    []*KeyValue
	Comments []*Comment
}

type KeyValue struct {
	Range      Range
	Key        string
	KeyRange   Range // for hover on key
	Value      string
	ValueRange Range // for completion after =
}

type Comment struct {
	Range Range
	Text  string // includes # or ;
}
```

### Error Recovery Strategies (PARSE-06)

**LSP Context:** Parser must handle incomplete/invalid syntax (user is mid-typing)

**Strategies:**

1. **Panic-Mode Recovery:**
   - When error detected, skip tokens until synchronization point (next newline, section start)
   - Insert placeholder node with error flag
   - Continue parsing from sync point

2. **Graceful Degradation:**
   - Missing `]` in section header → treat rest of line as glob, mark error
   - Missing `=` in key-value → treat entire line as key with empty value, mark error
   - Unknown characters → insert error token, continue

3. **Partial Node Creation:**
   - `[incomplete` → create Section with Header="incomplete", mark as malformed
   - `key =` (no value) → create KeyValue with Value="", mark as incomplete
   - `= value` (no key) → create KeyValue with Key="", mark as malformed

4. **Error Node Tracking:**
```go
type ParseError struct {
	Range   Range
	Message string
	Code    string // e.g., "missing-section-close", "unexpected-equals"
}

type Document struct {
	// ... other fields
	Errors []ParseError // collected during parse
}
```

**Critical:** Parser should NEVER panic. Always return a Document (even if empty) + list of errors.

### Testing Strategy

**Test Categories:**

1. **Valid Syntax (PARSE-02, PARSE-03, PARSE-04, PARSE-05):**
   - Preamble with `root = true`
   - Sections with various glob patterns
   - Key-value pairs (all official properties)
   - Comments (both `#` and `;`)
   - Mixed whitespace (tabs, spaces, trailing)

2. **Position Accuracy (PARSE-01):**
   - Verify every node's Range matches source text
   - Multi-line files with varying line lengths
   - Unicode characters (affects byte offset vs char offset)

3. **Malformed Input (PARSE-06):**
   - Unclosed section headers: `[*.go`
   - Missing equals: `indent_style tab`
   - Empty lines, whitespace-only lines
   - Invalid UTF-8 sequences (should error gracefully)
   - Mid-typing scenarios: `[*.`, `indent_sty`, `indent_style =`

4. **Node Type Identification (PARSE-07):**
   - Correctly classify preamble vs section key-value pairs
   - Preserve comment vs content distinction
   - Handle edge cases (e.g., `root = true` inside section — valid parse, invalid semantics)

**Test File Examples:**

```
testdata/
├── valid/
│   ├── simple.editorconfig
│   ├── preamble.editorconfig
│   ├── multi-section.editorconfig
│   ├── complex-globs.editorconfig
│   └── comments.editorconfig
├── malformed/
│   ├── unclosed-section.editorconfig
│   ├── missing-equals.editorconfig
│   ├── empty-key.editorconfig
│   └── invalid-utf8.editorconfig
└── positions/
    ├── multiline.editorconfig
    ├── unicode.editorconfig
    └── mixed-whitespace.editorconfig
```

**Minimum Test Count:** 20+ test cases (per success criteria)

## Common Pitfalls

### 1. UTF-8 Byte Offset vs Character Position

**Problem:** Go's `range` over strings iterates runes, not bytes. LSP uses UTF-16 code units for positions (legacy from VS Code), but Go source is UTF-8 bytes.

**Solution:**
- Track byte offsets during lexing
- Provide conversion functions: `ByteOffsetToLineColumn`, `LineColumnToByteOffset`
- Use `utf8.RuneCount` or `utf16.Encode` for LSP position conversion (Phase 5)

**For Phase 1:** Use 0-indexed byte offsets + 1-indexed lines. Leave UTF-16 conversion for Phase 5 (LSP integration).

### 2. Line Ending Normalization

**Problem:** EditorConfig files can have LF or CRLF line endings. Positions must be consistent.

**Solution:**
- Lexer tracks original line endings (don't normalize during parse)
- Line breaks = `\n` or `\r\n` (treat `\r` alone as line break too per spec)
- Position.Line increments on any line break

### 3. Whitespace Handling

**Problem:** Spec says "trim leading/trailing whitespace" but internal whitespace is preserved.

**Solution:**
- Lexer emits tokens with original text (including whitespace)
- Parser trims when creating Key/Value strings for AST
- Store trimmed strings in AST, but preserve original Range for editor highlighting

### 4. Section Header Bracket Matching

**Problem:** Spec allows any characters between `[` and `]`, including nested brackets.

**Solution:**
- Lexer scans from `[` to first `]` on same line
- If no `]` found before newline → malformed section, emit error, continue
- Parser creates Section with Header = content between brackets (trimmed)

### 5. Key-Value Separator

**Problem:** `=` is required, but value can be empty (`key =` is valid, means value = "").

**Solution:**
- Lexer: After IDENTIFIER token, if next non-whitespace is `=`, emit EQUALS token
- Parser: KeyValue requires IDENTIFIER + EQUALS. If EQUALS has no following VALUE token → Value = ""
- If IDENTIFIER followed by newline (no `=`) → parse error, emit KeyValue with error flag

## Don't Hand-Roll

**Use stdlib where possible:**

- `bufio.Scanner` — line-by-line reading (but we need byte positions, so `bufio.Reader` + manual tracking)
- `unicode/utf8` — rune validation, counting
- `strings.TrimSpace` — whitespace trimming
- `testing` package with table-driven tests

**DO hand-roll:**
- Lexer (simple state machine, tracks positions)
- Parser (recursive descent or table-driven, creates AST)
- Position tracking (no stdlib support for this)

**Libraries to avoid:**
- `gopkg.in/ini.v1` — no position tracking
- Tree-sitter — decided against per requirements (cgo complexity, cross-compilation)
- Parser generators (yacc, antlr) — overkill for simple syntax, harder to customize error recovery

## Validation Architecture

*Note: This section is for Phase 2 planning. Phase 1 only parses.*

Phase 1 outputs an AST. Phase 2 will:
1. Walk the AST
2. Validate property keys (are they official EditorConfig properties?)
3. Validate property values (does `indent_style` = "tab" or "space"?)
4. Emit Diagnostics (LSP error/warning messages with Range)

**Interface for Phase 2:**
```go
// Phase 1 provides:
func Parse(source []byte) (*Document, error)

// Phase 2 will use:
func (doc *Document) Walk(visitor NodeVisitor)
func (kv *KeyValue) GetKey() (string, Range)
func (kv *KeyValue) GetValue() (string, Range)
```

## Summary

**Key Insights for Planning:**

1. **Lexer + Parser split** — cleanest position tracking, best error recovery
2. **Every AST node needs Range** — Position struct with Line (1-indexed), Column (0-indexed), Offset (bytes)
3. **Error recovery is critical** — LSP runs on incomplete code, parser must not panic
4. **Test malformed input heavily** — 50% of test cases should be invalid syntax
5. **Reuse editorconfig-core-go's glob logic** — don't reimplement FnmatchCase
6. **Phase 1 parses only** — validation happens in Phase 2 (separate concern)
7. **UTF-8 byte offsets** — leave UTF-16 conversion for Phase 5 (LSP protocol layer)

**Implementation Order (for planner):**

1. **Token types + Position struct** — foundation
2. **Lexer** — scan source to tokens with positions (handles PARSE-06 at token level)
3. **AST node types** — Document, Section, KeyValue, Comment with ranges
4. **Parser** — tokens to AST (handles PARSE-02, PARSE-03, PARSE-04, PARSE-05, PARSE-07)
5. **Error recovery** — parser continues after errors (PARSE-06 at parse level)
6. **Tests** — valid syntax, malformed input, position accuracy (20+ cases)

**Files to Create:**

```
internal/parser/
├── token.go         // TokenType, Token struct, Position, Range
├── lexer.go         // Lexer struct, Scan() method
├── lexer_test.go    // Token-level tests
├── ast.go           // AST node types (Document, Section, KeyValue, Comment)
├── parser.go        // Parser struct, Parse() method
├── parser_test.go   // Parse-level tests
└── testdata/        // Test .editorconfig files
    ├── valid/
    ├── malformed/
    └── positions/
```

**Phase 1 Success Criteria (for verification):**

- [ ] `Parse(source)` returns Document with all nodes having valid Range
- [ ] Lexer handles all EditorConfig syntax (comments, sections, key-values, preamble)
- [ ] Parser correctly identifies node types (preamble vs section keys)
- [ ] Parser handles malformed input without panicking (partial AST + errors)
- [ ] Tests cover 20+ cases (valid, malformed, position edge cases)
- [ ] All tests pass: `go test ./internal/parser/...`

---

*Research complete. Ready for phase planning.*
