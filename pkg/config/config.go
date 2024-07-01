package config

import (
	"encoding/json"
	"github.com/pterm/pterm"
	"os"
)

// Configuration of the TAF, including its subcomponents.
type Configuration struct {
	Logging                    LogConfiguration
	ChanBufSize                int
	V2X                        V2XConfiguration
	TAM                        TAMConfiguration
	EvidenceCollection         EvidenceCollectionConfiguration
	CommunicationConfiguration CommunicationConfiguration
	TLEE                       TLEEConfig
}

type CommunicationConfiguration struct {
	Handler string
	Kafka   KafkaConfig
}

type KafkaConfig struct {
	Broker string
	Topics []string
}

type LogConfiguration struct {
	LogLevel pterm.LogLevel
	LogStyle string //PLAIN, PRETTY, JSON
}

// V2XConfiguration stores the config of the [v2xlistener].
type V2XConfiguration struct {
	SendIntervalNs int
}

// TAMConfiguration stores the config of the [tam.tam].
type TAMConfiguration struct {
	TrustModelInstanceShards int
	UpdateStateOp            string
	UpdateResultsOp          string
}

// TAMConfiguration stores the config of the [tam.tam].
type EvidenceCollectionConfiguration struct {
	Adapters []AdapterConfig
}

type AdapterConfig struct {
	Name   string
	Params map[string]string
}

type TLEEConfig struct {
	UseInternalTLEE bool
}

var (
	// Default configuration of the TAF.
	// This configuration will be used if no configuration
	// file is specified explicitly by the user.
	// In case the user-specified configuration file
	// misses values, this struct defines the corresponding
	// default values.
	DefaultConfig Configuration = Configuration{
		Logging:     LogConfiguration{LogLevel: pterm.LogLevelDebug, LogStyle: "PRETTY"},
		ChanBufSize: 1_000,
		V2X: V2XConfiguration{
			SendIntervalNs: 1_000_000_000,
		},
		TAM: TAMConfiguration{
			TrustModelInstanceShards: 1,
			UpdateResultsOp:          "add",
			UpdateStateOp:            "TODO", //TODO
		},
		EvidenceCollection: EvidenceCollectionConfiguration{
			Adapters: []AdapterConfig{
				{"filebased", map[string]string{"path": "res/file_based_evidence_1.csv"}},
			},
		},
		CommunicationConfiguration: CommunicationConfiguration{
			Handler: "file-based",
			Kafka: KafkaConfig{
				Broker: "localhost:9092",
				Topics: []string{"taf"},
			},
		},
		TLEE: TLEEConfig{
			UseInternalTLEE: true,
		},
	}
)

// Load a configuration from a JSON file.
func LoadJSON(filepath string) (Configuration, error) {
	// TODO figure out whether deep-copy is necessary here.
	config := DefaultConfig
	raw, err := os.ReadFile(filepath)
	if err != nil {
		return Configuration{}, err
	}
	err = json.Unmarshal(raw, &config)
	if err != nil {
		return Configuration{}, err
	}
	return config, nil
}
