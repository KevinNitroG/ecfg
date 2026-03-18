# Functional Specification: EditorConfig Language Server & Linter (Go)

## 1. Feature Specifications

### 1.1 Core Standard Settings (Official Spec)

The server will enforce and provide intelligence for the official standard properties defined at [spec.editorconfig.org](https://spec.editorconfig.org/):

- `root`: (Values: `true`) Must be at the top of the file (preamble), outside of any section.
- `indent_style`: (Values: `tab`, `space`).
- `indent_size`: (Values: `1`-`8`, `tab`).
- `tab_width`: (Values: Positive integer). Defaults to `indent_size`.
- `end_of_line`: (Values: `lf`, `crlf`, `cr`).
- `charset`: (Values: `utf-8`, `utf-8-bom`, `utf-16be`, `utf-16le`, `latin1`).
- `trim_trailing_whitespace`: (Values: `true`, `false`).
- `insert_final_newline`: (Values: `true`, `false`).

### 1.2 Language Server Capabilities

- Hover Explanations: When the cursor hovers over a property (e.g., `trim_trailing_whitespace`), the LSP returns a Markdown tooltip pulling directly from the official spec definitions.
- Autocomplete (IntelliSense):
  - _Context-aware:_ If typing before an `=`, suggest standard keys. If typing after `=`, suggest the constrained enum values (e.g., suggesting `lf`, `cr`, `crlf` for `end_of_line`).
- Diagnostics (Linting & Warnings):
  - _Type Validation:_ Flagging `indent_size = large` as an error (expects integer or `tab`).
  - _Positional Errors:_ Emitting an error if `root = true` is placed inside a `[*.go]` section instead of the top-level preamble.
  - _Mutually Exclusive / Conflict Diagnostics:_
    - Warning if a section contains duplicate keys (e.g., defining `charset` twice in the same section).
    - Warning if `indent_style = tab` but `indent_size` is set to a numeric value (while technically allowed, it is often a logical conflict; `tab_width` is the correct property here).

### 1.3 Global & Parent Resolution

EditorConfig relies heavily on directory hierarchies. The LSP should contextualize the current file by looking at its parents:

- Inheritance Diagnostics: If `[*.js]` in `src/.editorconfig` sets `indent_size = 4`, but the parent `/project/.editorconfig` already enforces `indent_size = 4` for `[*.js]`, the LSP can provide an "Info" diagnostic: _Redundant property: inherited from parent .editorconfig_.
- Root Tracing: The server will traverse the file system upwards until it hits `root = true` or the system root, allowing the LSP to simulate the exact final configuration of any given file.

---

## 2. The Parser Architecture (Why standard `go-ini` is not enough)

Standard Go INI packages (like `gopkg.in/ini.v1`) are built for _configuration reading/writing_, not language intelligence.

The Limitation of `go-ini`: Standard parsers map text into Go `structs` or `maps`. They discard token boundaries, whitespace, and crucially, Line and Column coordinate data. For an LSP to provide autocomplete or hover data, it must know exactly what syntax node exists at `Line 5, Character 12`. A standard INI parser cannot answer this.

The Solution: [Tree-Sitter](https://github.com/ValdezFOmar/tree-sitter-editorconfig) (No manual recursive parser needed)
Instead of writing a custom recursive descent parser, you should use Tree-Sitter. This is the industry standard for modern LSPs (used natively by Neovim, Helix, and GitHub).

- How it works: Tree-Sitter generates an error-resilient Concrete Syntax Tree (CST). Even if the user is in the middle of typing an invalid line, Tree-Sitter parses the rest of the file successfully.
- Positional Data: Every single node (Key, Value, Section, Comment) automatically comes with `.StartPoint()` and `.EndPoint()` (containing precise Line and Byte/Column offsets), mapping perfectly to the LSP `Range` object.

---

## 3. Implementation Roadmap

### Phase 1: Core Stack Setup

1.  LSP Framework: Import `go.lsp.dev/protocol` or `github.com/tliron/glsp` to handle the JSON-RPC communication lifecycle (`Initialize`, `textDocument/didOpen`, etc.).
2.  Parser Integration: \* Import `github.com/smacker/go-tree-sitter`.
    - Bind the C-based `tree-sitter-editorconfig` or `tree-sitter-ini` grammar to your Go application.
3.  Core Evaluator: Import `github.com/editorconfig/editorconfig-core-go/v2`. You will use Tree-Sitter to parse the _current_ unsaved document for LSP features, but you will use the official core library to resolve the _parent_ files and global inheritance.

### Phase 2: State Management (Virtual File System)

1.  Implement handlers for `textDocument/didOpen`, `didChange`, and `didClose`.
2.  Maintain an in-memory map of the file contents. When `didChange` fires, update the memory buffer and re-trigger the Tree-Sitter parse to generate a fresh CST.

### Phase 3: The Intelligence Engine (Linter & LSP Features)

1.  Schema Map: Create a Go dictionary mapping standard keys to their expected types, enums, and Markdown documentation strings.
2.  Hover Endpoint: On `textDocument/hover`, take the cursor coordinates, ask Tree-Sitter for the node at that coordinate. If it's a "Key" node, return the matching documentation from the Schema Map.
3.  Completion Endpoint: On `textDocument/completion`, check the node immediately preceding the cursor. Provide completions based on whether the context expects a Section, a Key, or a Value.
4.  Diagnostics Endpoint: Traverse the Tree-Sitter CST upon every file change.
    - Check for duplicates.
    - Validate values against the Schema Map.
    - Push an array of `Diagnostic` objects back to the client editor.

### Phase 4: Editor Extensions

1.  Build a lightweight VS Code extension (a basic `package.json` and a few lines of TypeScript) that simply spawns your compiled Go binary and pipes standard I/O.
2.  Provide configuration instructions for Neovim's `lspconfig` to attach to your binary for `editorconfig` filetypes.

---

## 4. Consider

- Use git submodule to download https://github.com/ValdezFOmar/tree-sitter-editorconfig into go vendor. Set up Renovate to auto update, and setup [goreleaser](https://goreleaser.com/) with GitHub action, using [release please](https://github.com/googleapis/release-please)
