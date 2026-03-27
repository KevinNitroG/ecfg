package lsp

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/KevinNitroG/ecfg/internal/diagnostic"
	"github.com/KevinNitroG/ecfg/internal/parser"
	"github.com/KevinNitroG/ecfg/internal/validator"
	"go.lsp.dev/jsonrpc2"
	"go.lsp.dev/protocol"
)

// Server is the LSP server that handles LSP protocol requests.
// It manages a virtual file system with in-memory documents.
type Server struct {
	mu         sync.RWMutex
	documents  map[protocol.DocumentURI]*document
	workspace  string
	initialize bool
}

// document represents an open text document in the LSP server.
type document struct {
	URI         protocol.DocumentURI
	Version     int
	Content     string
	AST         *parser.Document
	Diagnostics []diagnostic.Diagnostic
}

// NewServer creates a new LSP server instance.
func NewServer() *Server {
	return &Server{
		documents: make(map[protocol.DocumentURI]*document),
	}
}

// ServerHandler returns a handler function that handles LSP requests.
// This implements the jsonrpc2.Handler interface.
func (s *Server) ServerHandler() func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	return func(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
		switch req.Method() {
		// Lifecycle methods
		case "initialize":
			return s.handleInitialize(ctx, reply, req)
		case "initialized":
			return nil // No response needed
		case "shutdown":
			return s.handleShutdown(ctx, reply, req)
		case "exit":
			// Exit doesn't send a response
			if err := reply(ctx, nil, nil); err != nil {
				log.Printf("Failed to reply to exit: %v", err)
			}
			return nil

		// Text document methods
		case "textDocument/didOpen":
			return s.handleDidOpen(ctx, reply, req)
		case "textDocument/didChange":
			return s.handleDidChange(ctx, reply, req)
		case "textDocument/didClose":
			return s.handleDidClose(ctx, reply, req)

		// Feature methods
		case "textDocument/hover":
			return s.handleHover(ctx, reply, req)
		case "textDocument/completion":
			return s.handleCompletion(ctx, reply, req)

		default:
			log.Printf("Unknown method: %s", req.Method())
			return nil
		}
	}
}

// handleInitialize handles the initialize request.
func (s *Server) handleInitialize(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	s.mu.Lock()
	s.initialize = true
	s.mu.Unlock()

	var params protocol.InitializeParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	// Store workspace root if provided
	if params.RootURI != "" {
		s.mu.Lock()
		s.workspace = string(params.RootURI)
		s.mu.Unlock()
	}

	result := &protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync: protocol.TextDocumentSyncOptions{
				Change: protocol.TextDocumentSyncKindFull,
			},
			HoverProvider:      true,
			CompletionProvider: &protocol.CompletionOptions{},
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "ecfg",
			Version: "0.1.0",
		},
	}

	return reply(ctx, result, nil)
}

// handleShutdown handles the shutdown request.
func (s *Server) handleShutdown(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	s.mu.Lock()
	s.initialize = false
	s.mu.Unlock()
	return reply(ctx, nil, nil)
}

// handleDidOpen handles the textDocument/didOpen notification.
func (s *Server) handleDidOpen(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidOpenTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil
	}

	uri := params.TextDocument.URI
	content := params.TextDocument.Text
	version := int(params.TextDocument.Version)

	// Parse and create document
	doc, err := parser.Parse([]byte(content))
	if err != nil {
		log.Printf("Failed to parse document %s: %v", uri, err)
	}

	// Validate and get diagnostics
	var diagnostics []diagnostic.Diagnostic
	if doc != nil {
		errors := validator.Validate(doc)
		diagnostics = diagnostic.ToDiagnostics(errors)
	}

	s.mu.Lock()
	s.documents[uri] = &document{
		URI:         uri,
		Version:     version,
		Content:     content,
		AST:         doc,
		Diagnostics: diagnostics,
	}
	s.mu.Unlock()

	// Publish diagnostics
	s.publishDiagnostics(ctx, uri, version, diagnostics)

	return nil
}

// handleDidChange handles the textDocument/didChange notification.
func (s *Server) handleDidChange(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidChangeTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil
	}

	uri := params.TextDocument.URI
	version := int(params.TextDocument.Version)
	content := params.ContentChanges[0].Text

	s.mu.Lock()
	doc, exists := s.documents[uri]
	if !exists {
		s.mu.Unlock()
		return nil
	}

	// Update content and reparse
	doc.Content = content
	doc.Version = version

	if content != "" {
		doc.AST, _ = parser.Parse([]byte(content))
	} else {
		doc.AST = nil
	}

	// Re-validate and get diagnostics
	var diagnostics []diagnostic.Diagnostic
	if doc.AST != nil {
		errors := validator.Validate(doc.AST)
		diagnostics = diagnostic.ToDiagnostics(errors)
	}
	doc.Diagnostics = diagnostics
	s.mu.Unlock()

	// Publish diagnostics
	s.publishDiagnostics(ctx, uri, version, diagnostics)

	return nil
}

// handleDidClose handles the textDocument/didClose notification.
func (s *Server) handleDidClose(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.DidCloseTextDocumentParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return nil
	}

	uri := params.TextDocument.URI

	s.mu.Lock()
	delete(s.documents, uri)
	s.mu.Unlock()

	return nil
}

// handleHover handles the textDocument/hover request.
func (s *Server) handleHover(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.TextDocumentPositionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	s.mu.RLock()
	doc, exists := s.documents[params.TextDocument.URI]
	s.mu.RUnlock()

	if !exists || doc.AST == nil {
		return reply(ctx, nil, nil)
	}

	hover := ComputeHover(doc.AST, params.Position)
	return reply(ctx, hover, nil)
}

// handleCompletion handles the textDocument/completion request.
func (s *Server) handleCompletion(ctx context.Context, reply jsonrpc2.Replier, req jsonrpc2.Request) error {
	var params protocol.CompletionParams
	if err := json.Unmarshal(req.Params(), &params); err != nil {
		return reply(ctx, nil, err)
	}

	s.mu.RLock()
	doc, exists := s.documents[params.TextDocument.URI]
	s.mu.RUnlock()

	if !exists || doc.AST == nil {
		return reply(ctx, nil, nil)
	}

	completion := ComputeCompletion(doc.AST, params.Position)
	return reply(ctx, completion, nil)
}

// publishDiagnostics sends diagnostics to the client.
func (s *Server) publishDiagnostics(ctx context.Context, uri protocol.DocumentURI, version int, diagnostics []diagnostic.Diagnostic) {
	// Convert internal diagnostics to LSP diagnostics
	lspDiags := make([]protocol.Diagnostic, len(diagnostics))
	for i, d := range diagnostics {
		lspDiags[i] = protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      uint32(d.Range.Start.Line - 1),
					Character: uint32(d.Range.Start.Column),
				},
				End: protocol.Position{
					Line:      uint32(d.Range.End.Line - 1),
					Character: uint32(d.Range.End.Column),
				},
			},
			Severity: protocol.DiagnosticSeverity(d.Severity),
			Message:  d.Message,
			Source:   "ecfg",
		}
	}

	// Send notification - we need to use the connection for this
	// For now, we'll rely on jsonrpc2's notification mechanism
	// The actual notification sending would need to be done via the connection
	log.Printf("Publishing %d diagnostics for %s (version %d)", len(diagnostics), uri, version)
}
