# Requirements: EditorConfig Language Server (ecfg)

**Defined:** 2026-03-18
**Core Value:** Developers can write `.editorconfig` files with confidence through real-time validation, contextual autocomplete, and inline documentation, preventing configuration errors before they reach production.

## v1 Requirements

### Parser (PARSE)

- [x] **PARSE-01**: Parser generates AST with precise line/column position data for all nodes
- [x] **PARSE-02**: Parser recognizes preamble key-value pairs (before any section)
- [x] **PARSE-03**: Parser recognizes section headers with glob patterns `[*.ext]`
- [x] **PARSE-04**: Parser recognizes key-value pairs within sections
- [x] **PARSE-05**: Parser recognizes and preserves comments
- [x] **PARSE-06**: Parser handles malformed input gracefully (error recovery for LSP mid-typing scenarios)
- [x] **PARSE-07**: Parser provides node type identification (preamble, section, key, value, comment)

### Schema Validation (SCHEMA)

- [x] **SCHEMA-01**: Server validates official EditorConfig properties: `root`, `indent_style`, `indent_size`, `tab_width`, `end_of_line`, `charset`, `trim_trailing_whitespace`, `insert_final_newline`
- [x] **SCHEMA-02**: Server validates `root` property appears only in preamble (not inside sections)
- [x] **SCHEMA-03**: Server validates `indent_style` accepts only `tab` or `space`
- [x] **SCHEMA-04**: Server validates `indent_size` accepts integers 1-8 or `tab`
- [x] **SCHEMA-05**: Server validates `tab_width` accepts positive integers
- [x] **SCHEMA-06**: Server validates `end_of_line` accepts only `lf`, `crlf`, or `cr`
- [x] **SCHEMA-07**: Server validates `charset` accepts only `utf-8`, `utf-8-bom`, `utf-16be`, `utf-16le`, `latin1`
- [x] **SCHEMA-08**: Server validates `trim_trailing_whitespace` accepts only `true` or `false`
- [x] **SCHEMA-09**: Server validates `insert_final_newline` accepts only `true` or `false`

### Diagnostics (DIAG)

- [x] **DIAG-01**: Server emits error diagnostic for invalid property values (e.g., `indent_size = large`)
- [x] **DIAG-02**: Server emits error diagnostic for `root = true` placed inside section
- [x] **DIAG-03**: Server emits warning diagnostic for duplicate keys within same section
- [x] **DIAG-04**: Server emits warning diagnostic for logical conflicts (`indent_style = tab` with numeric `indent_size`)
- [x] **DIAG-05**: Server emits info diagnostic for redundant properties inherited from parent `.editorconfig`
- [x] **DIAG-06**: Diagnostics include precise Range (line/column start/end) for underline in editor

### Hover (HOVER)

- [x] **HOVER-01**: Server provides Markdown hover tooltip for property keys
- [x] **HOVER-02**: Hover content includes official spec description from spec.editorconfig.org
- [x] **HOVER-03**: Hover content includes valid values for the property
- [x] **HOVER-04**: Hover works when cursor is on key name (before `=`)

### Completion (COMP)

- [x] **COMP-01**: Server provides completion suggestions for property keys when cursor is before `=`
- [x] **COMP-02**: Server provides completion suggestions for enum values when cursor is after `=`
- [x] **COMP-03**: Completion suggestions are context-aware (no `root` suggestions inside sections)
- [x] **COMP-04**: Completion items include documentation snippets
- [x] **COMP-05**: Completion suggests only valid values for the property being edited

### File System Resolution (FS)

- [x] **FS-01**: Server traverses parent directories to find parent `.editorconfig` files
- [x] **FS-02**: Server stops traversal when `root = true` is found in a parent file
- [x] **FS-03**: Server stops traversal at system root if no `root = true` found
- [x] **FS-04**: Server uses editorconfig-core-go library for correct resolution semantics
- [x] **FS-05**: Server detects property inheritance from parent files for diagnostics

### LSP Lifecycle (LSP)

- [x] **LSP-01**: Server handles `initialize` request and responds with capabilities
- [x] **LSP-02**: Server handles `textDocument/didOpen` to register new documents
- [x] **LSP-03**: Server handles `textDocument/didChange` to update document state
- [x] **LSP-04**: Server handles `textDocument/didClose` to clean up document state
- [x] **LSP-05**: Server maintains in-memory virtual file system for unsaved changes
- [x] **LSP-06**: Server triggers full document reparse on `didChange` events
- [x] **LSP-07**: Server handles `textDocument/hover` requests with cursor position
- [x] **LSP-08**: Server handles `textDocument/completion` requests with cursor position
- [x] **LSP-09**: Server handles `textDocument/publishDiagnostics` to send errors/warnings to client

