package trustmodel

import (
	logging "github.com/vs-uulm/go-taf/internal/logger"
	"github.com/vs-uulm/go-taf/pkg/command"
	"github.com/vs-uulm/go-taf/pkg/core"
	"github.com/vs-uulm/go-taf/pkg/manager"
	v2xmsg "github.com/vs-uulm/go-taf/pkg/message/v2x"
	"github.com/vs-uulm/go-taf/pkg/trustmodel/trustmodeltemplate"
	"log/slog"
	"strings"
)

type Manager struct {
	tafContext             core.TafContext
	channels               core.TafChannels
	logger                 *slog.Logger
	tam                    manager.TrustAssessmentManager
	tsm                    manager.TrustSourceManager
	trustModelTemplateRepo map[string]trustmodeltemplate.TrustModelTemplate
}

func NewManager(tafContext core.TafContext, channels core.TafChannels) (*Manager, error) {
	tmm := &Manager{
		tafContext:             tafContext,
		channels:               channels,
		logger:                 logging.CreateChildLogger(tafContext.Logger, "TMM"),
		trustModelTemplateRepo: TemplateRepository,
	}

	tmtNames := make([]string, len(tmm.trustModelTemplateRepo))
	i := 0
	for k := range tmm.trustModelTemplateRepo {
		tmtNames[i] = k
		i++
	}

	tmm.logger.Info("Initializing Trust Model Manager", "Available trust models", strings.Join(tmtNames, ", "))
	return tmm, nil
}

func (tmm *Manager) SetManagers(managers manager.TafManagers) {
	tmm.tam = managers.TAM
	tmm.tsm = managers.TSM
}

func (tmm *Manager) HandleV2xCpmMessage(cmd command.HandleOneWay[v2xmsg.V2XCpm]) {
	//TODO
}

func (tmm *Manager) ResolveTMT(identifier string) trustmodeltemplate.TrustModelTemplate {
	tmt, exists := tmm.trustModelTemplateRepo[identifier]
	if !exists {
		return nil
	} else {
		return tmt
	}
}
