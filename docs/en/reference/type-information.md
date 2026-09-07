# Type information contract

The type catalog records evidence about one concrete analysis version. It does
not combine matching names into a global type or turn missing compiler data into
zero-valued facts. The initial executable producer is the Go AST reader, with an
optional real Go type-checking pass for a single file. C#, C++, CLR layout and
other language producers are future integrations, not implemented capabilities.

## Identity and lookup

`typeinfo.TypeDescriptor` in `src/back-end/internal/domain/typeinfo` contains its own identity and a `TypeScope` with
`workspaceId`, `projectId`, `buildProfileId` and immutable `versionId`. A type's
simple name and fully qualified name are facts, not identities. Go's current
qualified name is the syntactic package name plus type name; it is not claimed
to be a globally unique module-qualified compiler identity. The Go mapper also
includes repository identity in generated IDs, preventing identical paths in two
repositories from colliding.

`types.NewCatalog(descriptors)` validates and copies the descriptors.
`catalog.ByName(types.QueryScope{...}, name)` is the convenience entry point for
`types.NewByNameHandler(catalog).Handle(types.ByNameQuery{...})`. This read-only
query requires workspace scope and exact,
case-sensitive simple-name or qualified-name matching. Project, build profile
and version filters are optional. The result is `exact`, `ambiguous` or
`not_found`. Ambiguous results retain every candidate and its exact identity;
they never merge fields, methods, overloads or versions. The result's reference
is correspondingly `exact`, `candidate` or `unresolved`. Narrowing the scope
resolves a collision only when one descriptor remains.

Candidate counts describe **recorded** fields, properties, methods,
constructors, parameters, returns, generic parameters, attributes, enum members,
typed constants and occurrences. These are normalized record counts, not a
claim of complete language coverage. For example, zero recorded uses does not
prove that an enum member is unused. Return values and method parameters are
counted separately; receiver parameters are represented separately and are not
included in the ordinary parameter count. Property accessors retain their own
method identities and arguments rather than being merged into unrelated methods.

## Facts, references and evidence

Every scalar metadata fact uses `Fact[T]` with a nullable `value`, a status of
`known`, `unresolved` or `unsupported`, scope, producer/source evidence and an
optional reason. A known fact requires an explicit matching scope, a value and
evidence. Unresolved/unsupported facts must not contain a value. A zero-value
fact serializes as unresolved with a null value, never as known zero. Uncollected
facts inherit the enclosing descriptor's scope if they have no explicit scope.
Known empty strings or empty flag lists can express actual syntax absence.

Symbol references distinguish exact targets, candidate targets and unresolved
names. Written type expressions do not become exact references merely because
their names match a catalog entry. References may retain a target in another
explicit scope. Source occurrences have independent IDs, role facts, symbol
references and repository-relative, half-open ranges with explicit UTF-8 or
UTF-16 position encoding. The current Go mapper converts parser byte offsets to
zero-based UTF-8 line/character ranges.

## Members and language flags

The model represents classes, structs, interfaces, enums, aliases, arrays,
slices, maps, functions, pointers, named types, primitives, delegates and records;
kind strings remain extensible for other language constructs. Each type can
contain fields, properties/indexers, methods, constructors, generic parameters
and constraints. Methods retain individual IDs, signatures, receivers,
parameters and multiple return values. Properties can contain getter/setter
methods and indexer parameters. Each parameter and return has an identity and
position, with its own type and attributes. A `TypeReference` also supports
attributes on the type use itself, including the return type.

Visibility and modifier facts preserve language distinctions. Standard
visibility values include public, private, protected, internal,
protected_internal, private_protected, package, module and unknown. For C#,
`protected_internal` means family **or** assembly access, whereas
`private_protected` means family **and** assembly access. Modifier values include
virtual, override, abstract, final, sealed, static, readonly, const, embedded and
variadic; additional language-specific values can be retained. This vocabulary
is a data contract, not a claim that every producer can determine every flag.

