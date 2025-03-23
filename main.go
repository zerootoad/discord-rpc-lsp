package main

import (
	"log"
	"os"
	"path/filepath"

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

	lspHandler, err := handler.NewLSPHandler("discord-rpc-lsp", "0.0.4")
	if err != nil {
		log.Fatalf("Failed to create LSP handler: %v", err)
	}

	server := lspHandler.NewServer()
	server.RunStdio()
}