### Editor Integration (EDITOR)

- [x] **EDITOR-01**: Neovim integration documented in README with `lspconfig` configuration example
- [x] **EDITOR-02**: Server binary is cross-platform (Linux, macOS, Windows)

## v2 Requirements

### Advanced Features

- **ADV-01**: Code actions to fix common issues (e.g., "Move root to preamble")
- **ADV-02**: Formatting support for `.editorconfig` files (consistent spacing, alignment)
- **ADV-03**: Workspace symbol search across all `.editorconfig` files in project
- **ADV-04**: Reference provider (find all sections matching a file path)
- **ADV-05**: Signature help for glob pattern syntax in section headers

### Tooling & Distribution

- **TOOL-01**: CLI tool to validate `.editorconfig` files without editor
- **TOOL-02**: GitHub Action for CI validation of `.editorconfig` files
- **TOOL-03**: Pre-commit hook integration
- **TOOL-04**: Homebrew formula for easy installation
- **TOOL-05**: GoReleaser setup with GitHub Actions for automated releases
- **TOOL-06**: Renovate configuration for dependency updates

### Extended Editor Support

- **ED-01**: Emacs integration via `lsp-mode`
- **ED-02**: VS Code extension

## Out of Scope

| Feature | Reason |
|---------|--------|
| Tree-sitter parser | Custom Go parser is simpler, no cgo complexity, easier cross-compilation. EditorConfig syntax is simple enough. |
| Non-standard properties | Only official spec from spec.editorconfig.org. Custom properties vary by editor and create maintenance burden. |
| Real-time performance optimization | EditorConfig files are typically <100 lines. Full reparse on change is fast enough. |
| Web-based editor extensions | Focus on native LSP clients (VS Code, Neovim). Web editors have different integration patterns. |
| Incremental parsing | Not needed for small files. Full reparse is simpler and sufficient. |
| Configuration file for LSP server | Zero-config by default. Server should work out of the box. |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| PARSE-01 | Phase 1 | Complete |
| PARSE-02 | Phase 1 | Complete |
| PARSE-03 | Phase 1 | Complete |
| PARSE-04 | Phase 1 | Complete |
| PARSE-05 | Phase 1 | Complete |
| PARSE-06 | Phase 1 | Complete |
| PARSE-07 | Phase 1 | Complete |
| SCHEMA-01 | Phase 2 | Complete |
| SCHEMA-02 | Phase 2 | Complete |
| SCHEMA-03 | Phase 2 | Complete |
| SCHEMA-04 | Phase 2 | Complete |
| SCHEMA-05 | Phase 2 | Complete |
| SCHEMA-06 | Phase 2 | Complete |
| SCHEMA-07 | Phase 2 | Complete |
| SCHEMA-08 | Phase 2 | Complete |
| SCHEMA-09 | Phase 2 | Complete |
| DIAG-01 | Phase 2 | Complete |
| DIAG-02 | Phase 2 | Complete |
| DIAG-03 | Phase 2 | Complete |
| DIAG-04 | Phase 2 | Complete |
| DIAG-05 | Phase 2 | Complete |
| DIAG-06 | Phase 2 | Complete |
| HOVER-01 | Phase 3 | Complete |
| HOVER-02 | Phase 3 | Complete |
| HOVER-03 | Phase 3 | Complete |
| HOVER-04 | Phase 3 | Complete |
| COMP-01 | Phase 3 | Complete |
| COMP-02 | Phase 3 | Complete |
| COMP-03 | Phase 3 | Complete |
| COMP-04 | Phase 3 | Complete |
| COMP-05 | Phase 3 | Complete |
| FS-01 | Phase 4 | Complete |
| FS-02 | Phase 4 | Complete |
| FS-03 | Phase 4 | Complete |
| FS-04 | Phase 4 | Complete |
| FS-05 | Phase 4 | Complete |
| LSP-01 | Phase 5 | Complete |
| LSP-02 | Phase 5 | Complete |
| LSP-03 | Phase 5 | Complete |
| LSP-04 | Phase 5 | Complete |
| LSP-05 | Phase 5 | Complete |
| LSP-06 | Phase 5 | Complete |
| LSP-07 | Phase 5 | Complete |
| LSP-08 | Phase 5 | Complete |
| LSP-09 | Phase 5 | Complete |
| EDITOR-01 | Phase 5 | Complete |
| EDITOR-02 | Phase 5 | Complete |

**Coverage:**
- v1 requirements: 36 total
- Mapped to phases: 36
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-18*
*Last updated: 2026-03-18 after initial definition*
