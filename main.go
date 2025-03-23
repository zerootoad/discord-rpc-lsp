package main

import (
	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"configDir": configDir,
			"error":     err,
		}).Fatal("Failed to create config directory")
	}

	logFilePath := filepath.Join(configDir, "lsp.log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"logFilePath": logFilePath,
			"error":       err,
		}).Fatal("Failed to open log file")
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)

	log.AddHook(&writerHook{
		Writer: os.Stdout,
		LogLevels: []log.Level{
			log.InfoLevel,
			log.WarnLevel,
			log.ErrorLevel,
			log.FatalLevel,
			log.PanicLevel,
		},
	})

	lspHandler, err := handler.NewLSPHandler("discord-rpc-lsp", "0.0.4")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to create LSP handler")
	}

	server := lspHandler.NewServer()
	log.Info("Starting LSP server")
	err = server.RunStdio()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Failed to run stdio server")
	}

}

type writerHook struct {
	Writer    *os.File
	LogLevels []log.Level
}

func (hook *writerHook) Fire(entry *log.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

func (hook *writerHook) Levels() []log.Level {
	return hook.LogLevels
}
