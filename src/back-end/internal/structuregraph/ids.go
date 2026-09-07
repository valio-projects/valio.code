package structuregraph

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
)

func graphID(parts ...string) string {
	hash := sha256.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return "sg-" + hex.EncodeToString(hash.Sum(nil)[:16])
}

func sourceNodeID(kind NodeKind, sourceID, reportID string, start, end int) string {
	return graphID("node", string(kind), sourceID, reportID, strconv.Itoa(start), strconv.Itoa(end))
}

func relationID(kind RelationKind, sourceID, targetID string, resolution Resolution, rangeValue ByteRange) string {
	return graphID("edge", string(kind), sourceID, targetID, string(resolution), rangeValue.FileID, strconv.Itoa(rangeValue.Start), strconv.Itoa(rangeValue.End))
}
