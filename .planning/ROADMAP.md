# Roadmap: EditorConfig Language Server (ecfg)

**Created:** 2026-03-18  
**Granularity:** Coarse (3-5 phases)  
**Coverage:** 41/41 requirements mapped ✓

## Overview

| Phase | Goal | Requirements | Status |
|-------|------|--------------|--------|
| 1 | Core Parser & AST | 7 | Pending |
| 2 | Schema Validation & Diagnostics | 15 | Pending |
| 3 | LSP Intelligence Features | 9 | Pending |
| 4 | File System Resolution | 5 | Pending |
| 5 | Editor Integration | 5 | Pending |

**Total: 5 phases covering 41 requirements**

---

## Phase 1: Core Parser & AST

**Goal:** Parse `.editorconfig` files into AST with precise position tracking for LSP features.

**Why this first:** LSP features (hover, completion, diagnostics) require AST with exact line/column positions. Everything depends on the parser.

**Requirements:**
- PARSE-01: Parser generates AST with precise line/column position data
- PARSE-02: Parser recognizes preamble key-value pairs
- PARSE-03: Parser recognizes section headers with glob patterns
- PARSE-04: Parser recognizes key-value pairs within sections
- PARSE-05: Parser recognizes and preserves comments
- PARSE-06: Parser handles malformed input gracefully
- PARSE-07: Parser provides node type identification

**Success Criteria:**
1. Parser correctly identifies all node types (preamble, section, key, value, comment) in test files
2. Every AST node includes start/end line and column positions
3. Parser recovers from syntax errors without crashing (e.g., mid-typing scenarios)
4. Test suite covers valid and invalid EditorConfig syntax

**Deliverables:**
- `internal/parser/` package with lexer and parser
- AST node types with position data
- Test suite with 20+ test cases

---

## Phase 2: Schema Validation & Diagnostics

**Goal:** Validate EditorConfig properties against official spec and emit diagnostics for errors/warnings.

**Why this second:** With AST available, implement validation rules. This provides immediate value (linting) before LSP server setup.

**Requirements:**
- SCHEMA-01: Validates all official EditorConfig properties
- SCHEMA-02: Validates `root` only in preamble
- SCHEMA-03: Validates `indent_style` enum values
- SCHEMA-04: Validates `indent_size` integer or `tab`
- SCHEMA-05: Validates `tab_width` positive integer
- SCHEMA-06: Validates `end_of_line` enum values
- SCHEMA-07: Validates `charset` enum values
- SCHEMA-08: Validates `trim_trailing_whitespace` boolean
- SCHEMA-09: Validates `insert_final_newline` boolean
- DIAG-01: Emits error for invalid property values
- DIAG-02: Emits error for misplaced `root`
- DIAG-03: Emits warning for duplicate keys
- DIAG-04: Emits warning for logical conflicts
- DIAG-05: Emits info for redundant inherited properties
- DIAG-06: Diagnostics include precise Range

**Success Criteria:**
1. Validator detects all invalid property values from test suite
2. Validator correctly identifies positional errors (`root` in section)
3. Validator warns about duplicate keys in same section
4. Validator warns about `indent_style=tab` + numeric `indent_size` conflict
5. Each diagnostic includes precise line/column Range for editor underline

**Deliverables:**
- `internal/validator/` package with schema rules
- `internal/diagnostic/` package for LSP diagnostic generation
- Schema map with official property definitions
- Test suite for all validation rules

---

## Phase 3: LSP Intelligence Features

**Goal:** Implement hover documentation and context-aware autocomplete for EditorConfig properties.

**Why this third:** With validation working, add intelligence features that make editing easier (hover + completion).

**Requirements:**
- HOVER-01: Provides Markdown hover tooltip for property keys
- HOVER-02: Hover includes official spec description
- HOVER-03: Hover includes valid values for property
- HOVER-04: Hover works when cursor on key name
- COMP-01: Completion suggestions for property keys before `=`
- COMP-02: Completion suggestions for enum values after `=`
- COMP-03: Context-aware completion (no `root` in sections)
- COMP-04: Completion items include documentation
- COMP-05: Completion suggests only valid values for property

