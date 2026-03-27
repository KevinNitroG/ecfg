package main

import (
	"context"
	"io"
	"os"

	"github.com/KevinNitroG/ecfg/internal/lsp"
	"go.lsp.dev/jsonrpc2"
)

func main() {
	// Create the LSP server
	server := lsp.NewServer()

	// Create a stream from stdin/stdout for stdio communication
	stream := jsonrpc2.NewStream(struct{ io.ReadWriteCloser }{os.Stdin})

	// Create a connection using jsonrpc2
	conn := jsonrpc2.NewConn(stream)

	// Run the server - attach the LSP server handler
	conn.Go(context.Background(), server.ServerHandler())

	// Wait for the connection to close
	<-conn.Done()
}
