package surreal

import (
	"context"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

type rpcRoundTripPayload struct {
	Strings  map[string]string `json:"strings"`
	Artifact json.RawMessage   `json:"artifact"`
	Unsigned uint64            `json:"unsigned"`
	Signed   int64             `json:"signed"`
}

func TestRPCCBORPreservesBoundAndPersistedSourceStrings(t *testing.T) {
	client := embeddingIntegrationClient(t)
	ctx := context.Background()
	payload := rpcRoundTripPayload{
		Strings: map[string]string{
			"record_with_space": "id: number",
			"record":            "person:one",
			"source_like":       "x: y z",
			"uuid_like":         "11111111-1111-1111-1111-111111111111",
			"datetime_like":     "2024-01-01T00:00:00Z",
			"whitespace":        " \tleading and trailing\t ",
			"unicode_crlf":      "Привет 😀\r\nvalue: string",
		},
		Artifact: json.RawMessage(`{"artifact":{"source":"id: number","target":"person:one"},"lines":["x: y z","Привет 😀\r\nvalue: string"]}`),
		Unsigned: uint64(math.MaxInt64),
		Signed:   math.MinInt64,
	}

	rows, err := client.Query(ctx, "RETURN $payload;", map[string]any{"payload": payload})
	if err != nil {
		t.Fatal(err)
	}
	direct, err := DecodeLast[rpcRoundTripPayload](rows)
	if err != nil {
		t.Fatal(err)
	}
	assertRPCRoundTrip(t, payload, direct)

	if err := client.Create(ctx, "rpc_roundtrip", "source", payload); err != nil {
		t.Fatal(err)
	}
	persisted, err := Get[rpcRoundTripPayload](ctx, client, "rpc_roundtrip", "source")
	if err != nil {
		t.Fatal(err)
	}
	assertRPCRoundTrip(t, payload, persisted)
}

func assertRPCRoundTrip(t *testing.T, want, got rpcRoundTripPayload) {
	t.Helper()
	if !reflect.DeepEqual(got.Strings, want.Strings) {
		t.Fatalf("source-like strings changed: got=%q want=%q", got.Strings, want.Strings)
	}
	if got.Unsigned != want.Unsigned || got.Signed != want.Signed {
		t.Fatalf("supported signed-integer range changed: got unsigned=%d signed=%d", got.Unsigned, got.Signed)
	}
	var wantArtifact any
	var gotArtifact any
	if json.Unmarshal(want.Artifact, &wantArtifact) != nil || json.Unmarshal(got.Artifact, &gotArtifact) != nil || !reflect.DeepEqual(gotArtifact, wantArtifact) {
		t.Fatalf("nested artifact/source JSON changed: got=%s want=%s", got.Artifact, want.Artifact)
	}
}
