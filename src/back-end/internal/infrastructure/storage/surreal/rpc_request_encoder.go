package surreal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/fxamacker/cbor/v2"
)

// rpcRequestEncoder preserves the existing JSON model contracts while sending
// text as CBOR strings. SurrealDB's legacy JSON RPC parser can otherwise coerce
// source fragments such as "id: number" into record identifiers.
type rpcRequestEncoder struct{}

// Encode keeps RawMessage objects, JSON field tags and exact integer parameters.
// SQL and parameter values remain separate throughout serialization.
func (rpcRequestEncoder) Encode(sql string, parameters map[string]any) ([]byte, error) {
	encoded, err := json.Marshal(map[string]any{"id": "query", "method": "query", "params": []any{sql, parameters}})
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	value, err = rpcNumberValues(value)
	if err != nil {
		return nil, err
	}
	return cbor.Marshal(value)
}

// rpcNumberValues avoids float64 rounding of integral identifiers and counters.
// Out-of-range integers are rejected rather than silently rounded or stringified.
func rpcNumberValues(value any) (any, error) {
	switch typed := value.(type) {
	case json.Number:
		text := string(typed)
		if strings.ContainsAny(text, ".eE") {
			return strconv.ParseFloat(text, 64)
		}
		if strings.HasPrefix(text, "-") {
			return strconv.ParseInt(text, 10, 64)
		}
		return strconv.ParseUint(text, 10, 64)
	case map[string]any:
		for key, child := range typed {
			converted, err := rpcNumberValues(child)
			if err != nil {
				return nil, fmt.Errorf("invalid numeric RPC parameter: %w", err)
			}
			typed[key] = converted
		}
	case []any:
		for index, child := range typed {
			converted, err := rpcNumberValues(child)
			if err != nil {
				return nil, fmt.Errorf("invalid numeric RPC parameter: %w", err)
			}
			typed[index] = converted
		}
	}
	return value, nil
}
