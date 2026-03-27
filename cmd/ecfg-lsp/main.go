package main

import (
	"context"
	"io"
	"log"
	"os"

	"github.com/KevinNitroG/ecfg/internal/lsp"
	"go.lsp.dev/jsonrpc2"
)

type readWriteCloser struct {
	io.Reader
	io.Writer
	io.Closer
}

func (r readWriteCloser) Close() error {
	return nil
}

func main() {
	log.SetFlags(log.Lshortfile)
	log.Println("Starting ecfg-lsp server...")

	// Create the LSP server
	server := lsp.NewServer()

	log.Println("Server created, setting up stdio...")

	// Create a combined stdio reader/writer for LSP communication
	// stdin for reading, stdout for writing
	rwc := readWriteCloser{Reader: os.Stdin, Writer: os.Stdout}
	stream := jsonrpc2.NewStream(rwc)

	log.Println("Stream created, creating connection...")

	// Create a connection using jsonrpc2
	conn := jsonrpc2.NewConn(stream)

	// Pass the connection to the server so it can send notifications
	server.SetConn(conn)

	log.Println("Connection ready, starting handler...")

	// Run the server - attach the LSP server handler
	conn.Go(context.Background(), server.ServerHandler())

	// Wait for the connection to close
	<-conn.Done()

	log.Println("Connection closed, exiting")
}
