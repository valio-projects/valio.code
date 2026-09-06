package search

type Mode string

const (
	Substring Mode = "substring"
	Exact     Mode = "exact"
	Regex     Mode = "regex"
)
