package crypto

import (
	"encoding/json"
	"errors"
	"github.com/horizon-connect-eu/crypto-library-interface/pkg/crypto"
	aivmsg "github.com/horizon-connect-eu/go-taf/pkg/message/aiv"
	tchmsg "github.com/horizon-connect-eu/go-taf/pkg/message/tch"
	"log/slog"
)

/*
The Crypto struct is a lightweight TAF-side wrapper around the crypto library.
*/
type Crypto struct {
	cryptoEnabled          bool
	attestationCertificate string
}

func NewCrypto(logger *slog.Logger, keyPath string, cryptoEnabled bool) (*Crypto, error) {

	cr := &Crypto{
		cryptoEnabled: cryptoEnabled,
	}
	if cr.cryptoEnabled {
		err := crypto.Init(logger, keyPath)
		if err != nil {
			return nil, err
		}
		cert, err := crypto.LoadAttestationCertificateInBase64()
		if err != nil {
			return nil, err
		} else {
			cr.attestationCertificate = cert
			return cr, nil
		}
	} else {
		cr.attestationCertificate = ""
		return cr, nil
	}
}

/*
AttestationCertificate returne the attestation certificate to be used for outbound messages sent by the TAF.
*/
func (cr *Crypto) AttestationCertificate() string {
	return cr.attestationCertificate
}

/*
SignAivRequest signs an AIV_REQUEST from the TAF to the AIV by setting message fields accordingly.
*/
func (cr *Crypto) SignAivRequest(request *aivmsg.AivRequest) error {
	if cr.cryptoEnabled {
		cryptoEvidence, err := crypto.GenerateEvidence()
		if err != nil {
			return errors.New("error generating evidence for AIV request")
		}
		request.Evidence.KeyRef = cryptoEvidence.KeyRef
		request.Evidence.Nonce = *cryptoEvidence.Nonce
		request.Evidence.Signature = cryptoEvidence.Signature
		request.Evidence.SignatureAlgorithmType = cryptoEvidence.SignatureAlgorithmType
		request.Evidence.Timestamp = cryptoEvidence.Timestamp
		return nil
	} else {
		//Don't do anything
		return nil
	}
}

/*
SignAivSubscribeRequest signs an AIV_SUBSCRIBE_REQUEST from the TAF to the AIV by setting message fields accordingly.
*/
func (cr *Crypto) SignAivSubscribeRequest(request *aivmsg.AivSubscribeRequest) error {
	if cr.cryptoEnabled {
		cryptoEvidence, err := crypto.GenerateEvidence()
		if err != nil {
			return errors.New("error generating evidence for AIV subscribe request")
		}
		request.Evidence.KeyRef = cryptoEvidence.KeyRef
		request.Evidence.Nonce = *cryptoEvidence.Nonce
		request.Evidence.Signature = cryptoEvidence.Signature
		request.Evidence.SignatureAlgorithmType = cryptoEvidence.SignatureAlgorithmType
		request.Evidence.Timestamp = cryptoEvidence.Timestamp
		return nil
	} else {
		//Don't do anything
		return nil
	}
}

/*
VerifyAivResponse verifies an incoming AIV_RESPONSE and returns true, false, or an error.
*/
func (cr *Crypto) VerifyAivResponse(response *aivmsg.AivResponse) (bool, error) {
	if cr.cryptoEnabled {
		nonceByteArray, err := crypto.FromHexToByteArray(response.AivEvidence.Nonce)
		if err != nil {
			return false, errors.New("failed to decode nonce from AIV_RESPONSE")
		}
		trusteeReportByteStream, err := json.Marshal(response.TrusteeReports)
		if err != nil {
			return false, errors.New("failed to decode trustee report from AIV_RESPONSE")
		}
		byteStreamToBeSigned := append(nonceByteArray, trusteeReportByteStream...)
		verificationResult, err := crypto.Verify(byteStreamToBeSigned, response.AivEvidence.Signature, response.AivEvidence.KeyRef+".pem")
		if err != nil {
			return false, err
		} else {
			return verificationResult, nil
		}
	} else {
		//Don't do anything
		return true, nil
	}
}

/*
VerifyAivNotify verifies an incoming AIV_NOTIFY and returns true, false, or an error.
*/
func (cr *Crypto) VerifyAivNotify(notify *aivmsg.AivNotify) (bool, error) {
	if cr.cryptoEnabled {
		nonceByteArray, err := crypto.FromHexToByteArray(notify.AivEvidence.Nonce)
		if err != nil {
			return false, errors.New("failed to decode nonce from AIV_NOTIFY")
		}
		trusteeReportByteStream, err := json.Marshal(notify.TrusteeReports)
		if err != nil {
			return false, errors.New("failed to decode trustee report from AIV_NOTIFY")
		}
		byteStreamToBeSigned := append(nonceByteArray, trusteeReportByteStream...)
		verificationResult, err := crypto.Verify(byteStreamToBeSigned, notify.AivEvidence.Signature, notify.AivEvidence.KeyRef+".pem")
		if err != nil {
			return false, err
		} else {
			return verificationResult, nil
		}
	} else {
		//Don't do anything
		return true, nil
	}
}

/*
VerifyTchNotify verifies an incoming TCH_NOTIFY and returns true, false, or an error.
*/
func (cr *Crypto) VerifyTchNotify(notify *tchmsg.TchNotify) (bool, error) {
	if cr.cryptoEnabled {
		trusteeReportByteStream, err := json.Marshal(notify.TchReport)
		if err != nil {
			return false, errors.New("failed to decode trustee report from AIV_RESPONSE")
		}
		byteStreamToBeSigned := trusteeReportByteStream //append(nonceByteArray, trusteeReportByteStream...)
		verificationResult, err := crypto.Verify(byteStreamToBeSigned, notify.Evidence.Signature, notify.Evidence.KeyRef+".pem")
		if err != nil {
			return false, err
		} else {
			return verificationResult, nil
		}
	} else {
		//Don't do anything
		return true, nil
	}
}
