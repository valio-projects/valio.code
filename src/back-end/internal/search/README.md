# Verified search foundation

`Search(ctx, Source, text, Options)` parses and executes a file-level expression.
`Parse` and `Execute` separate those operations. The storage boundary is
`Source.Files(ctx)`, which must return a stable snapshot without silently
limiting candidates. `MemorySource` is useful for tests and local integration.

Engine owns source, options, query parser and index builder. Its Search/Execute
methods perform the work; free functions are compatibility facades. QueryParser
owns a 64 KiB default input budget, and IndexBuilder owns block sizing (16 KiB
default, minimum 12 bytes for safe Unicode overlap). Models, lexer, grammar,
matcher, result mapper and engine live in separate focused files.

The DSL supports uppercase `AND`, `OR`, `NOT`, parentheses, implicit AND,
single/double quoted values, and `/Go RE2 regex/`. Precedence is NOT > AND > OR.
Quoted strings accept escaped quotes, backslash, newline, carriage return and
tab. Regex delimiters can be escaped; regex backslashes otherwise remain intact.
Available filters are `project`, `repo`, `file`, `lang`, `path`, `content`,
`symbol`, `kind`, `test`, `generated`. Unknown filters are errors. Boolean
filters require lowercase `true` or `false`.

Content, path, basename (`file`), and symbol names use substring matching by
default. Project/repository IDs, language, and kind compare whole values.
Options.Mode can select `exact` (the entire field) or `regex`; `/.../` always
selects regex. Case-insensitive matching uses Unicode SimpleFold, consistently
with Go regex; this is not accent folding, normalization, or multi-rune full
Unicode folding. Original CRLF and punctuation are preserved. Results include
half-open original UTF-8 byte ranges, even when a folded rune changes byte size.

`BuildIndex` emits source-aligned 16 KiB blocks overlapping by two runes and
hex-encoded, folded rune trigrams. Every required trigram must occur somewhere
in the file; final verification checks actual contiguity and exact semantics.
Regex required literals come only from guaranteed syntax-tree paths; uncertain
patterns fall back to verification. NOT never subtracts approximate candidates.

Options scopes are checked before scanning. Results retain repository and
snapshot IDs and only the selected project memberships. The source must not
provide duplicate records for the same repository/snapshot/file identity.
Sorting is repository, snapshot, path, file ID. Pagination follows verification
of all scanned files and does not change Total. Complete=false means Total is a
lower bound and Truncated=true. Reaching an output page limit alone does not
make an otherwise exact total incomplete.

## Current scale limits

The initial Source interface materializes the full file list. Search rebuilds
its in-memory candidate index during the request; it does not yet consume
persisted SurrealDB postings. This is a correctness foundation, not a verified
10-million-line throughput implementation. Bounded scans default to 100,000
eligible files and 256 MiB; options can lower or raise those budgets. Short
literals and regexes with no safe required trigram use the same explicit
budgets. A budget stops the deterministic scan at the first unscanned eligible
file and marks totals incomplete. Candidate index construction and a single
file regex are not interruptible mid-file. Streaming storage iteration,
persisted-posting query plans, and measured scale targets remain future work.
