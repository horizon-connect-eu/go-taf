package trustmodel

import (
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"log/slog"
	"strings"
)

type Manager struct {
	tafContext core.TafContext
	logger     *slog.Logger
	tam        manager.TrustAssessmentManager
	tsm        manager.TrustSourceManager
	//trustmodeltemplate identifier->TMT
	trustModelTemplateRepo map[string]core.TrustModelTemplate
	v2xObserver            v2xObserver
	crypto                 *crypto.Crypto
	outbox                 chan core.Message
}

type v2xListener struct{}

func (v v2xListener) handleNodeAdded(identifier string) {
	//fmt.Println("Node added: " + identifier)
	//TODO
}

func (v v2xListener) handleNodeRemoved(identifier string) {
	//fmt.Println("Node removed: " + identifier)
	//TODO
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tmm := &Manager{
		tafContext:             tafContext,
		logger:                 logging.CreateChildLogger(tafContext.Logger, "TMM"),
		trustModelTemplateRepo: TemplateRepository,
		v2xObserver:            CreateListener(tafContext.Configuration.V2X.NodeTTLsec, tafContext.Configuration.V2X.CheckIntervalSec),
		crypto:                 tafContext.Crypto,
		outbox:                 channels.OutgoingMessageChannel,
	}

	tmtNames := make([]string, len(tmm.trustModelTemplateRepo))
	i := 0
	for k := range tmm.trustModelTemplateRepo {
		tmtNames[i] = k
		i++
	}

	tmm.v2xObserver.registerObserver(v2xListener{})

	tmm.logger.Info("Initializing Trust Model Manager", "Available trust models", strings.Join(tmtNames, ", "))
	for _, tmt := range tmtNames {
		tmm.logger.Info(tmt, "Description", tmm.trustModelTemplateRepo[tmt].Description(), "Evidence Sources", fmt.Sprintf("%+v", tmm.trustModelTemplateRepo[tmt].EvidenceTypes()))

	}
	return tmm, nil
}

func (tmm *Manager) SetManagers(managers manager.TafManagers) {
	tmm.tam = managers.TAM
	tmm.tsm = managers.TSM
}

func (tmm *Manager) HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm]) {
	sender := fmt.Sprintf("%g", cmd.OneWay.SourceID)
	tmm.v2xObserver.AddNode(sender)
}

func (tmm *Manager) ResolveTMT(identifier string) core.TrustModelTemplate {
	tmt, exists := tmm.trustModelTemplateRepo[identifier]
	if !exists {
		return nil
	} else {
		return tmt
	}
}
