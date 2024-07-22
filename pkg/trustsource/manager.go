package trustsource

import (
	"context"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	"log/slog"
)

type trustSourceManager struct {
	tafContext core.TafContext
	channels   core.TafChannels
	logger     *slog.Logger
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (trustSourceManager, error) {
	tsm := trustSourceManager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
	}
	return tsm, nil
}

func (tsm *trustSourceManager) Run() {
	// Cleanup function:
	defer func() {
		tsm.logger.Info("Shutting down")
	}()

	for {
		// Each iteration, check whether we've been cancelled.
		if err := context.Cause(tsm.tafContext.Context); err != nil {
			return
		}
		select {
		case <-tsm.tafContext.Context.Done():
			if len(tsm.channels.TMMChan) != 0 || len(tsm.channels.TAMChan) != 0 || len(tsm.channels.TSMChan) != 0 {
				continue
			}
			return
		case incomingCmd := <-tsm.channels.TSMChan:

			switch cmd := incomingCmd.(type) {
			case command.HandleResponse[aivmsg.AivResponse]:
				tsm.handleAivResponse(cmd)
			case command.HandleResponse[aivmsg.AivSubscribeResponse]:
				tsm.handleAivSubscribeResponse(cmd)
			case command.HandleResponse[aivmsg.AivUnsubscribeResponse]:
				tsm.handleAivUnsubscribeResponse(cmd)
			case command.HandleNotify[aivmsg.AivNotify]:
				tsm.handleAivNotify(cmd)
			case command.HandleResponse[mbdmsg.MBDSubscribeResponse]:
				tsm.handleMbdSubscribeResponse(cmd)
			case command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]:
				tsm.handleMbdUnsubscribeResponse(cmd)
			case command.HandleNotify[mbdmsg.MBDNotify]:
				tsm.handleMbdNotify(cmd)
			default:
				tsm.logger.Warn("Command with no associated handling logic received by TSM", "Command Type", cmd.Type())
			}
		}
	}
}

/* ------------ ------------ AIV Message Handling ------------ ------------ */

func (t *trustSourceManager) handleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	t.logger.Info("TODO: handle AIV_RESPONSE")
}

func (t *trustSourceManager) handleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse]) {
	t.logger.Info("TODO: handle AIV_SUBSCRIBE_RESPONSE")
}

func (t *trustSourceManager) handleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse]) {
	t.logger.Info("TODO: handle AIV_UNSUBSCRIBE_RESPONSE")
}

func (t *trustSourceManager) handleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify]) {
	t.logger.Info("TODO: handle AIV_NOTIFY")
}

/* ------------ ------------ MBD Message Handling ------------ ------------ */

func (t *trustSourceManager) handleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse]) {
	t.logger.Info("TODO: handle MBD_SUBSCRIBE_RESPONSE")
}

func (t *trustSourceManager) handleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]) {
	t.logger.Info("TODO: handle MBD_UNSUBSCRIBE_RESPONSE")
}

func (t *trustSourceManager) handleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify]) {
	t.logger.Info("TODO: handle MBD_NOTIFY")
}
