# Syntax helper protocol

`node index.js` reads newline-delimited UTF-8 JSON from standard input and writes one
report for each input line. It never executes submitted source. Install dependencies with
`npm ci`, then run `npm test` to load every bundled grammar and exercise the fixtures.

## Input

```json
{"schema":1,"language":"typescript","path":"src/example.ts","content":"export class Example {}"}
```

`language` accepts `c`, `cpp`/`c++`, `csharp`/`c#`, `java`, `javascript`/`js`,
`jsx`, `typescript`/`ts`, and `tsx`. When the caller identifies a file as JavaScript or
TypeScript, `.jsx` and `.tsx` paths select the corresponding grammar while the report
retains the caller's normalized language. The parser uses pinned `web-tree-sitter`
`0.20.8`, which is compatible with the Tree-sitter 0.20 grammar ABI in
`tree-sitter-wasms` `0.1.13`. Newer runtime `0.27.0` cannot load those grammar assets.

## Report contract

Each successful report has `schema: 1`, the normalized language, the supplied path,
`validSyntax`, and `capabilities`. `capabilities.syntax` means a grammar parsed the
content. `capabilities.semanticResolution` is always false: names, calls, and imports
are not compiler-resolved facts.

`symbols` contains syntax declarations with half-open `start` and `end` offsets in the
original UTF-8 byte stream. A symbol has deterministic per-report `id`, optional
`parentId`, `name`, and one of `class`, `interface`, `struct`, `enum`, `enum_member`,
`field`, `property`, `function`, or `method`. `type`, `visibility`, `modifiers`, and
`attributes` are direct grammar observations. Attribute lists retain their literal source
syntax, including attribute targets and arguments; absent or unsupported facts remain
`null`, `unknown`, or an empty list. `parameters` records parameter name, declared type,
modifiers, direct attributes, and UTF-8 byte range. `underlyingType` records an explicit
enum base type when its grammar writes one. `enumValue` contains explicit enum syntax only.

`references` records identifier tokens and syntactic calls with UTF-8 byte ranges and
`resolution: "unresolved"`. C# `using` directives and TypeScript import statements appear
in `imports` as syntax-only unresolved paths. Parse errors appear in `diagnostics` as
`parse_error` with source byte ranges. Invalid input and an unsupported language return
`validSyntax: false`, no symbols, and a stable diagnostic code without returning source
content.

The helper is an AST outline, not a compiler frontend. It does not infer inheritance,
overloads, aliases, call targets, types omitted by syntax, or cross-file relationships.
