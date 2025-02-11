package trustmodel

import (
	"fmt"
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/communication"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/crypto"
	"github.com/vs-uulm/go-taf/pkg/manager"
	messages "github.com/vs-uulm/go-taf/pkg/message"
	tasmsg "github.com/vs-uulm/go-taf/pkg/message/tas"
	tchmsg "github.com/vs-uulm/go-taf/pkg/message/tch"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	session2 "github.com/vs-uulm/go-taf/pkg/trustmodel/session"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodelupdate"
	"log/slog"
	"regexp"
	"strings"
)

type Manager struct {
	tafContext core.TafContext
	logger     *slog.Logger
	tam        manager.TrustAssessmentManager
	tsm        manager.TrustSourceManager
	//trustmodeltemplate identifier->TMT
	trustModelTemplateRepo map[string]core.TrustModelTemplate
	v2xObserver            EntityObserver //observer based on V2X_CPM messages
	tchObserver            EntityObserver //observer based on TCH messages
	crypto                 *crypto.Crypto
	outbox                 chan core.Message
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tmm := &Manager{
		tafContext:             tafContext,
		logger:                 logging.CreateChildLogger(tafContext.Logger, "TMM"),
		trustModelTemplateRepo: TemplateRepository,
		v2xObserver:            CreateListener(tafContext.Configuration.V2X.NodeTTLsec, tafContext.Configuration.V2X.CheckIntervalSec),
		tchObserver:            CreateListener(tafContext.Configuration.V2X.NodeTTLsec, tafContext.Configuration.V2X.CheckIntervalSec),
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
				tmm.v2xObserver.registerObserver(newV2xObserver(tmm))
			case core.TRUSTEE_TRIGGERED_TRUST_MODEL:
				//register TMM as handler for TCH observer
				tmm.tchObserver.registerObserver(newTchObserver(tmm))
			default: //Nothing to do
			}
		}
	}

}

func (tmm *Manager) initializeTrustModelTemplates() {
	for tmtName, tmt := range tmm.trustModelTemplateRepo {
		tmm.logger.Debug(tmtName, "Description", tmt.Description(), "Evidence Sources", fmt.Sprintf("%+v", tmt.EvidenceTypes()), "Trust Model Type", tmt.Type(), "Signing Hash", tmt.SigningHash())
	}
}

func (tmm *Manager) SetManagers(managers manager.TafManagers) {
	tmm.tam = managers.TAM
	tmm.tsm = managers.TSM
}

func (tmm *Manager) HandleTchNotify(cmd command.HandleNotify[tchmsg.TchNotify]) {
	r := regexp.MustCompile("^vehicle\\_(\\d+)$")
	match := r.FindStringSubmatch(cmd.Notify.TchReport.TrusteeID)

	if match == nil || len(match) < 2 {
		tmm.logger.Warn("Invalid trustee ID for TCH Notify")
		return
	}
	trusteeId := match[1]
	tmm.tchObserver.AddNode(trusteeId)
}

