package crypto

import (
	"crypto-library-interface/pkg/crypto"
	"encoding/json"
	"fmt"
	"log/slog"
)

type Crypto struct {
	cryptoEnabled bool
}

func NewCrypto(logger *slog.Logger, keyPath string, cryptoEnabled bool) (*Crypto, error) {
	err := crypto.Init(logger, keyPath)
	if err != nil {
		return nil, err
	} else {
		cr := &Crypto{
			cryptoEnabled: cryptoEnabled,
		}
		return cr, nil
	}
}

func (cr *Crypto) addAttestationCertificate() {
	if cr.cryptoEnabled {
		//TODO
	}
}

func (cr *Crypto) VerifyAivResponse(aivResposeBytestream []byte, trusteeReportByteStream []byte, logger *slog.Logger) {
	var jsonMap map[string]interface{}
	json.Unmarshal(aivResposeBytestream, &jsonMap)

	nonceByteArray, err := crypto.FromHexToByteArray(jsonMap["nonce"].(string))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to convert hex to bytes: %s", err))
	}

	byteStreamToBeSigned := append(nonceByteArray, trusteeReportByteStream...)

	print("HELLO FROM VERIFICATION\n")

	//TODO: fix absolute path
	verificationResult, _ := crypto.Verify(byteStreamToBeSigned, jsonMap["signature"].(string), "/home/stef/workspace/connect/aiv/"+jsonMap["keyRef"].(string)+".pem")

	logger.Info(fmt.Sprintf("AIV_REQUEST verification status is [ %v ]\n", verificationResult))
	print("HELLO FROM VERIFICATION\n")
}
