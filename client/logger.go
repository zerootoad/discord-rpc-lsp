package client

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	currentLogFile *os.File
	logMutex       sync.Mutex
)

func InitLogger(logfilePath string, level string, output string) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if currentLogFile != nil {
		_ = currentLogFile.Close()
		currentLogFile = nil
	}

	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Warn().Msgf("Invalid logging level '%s'. Defaulting to 'info'.", level)
	}

	var writer io.Writer
	switch output {
	case "file":
		file, err := os.OpenFile(logfilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to open log file")
		}
		currentLogFile = file
		writer = file
	case "stdout":
		writer = os.Stdout
	default:
		writer = os.Stdout
		log.Warn().Msgf("Invalid logging output '%s'. Defaulting to 'stdout'.", output)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        writer,
		TimeFormat: time.RFC3339,
	}).With().
		Timestamp().
		Logger()
}

func Info(msg string, fields map[string]any) {
	event := log.Info()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

func Error(msg string, fields map[string]any) {
	event := log.Error()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

func Warn(msg string, fields map[string]any) {
	event := log.Warn()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

func Debug(msg string, fields map[string]any) {
	event := log.Debug()
	for k, v := range fields {
		event.Interface(k, v)
	}
	event.Msg(msg)
}

func WithDuration(start time.Time, msg string, fields map[string]any) {
	if fields == nil {
		fields = make(map[string]any)
	}
	fields["elapsed"] = time.Since(start)
	Info(msg, fields)
}
