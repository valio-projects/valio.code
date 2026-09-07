# Structural candidate fingerprints

`Fingerprint(source)` parses Go and computes a SHA-256 of its AST shape using
the versioned `go-ast-shape-v1` scheme. Invalid syntax is rejected. Comments,
positions, and literal values are ignored. Literal token kinds, operators,
control-node kinds, declaration structure, and identifier spelling remain.

There is no binding renaming, algebraic equivalence, type resolution, effect
analysis, control/data flow, cross-language normalization, or semantic proof.
Every result says `evidence=syntax-shape` and `equivalence=unknown`.
Equal fingerprints are candidates for inspection: `return 1` and `return 999`
intentionally match, despite different behavior. Renaming a local variable can
change the fingerprint. Scheme versions must not be mixed in a single index.
