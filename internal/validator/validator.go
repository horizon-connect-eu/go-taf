package validator

import (
	"fmt"
	"github.com/vs-uulm/go-taf/internal/projectpath"
	"github.com/vs-uulm/go-taf/pkg/message"
	"github.com/xeipuuv/gojsonschema"
)

/*
Function takes a predefined messageSchema and JSON message as string, and either returns the validation result, a list of validation errors, and a general error in case of other problems.
*/
func Validate(messageSchema message.MessageSchema, message string) (bool, []string, error) {

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
