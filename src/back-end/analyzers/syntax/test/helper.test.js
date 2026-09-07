import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import test from 'node:test';

function analyze(language, path, content) {
  const result = spawnSync(process.execPath, ['index.js'], {cwd: new URL('..', import.meta.url), input: `${JSON.stringify({schema: 1, language, path, content})}\n`, encoding: 'utf8'});
  assert.equal(result.status, 0, result.stderr);
  return JSON.parse(result.stdout);
}

function symbol(report, name, kind) {
  const found = report.symbols.find((value) => value.name === name && value.kind === kind);
  assert.ok(found, `missing ${kind} ${name}: ${JSON.stringify(report.symbols)}`);
  return found;
}

test('loads and normalizes C and C++ declarations', () => {
  const c = analyze('c', 'sample.c', 'struct Point { int x; }; enum Mode { Fast = 1 }; int add(int x, int y) { return x + y; }');
  assert.equal(c.validSyntax, true);
  assert.equal(symbol(c, 'Point', 'struct').parentId, null);
  assert.equal(symbol(c, 'x', 'field').type, 'int');
  assert.equal(symbol(c, 'Fast', 'enum_member').enumValue, '1');
  assert.deepEqual(symbol(c, 'add', 'function').parameters.map((value) => value.name), ['x', 'y']);

  const cpp = analyze('c++', 'sample.cpp', 'class Counter { public: int value; int add(const int delta) const; };');
  const classSymbol = symbol(cpp, 'Counter', 'class');
  assert.equal(symbol(cpp, 'value', 'field').parentId, classSymbol.id);
  assert.equal(symbol(cpp, 'value', 'field').visibility, 'public');
  assert.deepEqual(symbol(cpp, 'add', 'method').parameters.map((value) => value.name), ['delta']);
});

test('normalizes C#, Java, JavaScript, and TypeScript members', () => {
  const csharp = analyze('csharp', 'sample.cs', '[Serializable] public class C { public int Field; public string Name { get; set; } public void Run(int x) {} } enum E { A = 1 }');
  assert.deepEqual(symbol(csharp, 'C', 'class').attributes, ['[Serializable]']);
  assert.equal(symbol(csharp, 'Field', 'field').visibility, 'public');
  assert.equal(symbol(csharp, 'Name', 'property').type, 'string');
  assert.deepEqual(symbol(csharp, 'Run', 'method').parameters.map((value) => value.name), ['x']);
  assert.equal(symbol(csharp, 'A', 'enum_member').enumValue, '1');

  const java = analyze('java', 'Sample.java', '@Deprecated public class C { private int field; public void run(String x) {} enum E { A } }');
  assert.deepEqual(symbol(java, 'C', 'class').attributes, ['@Deprecated']);
  assert.equal(symbol(java, 'field', 'field').visibility, 'private');
  assert.equal(symbol(java, 'run', 'method').parameters[0].type, 'String');

  const javascript = analyze('js', 'sample.js', 'class C { field = 1; method(x) { return x; } } function fn(a) {}');
  assert.equal(symbol(javascript, 'field', 'field').enumValue, null);
  assert.deepEqual(symbol(javascript, 'method', 'method').parameters.map((value) => value.name), ['x']);
  assert.deepEqual(symbol(javascript, 'fn', 'function').parameters.map((value) => value.name), ['a']);

  const typescript = analyze('ts', 'sample.ts', '@dec export class C { public readonly field: string; method(x: number): void {} } enum E { A = 1 } interface I { foo(x: string): void; }');
  assert.deepEqual(symbol(typescript, 'C', 'class').attributes, ['@dec']);
  assert.deepEqual(symbol(typescript, 'field', 'field').modifiers, ['public', 'readonly']);
  assert.equal(symbol(typescript, 'field', 'field').type, 'string');
  assert.equal(symbol(typescript, 'method', 'method').parameters[0].type, 'number');
  assert.equal(symbol(typescript, 'A', 'enum_member').enumValue, '1');
  assert.ok(symbol(typescript, 'I', 'interface'));

  const tsx = analyze('typescript', 'component.tsx', 'export const Component = () => <div />;');
  assert.equal(tsx.language, 'typescript');
  assert.equal(tsx.validSyntax, true);
  const jsx = analyze('javascript', 'component.jsx', 'const Component = () => <div />;');
  assert.equal(jsx.language, 'javascript');
  assert.equal(jsx.validSyntax, true);
});

