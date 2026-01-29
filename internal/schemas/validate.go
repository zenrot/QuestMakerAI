package schemas

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

func ValidateSchema(schema json.RawMessage, input json.RawMessage) (bool, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schema)
	inputLoader := gojsonschema.NewBytesLoader(input)

	result, err := gojsonschema.Validate(schemaLoader, inputLoader)
	if err != nil {
		return false, fmt.Errorf("validateSchema: Validate error: %w", err)
	}
	if result.Valid() {
		return true, nil
	} else {
		return false, nil
	}
}
