package crypto

import (
	"crypto-library-interface/pkg/crypto"
	"encoding/json"
	"fmt"
	"log/slog"
)

func VerifyAivResponse(aivResposeBytestream []byte, trusteeReportByteStream []byte, logger *slog.Logger) {
	var jsonMap map[string]interface{}
	json.Unmarshal(aivResposeBytestream, &jsonMap)

	nonceByteArray, err := crypto.FromHexToByteArray(jsonMap["nonce"].(string))
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to convert hex to bytes: %w", err))
	}

	byteStreamToBeSigned := append(nonceByteArray, trusteeReportByteStream...)

	print("HELLO FROM VERIFICATION\n")

	//TODO: fix absolute path
	verificationResult, _ := crypto.Verify(byteStreamToBeSigned, jsonMap["signature"].(string), "/home/stef/workspace/connect/aiv/"+jsonMap["keyRef"].(string)+".pem")

	logger.Info(fmt.Sprintf("AIV_REQUEST verification status is [ %v ]\n", verificationResult))
	print("HELLO FROM VERIFICATION\n")
}
