# EditorConfig Language Server (ecfg)

## What This Is

A Go-based Language Server Protocol (LSP) implementation for EditorConfig files that provides hover documentation, autocomplete, diagnostics (linting), and validation for the official EditorConfig specification. It helps developers write correct `.editorconfig` files with intelligent editor support.

## Core Value

Developers can write `.editorconfig` files with confidence through real-time validation, contextual autocomplete, and inline documentation, preventing configuration errors before they reach production.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] LSP server provides hover tooltips with official spec documentation for EditorConfig properties
- [ ] LSP server provides context-aware autocomplete (property keys before `=`, enum values after `=`)
- [ ] LSP server validates property types (e.g., rejects `indent_size = large`)
- [ ] LSP server validates property positioning (e.g., `root = true` only in preamble)
- [ ] LSP server detects duplicate keys within sections
- [ ] LSP server warns about logical conflicts (e.g., `indent_style = tab` with numeric `indent_size`)
- [ ] LSP server resolves parent `.editorconfig` files for inheritance diagnostics
- [ ] LSP server identifies redundant properties inherited from parent files
- [ ] Custom Go parser generates AST with precise line/column position data
- [ ] Parser handles all EditorConfig syntax (comments, preamble, sections, key-value pairs, glob patterns)
- [ ] Server maintains virtual file system for unsaved document state
- [ ] Neovim integration via `lspconfig` and configuration documentation

### Out of Scope

- Tree-sitter parser integration — Custom Go parser is simpler, more maintainable, and sufficient for EditorConfig's simple syntax
- Non-standard EditorConfig properties — Only official spec properties from spec.editorconfig.org
- VS Code extension — Focus on Neovim/lspconfig as primary editor target
- Sublime Text, Emacs, IntelliJ — Desktop editor support limited to Neovim
- Web-based editor extensions — Focus on native editor LSP clients only
- Mobile editor support — Desktop editors only

## Context

**Problem:** Writing `.editorconfig` files is error-prone. Developers:
- Make typos in property names (`indent_sytle` vs `indent_style`)
- Use invalid values (`indent_size = large` instead of integer)
- Place `root = true` incorrectly (inside sections instead of preamble)
- Create duplicate keys without realizing
- Don't know which properties exist or what values are valid
- Can't easily discover parent file inheritance issues

**Solution:** An LSP server that brings IDE-quality intelligence to EditorConfig editing.

**Technical Environment:**
- Go programming language (for LSP implementation)
- EditorConfig specification from spec.editorconfig.org
- LSP protocol via `go.lsp.dev/protocol` or `github.com/tliron/glsp`
- EditorConfig core library: `github.com/editorconfig/editorconfig-core-go/v2` (for inheritance resolution)
- Custom parser (handwritten or PEG-based) instead of Tree-sitter

**Key Insight from Research:**
Custom Go parser preferred over Tree-sitter because:
- EditorConfig syntax is simple (comments, sections, key-value pairs)
- No cgo complexity, easy cross-compilation
- Full control over AST and position tracking
- Sufficient for LSP needs (hover, completion, diagnostics)
- Easier maintenance and contribution

## Constraints

- **Tech Stack**: Go for LSP server (pure Go, no cgo) — Required for cross-platform compatibility and easy distribution
- **Spec Compliance**: Must follow official EditorConfig spec at spec.editorconfig.org — Non-negotiable for correctness
- **Editor Support**: Neovim via lspconfig — Focus on single LSP-capable editor for simplicity
- **Performance**: Full document reparse on change is acceptable — EditorConfig files are small (<100 lines typically)
- **Distribution**: Single binary distribution via GoReleaser — Easy installation for users

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Custom Go parser instead of Tree-sitter | EditorConfig syntax is simple; Tree-sitter adds cgo complexity, cross-compile difficulties, and single-maintainer grammar dependency. Custom parser gives full control over AST/positions without overhead. | ✓ Good |
| Use editorconfig-core-go for inheritance | Official library handles file system traversal and parent resolution correctly. Parser focuses on single-file AST, core library handles semantics. | — Pending |
| Pure Go implementation (no cgo) | Enables trivial cross-compilation, smaller binaries, easier CI/CD, simpler contributor onboarding. | — Pending |
| LSP protocol via go.lsp.dev/protocol | Standard, well-maintained Go LSP library with good community support. | — Pending |
| Neovim as primary editor target | Simplified scope; focus on lspconfig integration for excellent LSP experience in Neovim. | ✓ Good |

---
*Last updated: 2026-03-18 after initialization*
