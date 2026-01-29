package schemas

import (
	_ "embed"
	"encoding/json"
)

//go:embed questions_schema.json
var QuestionSchema json.RawMessage
