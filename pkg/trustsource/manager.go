package trustsource

import (
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/manager"
	aivmsg "github.com/vs-uulm/go-taf/pkg/message/aiv"
	mbdmsg "github.com/vs-uulm/go-taf/pkg/message/mbd"
	"log/slog"
)

type Manager struct {
	tafContext core.TafContext
	channels   core.TafChannels
	logger     *slog.Logger
	tam        manager.TrustAssessmentManager
	tmm        manager.TrustModelManager
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tsm := &Manager{
		tafContext: tafContext,
		channels:   channels,
		logger:     logging.CreateChildLogger(tafContext.Logger, "TSM"),
	}
	return tsm, nil
}

func (tsm *Manager) SetManagers(managers manager.TafManagers) {
	tsm.tam = managers.TAM
	tsm.tmm = managers.TMM
}

/* ------------ ------------ AIV Message Handling ------------ ------------ */

func (tsm *Manager) HandleAivResponse(cmd command.HandleResponse[aivmsg.AivResponse]) {
	tsm.logger.Info("TODO: handle AIV_RESPONSE")
}

func (tsm *Manager) HandleAivSubscribeResponse(cmd command.HandleResponse[aivmsg.AivSubscribeResponse]) {
	tsm.logger.Info("TODO: handle AIV_SUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleAivUnsubscribeResponse(cmd command.HandleResponse[aivmsg.AivUnsubscribeResponse]) {
	tsm.logger.Info("TODO: handle AIV_UNSUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleAivNotify(cmd command.HandleNotify[aivmsg.AivNotify]) {
	tsm.logger.Info("TODO: handle AIV_NOTIFY")
	tsm.tam.Hello()
}

/* ------------ ------------ MBD Message Handling ------------ ------------ */

func (tsm *Manager) HandleMbdSubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDSubscribeResponse]) {
	tsm.logger.Info("TODO: handle MBD_SUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleMbdUnsubscribeResponse(cmd command.HandleResponse[mbdmsg.MBDUnsubscribeResponse]) {
	tsm.logger.Info("TODO: handle MBD_UNSUBSCRIBE_RESPONSE")
}

func (tsm *Manager) HandleMbdNotify(cmd command.HandleNotify[mbdmsg.MBDNotify]) {
	tsm.logger.Info("TODO: handle MBD_NOTIFY")
}
