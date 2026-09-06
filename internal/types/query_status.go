package types

type QueryStatus string

const (
	MatchExact     QueryStatus = "exact"
	MatchAmbiguous QueryStatus = "ambiguous"
	MatchNotFound  QueryStatus = "not_found"
)
