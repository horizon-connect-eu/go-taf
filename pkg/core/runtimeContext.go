package core

import (
	"context"
	"github.com/vs-uulm/go-taf/pkg/config"
	"log/slog"
)

// A context struct that captures several relevant properties needed by different subcomponents
type RuntimeContext struct {
	Configuration config.Configuration
	Logger        *slog.Logger
	Context       context.Context
	Identifier    string
}
