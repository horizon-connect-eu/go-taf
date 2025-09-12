package validator

import (
	"fmt"
	embedded "github.com/horizon-connect-eu/go-taf"
	"github.com/horizon-connect-eu/go-taf/pkg/message"
	"github.com/xeipuuv/gojsonschema"
)

/*
Function takes a predefined messageSchema and JSON message as string, and either returns the validation result, a list of validation errors, and a general error in case of other problems.
*/
func Validate(messageSchema message.MessageSchema, message string) (bool, []string, error) {

	schemaContent, err := embedded.Schemas.ReadFile("res/schemas/" + string(messageSchema) + ".json")
	if err != nil {
		return false, nil, err
	}
	schema := gojsonschema.NewBytesLoader(schemaContent)
	document := gojsonschema.NewStringLoader(message)

	result, err := gojsonschema.Validate(schema, document)
	if err != nil {
		return false, nil, err
	} else if !result.Valid() {
		errMsgs := make([]string, 0)
		for _, desc := range result.Errors() {
			errMsgs = append(errMsgs, fmt.Sprintf("%s", desc))
		}
		return false, errMsgs, nil
	} else {
		return result.Valid(), nil, nil
	}
}