func (tmm *Manager) HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm]) {
	sender := fmt.Sprintf("%g", cmd.OneWay.SourceID)
	tmm.v2xObserver.AddNode(sender)

	//check whether TMIs are interesteed in RefreshCPM messages
	targetTMIIDs := make([]string, 0)
	for _, tmt := range tmm.trustModelTemplateRepo {
		if tmt.Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL {
			//relevant TMIs must have a VEHICLE_TRIGGERED_TRUST_MODEL TMT and the TMI ID must be identical to the sender
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

		if len(objects) > 0 {
			for _, fullTMIID := range targetTMIIDs {
				updateCmd := command.CreateHandleTMIUpdate(fullTMIID, trustmodelupdate.CreateRefreshCPM(sender, objects))
				tmm.tam.DispatchToWorkerByFullTMIID(fullTMIID, updateCmd)
			}
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

func (tmm *Manager) GetAllTMTs() []core.TrustModelTemplate {
	tmts := make([]core.TrustModelTemplate, len(tmm.trustModelTemplateRepo))

	i := 0
	for _, v := range tmm.trustModelTemplateRepo {
		tmts[i] = v
		i++
	}
	return tmts
}

func (tmm *Manager) handleV2XNodeAdded(identifier string) {
	tmm.logger.Debug("New sender vehicle added", "Identifier", identifier)
	for sessionID, session := range tmm.tam.Sessions() {
		if session.TrustModelTemplate().Type() == core.VEHICLE_TRIGGERED_TRUST_MODEL && session.State() == session2.ESTABLISHED {
			spawner := session.DynamicSpawner()
			if spawner != nil {
				tmi, err := spawner.OnNewVehicle(identifier, nil)
				if err != nil {
					tmm.logger.Warn("Error while spawning trust model instance", "TMT", session.TrustModelTemplate(), "Identifier used for dynamic spawning", identifier)
				} else {
					tmi.Initialize(map[string]interface{}{
						"sourceID": identifier,
					})
					tmm.tam.AddNewTrustModelInstance(tmi, sessionID)
				}
			}
		}
	}

}

func (tmm *Manager) handleV2XNodeRemoved(identifier string) {
	tmm.logger.Debug("Sender vehicle removed", "Identifier", identifier)

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
				tmm.tam.RemoveTrustModelInstance(fullTMIID, sessionID)
			}
		}
	}
}

func (tmm *Manager) handleTCHNodeAdded(identifier string) {
	tmm.logger.Debug("New trustee added", "Identifier", identifier)
	for sessionID, session := range tmm.tam.Sessions() {
		if session.TrustModelTemplate().Type() == core.TRUSTEE_TRIGGERED_TRUST_MODEL && session.State() == session2.ESTABLISHED {
			spawner := session.DynamicSpawner()
			if spawner != nil {
				tmi, err := spawner.OnNewTrustee(identifier, nil)
				if err != nil {
					tmm.logger.Warn("Error while spawning trust model instance", "TMT", session.TrustModelTemplate(), "Identifier used for dynamic spawning", identifier)
				} else {
					tmi.Initialize(map[string]interface{}{
						"trusteeID": identifier,
					})
					tmm.tam.AddNewTrustModelInstance(tmi, sessionID)
				}
			}
		}
	}
}

func (tmm *Manager) handleTCHNodeRemoved(identifier string) {
	tmm.logger.Debug("TCH trustee removed", "Identifier", identifier)

	targetTMIIDs := make([]string, 0)
	for _, tmt := range tmm.trustModelTemplateRepo {
		if tmt.Type() == core.TRUSTEE_TRIGGERED_TRUST_MODEL {
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
				tmm.tam.RemoveTrustModelInstance(fullTMIID, sessionID)
			}
		}
	}
}

func (tmm *Manager) HandleTasTmtDiscover(cmd command.HandleRequest[tasmsg.TasTmtDiscover]) {
	if cmd.Request.TrustModelTemplates {

		tmts := make(map[string]tasmsg.TrustModelTemplate)

		for _, tmt := range tmm.GetAllTMTs() {
			tmts[tmt.Identifier()] = tasmsg.TrustModelTemplate{
				Name:        tmt.TemplateName(),
				Version:     tmt.Version(),
				Description: tmt.Description(),
				Hash:        tmt.SigningHash(),
			}
		}

		response := tasmsg.TasTmtOffer{
			TrustModelTemplates: tmts,
		}

		bytes, err := communication.BuildResponse(tmm.tafContext.Configuration.Communication.TafEndpoint, messages.TAS_TMT_OFFER, cmd.RequestID, response)
		if err != nil {
			tmm.logger.Error("Error marshalling response", "error", err)
			return
		}
		tmm.outbox <- core.NewMessage(bytes, "", cmd.ResponseTopic)
	} else {
		tmm.logger.Error("Invalid TAS_TMT_DISCOVER received.")
	}
}

func (tmm *Manager) ListRecentV2XNodes() []string {
	return tmm.v2xObserver.Nodes()
}
func (tmm *Manager) ListRecentTrustees() []string {
	return tmm.tchObserver.Nodes()
}

type v2xObserver struct {
	tmm *Manager
}

func newV2xObserver(tmm *Manager) *v2xObserver {
	return &v2xObserver{
		tmm: tmm,
	}
}

func (v2xObs *v2xObserver) handleNodeAdded(identifier string) {
	v2xObs.tmm.handleV2XNodeAdded(identifier)
}
func (v2xObs *v2xObserver) handleNodeRemoved(identifier string) {
	v2xObs.tmm.handleV2XNodeRemoved(identifier)

}

type tchObserver struct {
	tmm *Manager
}

func newTchObserver(tmm *Manager) *tchObserver {
	return &tchObserver{
		tmm: tmm,
	}
}

func (tchObs *tchObserver) handleNodeAdded(identifier string) {
	tchObs.tmm.handleTCHNodeAdded(identifier)
}
func (tchObs *tchObserver) handleNodeRemoved(identifier string) {
	tchObs.tmm.handleTCHNodeRemoved(identifier)

}
