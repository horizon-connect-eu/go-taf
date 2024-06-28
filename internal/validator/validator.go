package validator

import (
	"fmt"
	"github.com/vs-uulm/go-taf/internal/projectpath"
	"github.com/xeipuuv/gojsonschema"
)

type MessageSchema string

const (
	AIV_NOTIFY                    = "AIV_NOTIFY.json"
	AIV_REQUEST                   = "AIV_REQUEST.json"
	AIV_RESPONSE                  = "AIV_RESPONSE.json"
	AIV_SUBSCRIBE_REQUEST         = "AIV_SUBSCRIBE_REQUEST.json"
	AIV_SUBSCRIBE_RESPONSE        = "AIV_SUBSCRIBE_RESPONSE.json"
	AIV_UNSUBSCRIBE_REQUEST       = "AIV_UNSUBSCRIBE_REQUEST.json"
	AIV_UNSUBSCRIBE_RESPONSE      = "AIV_UNSUBSCRIBE_RESPONSE.json"
	GENERIC_ONE_WAY               = "GENERIC_ONE_WAY.json"
	GENERIC_REQUEST               = "GENERIC_REQUEST.json"
	GENERIC_RESPONSE              = "GENERIC_RESPONSE.json"
	GENERIC_SUBSCRIPTION_NOTIFY   = "GENERIC_SUBSCRIPTION_NOTIFY.json"
	GENERIC_SUBSCRIPTION_REQUEST  = "GENERIC_SUBSCRIPTION_REQUEST.json"
	GENERIC_SUBSCRIPTION_RESPONSE = "GENERIC_SUBSCRIPTION_RESPONSE.json"
	MBD_NOTIFY                    = "MBD_NOTIFY.json"
	MBD_SUBSCRIBE_REQUEST         = "MBD_SUBSCRIBE_REQUEST.json"
	MBD_SUBSCRIBE_RESPONSE        = "MBD_SUBSCRIBE_RESPONSE.json"
	MBD_UNSUBSCRIBE_REQUEST       = "MBD_UNSUBSCRIBE_REQUEST.json"
	MBD_UNSUBSCRIBE_RESPONSE      = "MBD_UNSUBSCRIBE_RESPONSE.json"
	TAS_INIT_REQUEST              = "TAS_INIT_REQUEST.json"
	TAS_INIT_RESPONSE             = "TAS_INIT_RESPONSE.json"
	TAS_NOTIFY                    = "TAS_NOTIFY.json"
	TAS_SUBSCRIBE_REQUEST         = "TAS_SUBSCRIBE_REQUEST.json"
	TAS_SUBSCRIBE_RESPONSE        = "TAS_SUBSCRIBE_RESPONSE.json"
	TAS_TA_REQUEST                = "TAS_TA_REQUEST.json"
	TAS_TA_RESPONSE               = "TAS_TA_RESPONSE.json"
	TAS_TEARDOWN_REQUEST          = "TAS_TEARDOWN_REQUEST.json"
	TAS_TEARDOWN_RESPONSE         = "TAS_TEARDOWN_RESPONSE.json"
	TAS_UNSUBSCRIBE_REQUEST       = "TAS_UNSUBSCRIBE_REQUEST.json"
	TAS_UNSUBSCRIBE_RESPONSE      = "TAS_UNSUBSCRIBE_RESPONSE.json"
	TEST_MESSAGE                  = "TEST_MESSAGE.json"
	V2X_CPM                       = "V2X_CPM.json"
	V2X_NTM                       = "V2X_NTM.json"
)

/*
Function takes a predefined messageSchema and JSON message as string, and either returns the validation result, a list of validation errors, and a general error in case of other problems.
*/
func Validate(messageSchema MessageSchema, message string) (bool, []string, error) {

	schema := gojsonschema.NewReferenceLoader("file://" + projectpath.Root + "/res/schemas/" + string(messageSchema))
	document := gojsonschema.NewStringLoader(message)

	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return false, nil, err
	} else if !result.Valid() {
		errMsgs := make([]string, len(result.Errors()))
		for _, desc := range result.Errors() {
			errMsgs = append(errMsgs, fmt.Sprintf("%s", desc))
		}
		return false, errMsgs, nil
	} else {
		return result.Valid(), nil, nil
	}
}
