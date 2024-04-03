package evidencecollection

import (
	"context"
	"fmt"
	"github.com/vs-uulm/go-taf/pkg/config"
	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/vs-uulm/go-taf/pkg/util"
)

// Holds the available functions for updating
// worker Results.
var adapters = map[string]Adapter{}

// Register a new ResultUpdater under a name.
// The name can be used in the config to refer to the registered function.
// The ResultUpdater is called by a worker at a point in execution when the
// Results it is responsible for should be refreshed.
func RegisterEvidenceCollectionAdapter(name string, f Adapter) {
	adapters[name] = f
}

type evidenceCollectionInterface struct {
	conf          config.Configuration
	adapters      map[int]Adapter
	inputChannels []chan message.EvidenceCollectionMessage
	outputChannel chan<- message.EvidenceCollectionMessage
}

func New(output chan message.EvidenceCollectionMessage,
	conf config.Configuration) (evidenceCollectionInterface, error) {
	evidenceCollector := evidenceCollectionInterface{
		conf:          conf,
		outputChannel: output,
		adapters:      make(map[int]Adapter),
	}

	for id, adapter := range conf.EvidenceCollection.Adapters {
		if f, ok := adapters[adapter.Name]; ok {
			channel := make(chan message.EvidenceCollectionMessage, conf.ChanBufSize)
			evidenceCollector.inputChannels = append(evidenceCollector.inputChannels, channel)
			evidenceCollector.adapters[id] = f
		} else {
			//LOG: log.Printf("[ECI] cannot find adapter plugin \"%s\"\n", adapter.Name)
		}
	}

	/*
		var err error
		f, err := getAdapterFactoryFunc(conf.EvidenceCollection.Adapters)
		if err != nil {
			return evidenceCollectionInterface{}, err
		}
		evidenceCollector.adapter = f
	*/
	return evidenceCollector, nil
}

func getAdapterFactoryFunc(name string) (Adapter, error) {
	if f, ok := adapters[name]; ok {
		return f, nil
	}
	return nil, fmt.Errorf("TrustAssessmentManager: no update result function named %s registered", name)
}

func (eci evidenceCollectionInterface) Run(ctx context.Context) {
	/*	defer func() {
			log.Println("EvidenceCollectionInterface: shutting down")
		}()
	*/
	//LOG: fmt.Println("Hello from ECI")

	for id, adapter := range eci.adapters {
		go adapter(ctx, id, eci.inputChannels[id], eci.conf)
	}
	util.Mux(eci.outputChannel, eci.inputChannels...)
}
