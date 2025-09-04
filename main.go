package main

import (
	"os"
	"path/filepath"

	"github.com/zerootoad/discord-rpc-lsp/client"
	"github.com/zerootoad/discord-rpc-lsp/handler"
	"github.com/zerootoad/discord-rpc-lsp/utils"
)

func main() {
	homedir := utils.GetUserHomeDir()
	configDir := filepath.Join(homedir, ".discord-rpc-lsp")

	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		client.Error("Failed to create config directory", map[string]any{
			"configDir": configDir,
			"error":     err,
		})
	}

	logFilePath := filepath.Join(configDir, "lsp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		client.Error("Failed to open log file", map[string]any{
			"logFilePath": logFilePath,
			"error":       err,
		})
	}
	defer logFile.Close()

	client.InitLogger(logFilePath, "info", "stdout")
	client.Debug("Starting app with default logger", nil)

	configFilePath := filepath.Join(configDir, "config.toml")
	config, err := client.LoadConfig(configFilePath)
	if err != nil {
		client.Error("Failed to load or create configuration", map[string]any{
			"configFilePath": configFilePath,
			"error":          err,
		})
		config = client.DefaultConfig()
	}

	client.InitLogger(logFilePath, config.Logging.Level, config.Logging.Output)
	client.Debug("Logger re-initialized with user config", map[string]any{
		"level":  config.Logging.Level,
		"output": config.Logging.Output,
	})

	lspHandler, err := handler.NewLSPHandler("discord-rpc-lsp", "1.0.1", config)
	if err != nil {
		client.Error("Failed to create LSP handler", map[string]any{
			"error": err,
		})
	}

	server := lspHandler.NewServer()
	client.Debug("Starting LSP server", map[string]any{})
	err = server.RunStdio()
	if err != nil {
		client.Error("Failed to run stdio server", map[string]any{
			"error": err,
		})
	}

}
