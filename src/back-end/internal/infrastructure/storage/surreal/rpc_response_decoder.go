package surreal

import (
	"encoding/json"
	"github.com/fxamacker/cbor/v2"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"github.com/surrealdb/surrealdb.go/surrealcbor"
)

// rpcResponseDecoder delegates SurrealDB's CBOR tags (including NONE, records
// and timestamps) to the official SDK, then retains existing JSON read models.
type rpcResponseDecoder struct{}

// Decode validates a typed RPC response before adapting it to JSON contracts.
func (rpcResponseDecoder) Decode(data []byte, target any) error {
	// Validate nesting and trailing data before the SDK's recursive tag decoder.
	if err := cbor.Wellformed(data); err != nil {
		return err
	}
	var value any
	if err := surrealcbor.Unmarshal(data, &value); err != nil {
		return err
	}
	encoded, err := json.Marshal(rpcJSONValues(value))
	if err != nil {
		return err
	}
	return json.Unmarshal(encoded, target)
}

// rpcJSONValues retains the former JSON-facing record representation for query
// metadata and pre-CBOR records. Ordinary source strings never enter this case.
func rpcJSONValues(value any) any {
	switch typed := value.(type) {
	case models.RecordID:
		return typed.String()
	case *models.RecordID:
		if typed == nil {
			return nil
		}
		return typed.String()
	case models.CustomDuration:
		return typed.Duration.String()
	case map[string]any:
		for key, child := range typed {
			typed[key] = rpcJSONValues(child)
		}
	case []any:
		for index, child := range typed {
			typed[index] = rpcJSONValues(child)
		}
	}
	return value
}