**Success Criteria:**
1. Hovering over `indent_style` shows spec description and valid values (`tab`, `space`)
2. Typing before `=` suggests all valid property keys for context
3. Typing after `=` for `end_of_line` suggests `lf`, `crlf`, `cr` only
4. Completion does not suggest `root` when cursor inside a section
5. All completion items include brief documentation snippets

**Deliverables:**
- `internal/lsp/hover.go` implementing `textDocument/hover`
- `internal/lsp/completion.go` implementing `textDocument/completion`
- Schema map extended with Markdown documentation strings
- Test suite for hover and completion edge cases

---

## Phase 4: File System Resolution

**Goal:** Resolve parent `.editorconfig` files for inheritance diagnostics.

**Why this fourth:** Inheritance detection requires parser + validator working. Adds advanced diagnostics (redundant properties).

**Requirements:**
- FS-01: Traverses parent directories to find parent files
- FS-02: Stops traversal at `root = true`
- FS-03: Stops traversal at system root
- FS-04: Uses editorconfig-core-go for resolution
- FS-05: Detects property inheritance for diagnostics

**Success Criteria:**
1. Server correctly finds parent `.editorconfig` files up directory tree
2. Server stops at `root = true` in parent file
3. Server stops at filesystem root if no `root = true` found
4. Server emits info diagnostic when property is redundant (inherited from parent)
5. Integration with editorconfig-core-go library works correctly

**Deliverables:**
- `internal/resolver/` package wrapping editorconfig-core-go
- Inheritance detection logic
- Test suite with multi-level `.editorconfig` hierarchies

---

## Phase 5: Editor Integration

**Goal:** Package LSP server and create editor extensions for VS Code and Neovim.

**Why this last:** LSP features must work before editor integration. This phase packages everything for end users.

**Requirements:**
- LSP-01: Handles `initialize` request
- LSP-02: Handles `textDocument/didOpen`
- LSP-03: Handles `textDocument/didChange`
- LSP-04: Handles `textDocument/didClose`
- LSP-05: Maintains in-memory virtual file system
- LSP-06: Triggers reparse on `didChange`
- LSP-07: Handles `textDocument/hover`
- LSP-08: Handles `textDocument/completion`
- LSP-09: Handles `textDocument/publishDiagnostics`
- EDITOR-01: VS Code extension spawns binary via stdio
- EDITOR-02: VS Code extension activates on `.editorconfig`
- EDITOR-03: VS Code extension passes LSP messages
- EDITOR-04: Neovim integration documented
- EDITOR-05: Server binary is cross-platform

**Success Criteria:**
1. LSP server responds to `initialize` with correct capabilities
2. Server maintains virtual file system for unsaved changes
3. VS Code extension activates on opening `.editorconfig` file
4. VS Code extension shows hover, completion, diagnostics in editor
5. Neovim users can configure server via `lspconfig` following README

**Deliverables:**
- `cmd/ecfg-lsp/` binary with full LSP lifecycle
- `internal/lsp/server.go` implementing LSP protocol
- VS Code extension in `editors/vscode/`
- Neovim setup documentation in README
- Cross-platform build configuration (Linux, macOS, Windows)

---

## Dependencies

```
Phase 1 (Parser)
    ↓
Phase 2 (Validation) ← requires AST
    ↓
Phase 3 (Intelligence) ← requires validation
    ↓
Phase 4 (Resolution) ← requires validation
    ↓
Phase 5 (Integration) ← requires all features
```

**Critical path:** 1 → 2 → 3 → 5 (minimum viable LSP)  
**Parallel opportunity:** Phase 4 can be developed alongside Phase 3

---

## Validation

✓ All 41 v1 requirements mapped  
✓ No unmapped requirements  
✓ Each phase has clear deliverables  
✓ Success criteria are observable/testable  
✓ Dependencies follow logical build order

---

*Roadmap created: 2026-03-18*
