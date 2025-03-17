package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/zerootoad/discord-rpc-lsp/client"
	"github.com/zerootoad/discord-rpc-lsp/handler"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

func main() {
	homedir := utils.GetUserHomeDir()
	configDir := filepath.Join(homedir, ".discord-rpc-lsp")

	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		log.Fatalf("Failed to create config directory '%s': %v", configDir, err)
	}

	logFilePath := filepath.Join(configDir, "log.txt")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file '%s': %v", logFilePath, err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	err = client.Login("1350478888027029596")
	if err != nil {
		log.Fatalf("Failed to create Discord RPC client: %v", err)
	}
	defer client.Logout()

	lspHandler, err := handler.NewLSPHandler()
	if err != nil {
		log.Fatalf("Failed to create LSP handler: %v", err)
	}

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
