# Go syntax and local compiler evidence

`Parser.Analyze(path, content)` reports declarations, identifiers, imports and
syntax diagnostics using Go's parser. Ranges are half-open original UTF-8 byte
offsets. Invalid input returns partial evidence and validSyntax=false. Analyze
and AnalyzeWithOptions are compatibility facades.

Separate files hold each model, syntax readers, type-expression/type/method/
constant parsers, and GoTypeChecker. SyntaxProvider and CompilerProvider isolate
go/parser, go/types and go/importer. StandardGoCompiler accepts an injectable
module-aware importer. No custom compiler semantics are substituted.

Report.Types includes package-level types, aliases, underlying source type
expressions, visibility, generics, struct fields (including embedded fields),
Go struct tags, declared receiver methods, and interface methods/embeddings.
Signatures preserve pointer/generic receivers, named or unnamed parameters,
variadics and multiple returns. Report.Functions includes ordinary functions
and methods whose receiver declaration is absent in this file. Local types stay
in the syntax outline; promoted method sets and aliases' fields are not expanded.

Report.ConstantGroups preserves each Go const block and member identity. Typed
members also appear on their named type. Expression inheritance and iota syntax
are retained. These are Go const groups, never language enums. Go struct tags
are not C# attributes. Go properties, constructor declarations, or attribute
lists are not invented.

By default uses remain unresolved. Options.CheckTypes runs actual single-file
Go checking. TypeCheckStatus is not-run, complete, or partial; TypeDiagnostics
is separate from syntax diagnostics. A local use resolves only when checker
object identity maps to a local declaration. Typed constant values and per-value
references use that evidence, keeping shadowed same-name locals distinct.
Missing and external references remain unresolved. Numeric array dimensions
require checker evidence; syntax-only dimensions keep expressions and unknown
lengths. Invalid negative/string/bool dimensions remain unresolved.

Sibling files, build tags and module dependencies are not loaded automatically.
A partial check never claims a valid package. Capabilities.TypeChecked is true
only after success; comprehensive reference/call/flow capabilities remain false.
No SCIP, external reference graph, call graph or control/data flow is implemented.
Implicit receiver type-parameter definitions are not comprehensively outlined.

Size, alignment, field offsets, padding and ABI remain unknown, even after local
checking, because no compiler target/build profile was supplied. Outline IDs
are local; storage/domain mapping supplies project, repository, profile and
version identity.
