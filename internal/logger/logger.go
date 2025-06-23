package logger

import (
	"github.com/pterm/pterm"
	"github.com/vs-uulm/go-taf/pkg/config"
	"log/slog"
	"os"
	"strings"
)

/*
LOGGING HIERARCHY: DEBUG < INFO < WARN < ERROR
*/

const (
	PLAIN  = "PLAIN"
	PRETTY = "PRETTY"
	JSON   = "JSON"
)

// CreateMainLogger creates a new slog logger to be used as the main logger
func CreateMainLogger(configuration config.Log) *slog.Logger {
	switch strings.ToUpper(configuration.LogStyle) {
	case PRETTY:
		handler := pterm.NewSlogHandler(&pterm.DefaultLogger)
		logger := slog.New(handler)
		pterm.DefaultLogger.Level = configuration.LogLevel
		return logger
	case JSON:
		handlerOpts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewJSONHandler(os.Stderr, handlerOpts))
		slog.SetDefault(logger)
		return logger
	case PLAIN:
		fallthrough
	default:
		handlerOpts := &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(os.Stderr, handlerOpts))
		slog.SetDefault(logger)
		return logger
	}
}

// CreateChildLogger creates a child logger that appends the given contextName to the given mainLogger
func CreateChildLogger(mainLogger *slog.Logger, contextName string) *slog.Logger {
	group := slog.Group("Component", slog.String("Name", contextName))
	childLogger := mainLogger.With(group)
	return childLogger
}
