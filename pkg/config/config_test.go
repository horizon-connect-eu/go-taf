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

	//fmt.Printf("%+v\n", conf)

	if conf.ChanBufSize != 1337 ||
		conf.V2X.SendIntervalNs != 42 ||
		conf.EvidenceCollection.Adapters[0].Name != "filebased" ||
		conf.EvidenceCollection.Adapters[0].Params["path"] != "res/file_based_evidence_collection_1.csv" {
		t.Error("Read valid config file incorrectly")
	}
}
