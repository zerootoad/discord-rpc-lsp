package main

import (
	"context"
	"log"
	"os"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/zerootoad/discord-rpc-lsp/client"
	"github.com/zerootoad/discord-rpc-lsp/handler"
)

func main() {
	logFile, err := os.OpenFile("discord-lsp.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	err = client.Login("1350478888027029596")
	if err != nil {
		log.Fatalf("Failed to create Discord RPC client: %v", err)
	}
	defer client.Logout()

	lspHandler := handler.NewLSPHandler()

	stream := jsonrpc2.NewBufferedStream(stdrwc{}, jsonrpc2.VSCodeObjectCodec{})
	conn := jsonrpc2.NewConn(context.Background(), stream, lspHandler)

	log.Println("LSP server started!")
	<-conn.DisconnectNotify()
	log.Println("LSP server stopped!")
}

// stdin n stdout bs
type stdrwc struct{}

func (stdrwc) Read(p []byte) (int, error) {
	return os.Stdin.Read(p)
}

func (stdrwc) Write(p []byte) (int, error) {
	return os.Stdout.Write(p)
}

func (stdrwc) Close() error {
	return nil
}