## Attributes and typed constants

Attributes are independently attached to types, fields, properties, methods,
constructors, generic parameters, parameters, returns and type uses. Each
attribute has an identity, a name fact, an attribute-class symbol reference,
ordered positional arguments and named arguments. This can represent C#
attribute classes once a Roslyn producer supplies actual symbol evidence.
Argument values distinguish typed literals, unevaluated expressions and
redacted values. Literal values use exact text plus a type reference, avoiding
JSON number precision loss. Redacted values must not retain their literal or
expression payload and must give a redaction reason.

True enums have an underlying type and members with independent identities,
constant value facts, attributes, declarations and per-member usage occurrences.
Computed values are known only when established by a producer such as a
compiler. An unresolved value is null even when zero would be plausible.
Go typed constant groups are retained under `constants`, with group/member
identities, expressions, computed values when available and separate usage IDs.
They are **not** presented as language enums. Untyped Go constant blocks remain
in the analysis report; the type catalog only attaches constants associated with
a named type.

Go struct tags remain `FieldDescriptor.tag` facts. They are not converted into
C# attributes, runtime annotations or synthetic attribute-class references.

## Arrays and physical layout

Array shapes carry an element type, rank and per-dimension lengths/lower bounds,
each with its own fact status. Current Go nested fixed arrays are exposed as
their sequence of nested dimensions; this does not imply CLR multidimensional
array identity. Go array lower bounds are known zero from the language's array
semantics. Lengths become known only when supplied by the Go checker; unresolved
expressions stay unknown. Slices are identified as slices and are not assigned a
fixed array length.

Layout is target-specific: architecture, ABI, compiler and compiler version,
build-profile scope, kind (`auto`, `sequential`, `explicit`, `unknown`), total
size, alignment, packing and per-field byte/bit offsets and widths. All scalar
layout data carries the same fact/evidence contract. Physical size, alignment,
packing or offset facts require known architecture, ABI, compiler and version.
Missing layout is unresolved and null; it is never zero or an estimate based on
the machine running the catalog.

The Go mapper deliberately leaves layout unresolved. Future Roslyn/CLR and
Clang producers must supply actual compiler/runtime evidence for the exact
target and build profile. C# declaration order, `StructLayout`/`FieldOffset`
attribute syntax, host pointer width or a familiar ABI do not by themselves
prove final managed layout, object size, offsets or alignment. Declared layout
intent can be recorded as source evidence without claiming computed layout.

## Current coverage and boundaries

The Go reader currently extracts package-level named types and aliases, struct
fields and embedded fields, tags, interface method declarations, receiver
methods declared in the same analyzed file, generic parameters/constraint
expressions, parameter and return signatures, written underlying types and
fixed-array expressions. Go exported/package visibility and syntax modifiers
are reported directly. Go has no C#-style properties, constructors, enums or
attribute declarations; the mapper does not fabricate these records.

The optional `go/types` pass can establish computed constants, fixed-array
lengths and references mapped to local definition objects. It checks the one
supplied file; it is not a module-wide compiler pipeline. Diagnostics and
partial-check status remain in the analysis report. Imported/unavailable
semantic targets, cross-file method attachment, ABI layout, runtime attributes
and cross-language inheritance/member resolution are not claimed by this
catalog. Unsupported future features need real language producers and evidence.

## Package boundaries

Focused files in the domain-only `src/back-end/internal/domain/typeinfo` package define facts, references, kinds, attributes, members,
constants, layouts and type descriptors. `src/back-end/internal/analysis` owns Go parsing and
optional checking. `types.GoMapper` adds workspace/project/build-profile/version
scope, repository-scoped syntax IDs and source evidence to an `analysis.Report`.
These syntax IDs are local to the pinned analysis version; they do not claim
persistent compiler identity across file moves or declaration-offset changes.
`types.Catalog` handles validated immutable copies, deterministic lookup and
record counts. No SurrealDB adapter, API, frontend or compiler service is
implemented in this package.
