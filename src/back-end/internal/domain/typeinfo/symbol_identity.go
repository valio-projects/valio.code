package typeinfo

// SymbolIdentity identifies a symbol within one immutable analysis scope.
type SymbolIdentity struct {
	// ID is the producer's stable symbol identifier.
	ID string `json:"id"`
	// Scope prevents equivalently named symbols from different inputs colliding.
	Scope TypeScope `json:"scope"`
}
