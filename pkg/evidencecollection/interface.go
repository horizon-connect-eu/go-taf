package evidencecollection

import (
	"context"
	"fmt"

	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/config"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/message"
	"gitlab-vs.informatik.uni-ulm.de/connect/taf-brussels-demo/pkg/util"
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
	adapters      []Adapter
	inputChannels []chan message.EvidenceCollectionMessage
	outputChannel chan<- message.EvidenceCollectionMessage
}

func New(output chan message.EvidenceCollectionMessage,
	conf config.Configuration) (evidenceCollectionInterface, error) {
	evidenceCollector := evidenceCollectionInterface{
		conf:          conf,
		outputChannel: output,
	}

	for _, adapter := range conf.EvidenceCollection.Adapters {
		fmt.Println(adapter)
		if f, ok := adapters[adapter]; ok {
			channel := make(chan message.EvidenceCollectionMessage, conf.ChanBufSize)
			evidenceCollector.inputChannels = append(evidenceCollector.inputChannels, channel)
			evidenceCollector.adapters = append(evidenceCollector.adapters, f)
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
	fmt.Println("Hello from ECI")

	for i, adapter := range eci.adapters {
		go adapter(eci.inputChannels[i], eci.conf)
	}
	util.MuxMany(eci.inputChannels, eci.outputChannel)
}
