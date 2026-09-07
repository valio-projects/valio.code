package surreal

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/fxamacker/cbor/v2"
)

func TestRPCEncodingPreservesJSONContractsAndNumbers(t *testing.T) {
	parameters := map[string]any{"source": "id: number", "counter": int64(math.MaxInt64), "negative": int64(math.MinInt64), "weight": 0.25, "artifact": json.RawMessage(`{"name":"person:one","items":[1,null]}`)}
	data, err := (rpcRequestEncoder{}).Encode("RETURN $source;", parameters)
	if err != nil {
		t.Fatal(err)
	}
	var wire struct {
		Params []any `cbor:"params"`
	}
	mode, err := (cbor.DecOptions{DefaultMapType: nil}).DecMode()
	if err != nil {
		t.Fatal(err)
	}
	if err := mode.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	values := wire.Params[1].(map[any]any)
	if wire.Params[0] != "RETURN $source;" || values["source"] != "id: number" || values["counter"] != uint64(math.MaxInt64) || values["negative"] != int64(math.MinInt64) || values["weight"] != 0.25 {
		t.Fatalf("contract changed: %#v", wire)
	}
	if _, ok := values["artifact"].(map[any]any); !ok {
		t.Fatal("RawMessage became an opaque byte string")
	}
	if _, err := (rpcRequestEncoder{}).Encode("RETURN 1;", map[string]any{"bad": json.Number("18446744073709551616")}); err == nil {
		t.Fatal("overflowing integer accepted")
	}
}

func TestRPCResponseUsesSDKTagsAndRejectsTrailingData(t *testing.T) {
	data, err := cbor.Marshal(map[string]any{"record": cbor.Tag{Number: 8, Content: []any{"person", "one"}}, "none": cbor.Tag{Number: 6, Content: nil}, "source": "id: number"})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := (rpcResponseDecoder{}).Decode(data, &got); err != nil {
		t.Fatal(err)
	}
	if got["record"] != "person:one" || got["none"] != nil || got["source"] != "id: number" {
		t.Fatalf("unexpected tag projection: %#v", got)
	}
	if err := (rpcResponseDecoder{}).Decode(append(data, 0xf6), &got); err == nil {
		t.Fatal("trailing response accepted")
	}
}
