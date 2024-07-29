package core

import (
	"context"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"log/slog"
)

// A context struct that captures several relevant properties needed by different subcomponents
type TafContext struct {
	Configuration config.Configuration
	Logger        *slog.Logger
	Context       context.Context
	Identifier    string
	Crypto        *crypto.Crypto
}
