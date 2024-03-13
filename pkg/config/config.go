package config

import (
	"encoding/json"
	"os"
)

// Configuration of the TAF, including its subcomponents.
type Configuration struct {
	ChanBufSize int
	V2XConfig   V2XConfiguration
	TAMConfig   TAMConfiguration
}

// Configuration of the [v2xlistener].
type V2XConfiguration struct {
	SendIntervalNs int
}

type TAMConfiguration struct {
	TrustModelInstanceShards int
}

var (
	// Default configuration of the TAF.
	// This configuration will be used if no configuration
	// file is specified explicitly by the user.
	// In case the user-specified configuration file
	// misses values, this struct defines the corresponding
	// default values.
	DefaultConfig Configuration = Configuration{
		ChanBufSize: 1_000_000,
		V2XConfig: V2XConfiguration{
			SendIntervalNs: 1,
		},
		TAMConfig: TAMConfiguration{
			TrustModelInstanceShards: 1,
		},
	}
)

// Load a configuration from a JSON file.
func LoadJson(filepath string) (Configuration, error) {
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