test('preserves direct C# metadata and records syntax imports and calls as unresolved', () => {
  const csharp = analyze('csharp', 'state.cs', 'using System.Text; [Attr] public enum State : byte { A = 1, B = 2 } class C<T> { [FieldAttr] protected internal int field; [return:Attr] public virtual T Run([ArgAttr] ref T p) { return p; } void Call() { Run(default(T)); } }');
  const state = symbol(csharp, 'State', 'enum');
  assert.deepEqual(state.attributes, ['[Attr]']);
  assert.equal(state.underlyingType, 'byte');
  assert.deepEqual(symbol(csharp, 'field', 'field').modifiers, ['protected', 'internal']);
  const run = symbol(csharp, 'Run', 'method');
  assert.equal(run.type, 'T');
  assert.deepEqual(run.attributes, ['[return:Attr]']);
  assert.deepEqual(run.parameters[0].attributes, ['[ArgAttr]']);
  assert.deepEqual(run.parameters[0].modifiers, ['ref']);
  assert.equal(csharp.imports[0].path, 'System.Text');
  assert.ok(csharp.references.some((reference) => reference.kind === 'call' && reference.name === 'Run' && reference.resolution === 'unresolved'));

  const typescript = analyze('typescript', 'import.ts', 'import { helper } from "./helper"; export function run() { helper(); }');
  assert.equal(typescript.imports[0].path, './helper');
  assert.ok(typescript.references.some((reference) => reference.kind === 'call' && reference.name === 'helper' && reference.resolution === 'unresolved'));
});

test('uses direct modifier nodes and keeps compound C# visibility', () => {
  const csharp = analyze('csharp', 'visibility.cs', 'class C { [Attr("public async")] protected internal int Shared; private protected int Derived; }');
  const shared = symbol(csharp, 'Shared', 'field');
  assert.deepEqual(shared.modifiers, ['protected', 'internal']);
  assert.equal(shared.visibility, 'protected internal');
  assert.equal(symbol(csharp, 'Derived', 'field').visibility, 'private protected');
  assert.ok(!shared.modifiers.includes('public'));
  assert.ok(!shared.modifiers.includes('async'));
});

test('records the member name, rather than its receiver, for unresolved calls', () => {
  const csharp = analyze('csharp', 'calls.cs', 'class C { void Run(Service service) { service.Save(); } }');
  assert.ok(csharp.references.some((reference) => reference.kind === 'call' && reference.name === 'Save' && reference.resolution === 'unresolved'));
  assert.ok(!csharp.references.some((reference) => reference.kind === 'call' && reference.name === 'service'));

  const javascript = analyze('javascript', 'calls.js', 'class C { run(service) { service.client.execute(); } }');
  assert.ok(javascript.references.some((reference) => reference.kind === 'call' && reference.name === 'execute' && reference.resolution === 'unresolved'));
  assert.ok(!javascript.references.some((reference) => reference.kind === 'call' && reference.name === 'service'));
});

test('reports Tree-sitter missing tokens as syntax diagnostics', () => {
  const typescript = analyze('typescript', 'missing.ts', 'interface I { value: string');
  assert.equal(typescript.validSyntax, false);
  assert.ok(typescript.diagnostics.some((diagnostic) => diagnostic.kind === 'missing_token' && diagnostic.token === '}'));
});

test('returns UTF-8 byte ranges and syntax-only unresolved references', () => {
  const content = '// Привет 😀\nclass C { method() {} }';
  const report = analyze('javascript', 'unicode.js', content);
  const classSymbol = symbol(report, 'C', 'class');
  assert.equal(classSymbol.start, Buffer.byteLength(content.slice(0, content.indexOf('class C')), 'utf8'));
  assert.equal(report.capabilities.semanticResolution, false);
  assert.ok(report.references.every((reference) => reference.resolution === 'unresolved'));
});
