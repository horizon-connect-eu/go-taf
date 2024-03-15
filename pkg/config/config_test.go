package config

import (
	"testing"
)

func TestLoadJson(t *testing.T) {
	_, err := LoadJSON("../../test/invalidjson.json")
	if err == nil {
		t.Error("No error on malformed JSON config")
	}

	conf, err := LoadJSON("../../test/valid.json")
	if err != nil {
		t.Error("Error on existing and valid JSON config file")
	}
	if conf.ChanBufSize != 1337 || conf.V2XConfig.SendIntervalNs != 42 {
		t.Error("Read valid config file incorrectly")
	}
}
