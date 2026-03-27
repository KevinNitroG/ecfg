---
phase: 05-editor-integration
status: Ready for planning
---

<domain>
## Phase Boundary

Phase 5 delivers the complete LSP server binary and Neovim integration documentation. This packages all prior phases (parser, validation, LSP intelligence, file resolution) into a runnable server that Neovim users can configure via lspconfig.
</domain>

<decisions>
## Implementation Decisions

### LSP Server Architecture
- Use standard LSP library (lspzip or jsonrpc) for protocol handling
- Implement Initialize, Shutdown, and window/showMessage handlers
- Single-server instance with document-based state
- Capabilities: textDocumentSync (full), hover, completion

### Virtual File System
- In-memory map[uri]document for open buffers
- No file watching (rely on didChange/didOpen/didClose)
- Parse on didChange, cache AST until next change

### Neovim Integration
- Document in README.md with lspconfig setup example
- Provide install/build instructions
- No custom Neovim plugin - use existing lspconfig

### Cross-Platform Build
- Standard `go build` for each platform
- Output: ecfg-lsp (no platform suffix)
- Provide Makefile with build targets for linux/darwin/windows

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- internal/parser: AST types and Parse() function
- internal/validator: Schema and Validate() function
- internal/lsp: hover.go, completion.go handlers
- internal/resolver: Resolve() and InheritedProperties()

### Established Patterns
- go.mod at project root
- internal/ packages for core logic
- json-based LSP types via official lsp/protocol

### Integration Points
- LSP server entry point: cmd/ecfg-lsp/main.go
- Handler registration at server init
- Document lifecycle methods bridge to parser/validator

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches
</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope
</deferred>