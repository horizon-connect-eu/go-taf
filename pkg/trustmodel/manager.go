package trustmodel

import (
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	session2 "github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
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

	tmm.logger.Info("Initializing Trust Model Manager", "Available trust models", strings.Join(tmtNames, ", "))

	tmm.initializeTrustModelTemplateTypes()
	tmm.initializeTrustModelTemplates()

	return tmm, nil
}

func (tmm *Manager) initializeTrustModelTemplateTypes() {

	availableTypes := make(map[core.TrustModelTemplateType]bool)
	for _, tmt := range tmm.trustModelTemplateRepo {
		availableTypes[tmt.Type()] = true
	}

	for tmtType, available := range availableTypes {
		if available {
			switch tmtType {
			case core.VEHICLE_TRIGGERED_TRUST_MODEL:
				//register TMM as handler for V2X observer
				tmm.v2xObserver.registerObserver(tmm)
			}
		}
	}

}

func (tmm *Manager) initializeTrustModelTemplates() {
	for tmtName, tmt := range tmm.trustModelTemplateRepo {
		tmm.logger.Info(tmtName, "Description", tmt.Description(), "Evidence Sources", fmt.Sprintf("%+v", tmt.EvidenceTypes()), "Trust Model Type", tmt.Type())
	}
}

func (tmm *Manager) SetManagers(managers manager.TafManagers) {
	tmm.tam = managers.TAM
	tmm.tsm = managers.TSM
}

func (tmm *Manager) HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm]) {
	sender := fmt.Sprintf("%g", cmd.OneWay.SourceID)
	tmm.v2xObserver.AddNode(sender)

	//TODO: check whether RefreshCPM messages are necessary

	targetTMIIDs := make([]string, 0)
	for _, tmt := range tmm.trustModelTemplateRepo {
		if tmt.Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL {
			results, err := tmm.tam.QueryTMIs("//*/*/" + tmt.Identifier() + "/" + sender)
			if err == nil {
				targetTMIIDs = append(targetTMIIDs, results...)
			}
		}
	}

	if len(targetTMIIDs) > 0 {
		objects := make([]string, 0)

		for _, object := range cmd.OneWay.PerceivedObjectContainer.Objects {
			objects = append(objects, fmt.Sprintf("%g", object.ObjectID))
		}

		for _, fullTMIID := range targetTMIIDs {
			updateCmd := command.CreateHandleTMIUpdate(fullTMIID, trustmodelupdate.CreateRefreshCPM(sender, objects))
			tmm.tam.DispatchToWorkerByFullTMIID(fullTMIID, updateCmd)
		}
	}
}

func (tmm *Manager) ResolveTMT(identifier string) core.TrustModelTemplate {
	tmt, exists := tmm.trustModelTemplateRepo[identifier]
	if !exists {
		return nil
	} else {
		return tmt
	}
}

func (tmm *Manager) handleNodeAdded(identifier string) {
	tmm.logger.Info("New node added", "Identifier", identifier)
	for sessionID, session := range tmm.tam.Sessions() {
		if session.TrustModelTemplate().Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL && session.State() == session2.ESTABLISHED {
			spawner := session.DynamicSpawner()
			if spawner != nil {
				tmi, err := spawner.OnNewVehicle(identifier, nil)
				if err != nil {
					tmm.logger.Warn("Error while spawning trust model instance", "TMT", session.TrustModelTemplate(), "Identifier used for dynamic spawning", identifier)
				} else {
					tmi.Initialize(nil) //TODO: Params?
					tmm.tam.AddNewTrustModelInstance(tmi, sessionID)
				}
			}
		}
	}

}

func (tmm *Manager) handleNodeRemoved(identifier string) {
	tmm.logger.Info("Node removed", "Identifier", identifier)

	targetTMIIDs := make([]string, 0)
	for _, tmt := range tmm.trustModelTemplateRepo {
		if tmt.Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL {
			results, err := tmm.tam.QueryTMIs("//*/*/" + tmt.Identifier() + "/" + identifier)
			if err == nil {
				targetTMIIDs = append(targetTMIIDs, results...)
			}
		}
	}

	sessions := tmm.tam.Sessions()

	for _, fullTMIID := range targetTMIIDs {
		_, sessionID, _, tmiID := core.SplitFullTMIIdentifier(fullTMIID)
		if session, exists := sessions[sessionID]; exists && sessions[sessionID].State() == session2.ESTABLISHED {
			sessionTMIs := session.TrustModelInstances()
			if _, tmiExists := sessionTMIs[tmiID]; tmiExists {
				tmm.tam.RemoveTrustModelInstance(tmiID, sessionID)
			}
		}
	}
}

func (tmm *Manager) ListRecentV2XNodes() []string {
	return tmm.v2xObserver.Nodes()
}
