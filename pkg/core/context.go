package core

import (
	"context"
	"github.com/horizon-connect-eu/go-taf/pkg/config"
	"github.com/horizon-connect-eu/go-taf/pkg/crypto"
	"log/slog"
)

/*
The TafContext struct captures several relevant properties needed by different subcomponents.
*/
type TafContext struct {
	Configuration config.Configuration
	Logger        *slog.Logger
	Context       context.Context
	Identifier    string
	Crypto        *crypto.Crypto
}
