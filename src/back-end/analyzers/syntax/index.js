import readline from 'node:readline';
import {fileURLToPath} from 'node:url';
import Parser from 'web-tree-sitter';

const grammar = {c: 'tree-sitter-c.wasm', cpp: 'tree-sitter-cpp.wasm', csharp: 'tree-sitter-c_sharp.wasm', java: 'tree-sitter-java.wasm', javascript: 'tree-sitter-javascript.wasm', jsx: 'tree-sitter-javascript.wasm', typescript: 'tree-sitter-typescript.wasm', tsx: 'tree-sitter-tsx.wasm'};
const aliases = {cxx: 'cpp', 'c++': 'cpp', cs: 'csharp', 'c#': 'csharp', js: 'javascript', ts: 'typescript'};
const grammarRoot = new URL('./node_modules/tree-sitter-wasms/out/', import.meta.url);
const languageCache = new Map();

await Parser.init();

class InputError extends Error {
  constructor(code) { super(code); this.code = code; }
}

async function language(name, path) {
  const normalized = aliases[name] || name;
  let selected = normalized;
  if (normalized === 'javascript' && /\.jsx$/i.test(path)) selected = 'jsx';
  if (normalized === 'typescript' && /\.tsx$/i.test(path)) selected = 'tsx';
  const file = grammar[selected];
  if (!file) throw new InputError('unsupported_language');
  if (!languageCache.has(selected)) languageCache.set(selected, Parser.Language.load(fileURLToPath(new URL(file, grammarRoot))));
  return languageCache.get(selected);
}
function utf8Offset(content, utf16Index) { return Buffer.byteLength(content.slice(0, utf16Index), 'utf8'); }
function text(content, node) { return content.slice(node.startIndex, node.endIndex); }
function descendants(node, predicate, out = []) { for (const child of node.namedChildren) { if (predicate(child)) out.push(child); descendants(child, predicate, out); } return out; }
function firstDescendant(node, predicate) { for (const child of node.namedChildren) { if (predicate(child)) return child; const match = firstDescendant(child, predicate); if (match) return match; } return null; }
function field(node, ...names) { for (const name of names) { const value = node.childForFieldName(name); if (value) return value; } return null; }

const modifierWords = new Set(['public', 'private', 'protected', 'internal', 'static', 'abstract', 'virtual', 'override', 'sealed', 'readonly', 'const', 'async', 'export', 'default', 'extern', 'unsafe', 'partial', 'final', 'synchronized', 'ref', 'out', 'in', 'params', 'this']);
const modifierContainers = new Set(['modifier', 'modifiers']);
const attributeNodes = new Set(['attribute_list', 'annotation', 'marker_annotation', 'decorator']);
const identifierNodes = new Set(['identifier', 'property_identifier', 'field_identifier']);

function declarationKind(node, insideType) {
  switch (node.type) {
    case 'class_declaration': case 'class_specifier': return 'class';
    case 'interface_declaration': return 'interface';
    case 'struct_declaration': case 'struct_specifier': return 'struct';
    case 'enum_declaration': case 'enum_specifier': return 'enum';
    case 'enumerator': case 'enum_member_declaration': case 'enum_constant': case 'enum_assignment': return 'enum_member';
    case 'property_declaration': case 'property_signature': return 'property';
    case 'method_declaration': case 'method_signature': case 'method_definition': case 'constructor_declaration': return 'method';
    case 'function_definition': case 'function_declaration': return insideType ? 'method' : 'function';
    case 'field_definition': case 'public_field_definition': return 'field';
    case 'field_declaration': {
      const callable = !!firstDescendant(node, (child) => child.type === 'function_declarator');
      return callable ? (insideType ? 'method' : 'function') : 'field';
    }
    case 'declaration': return firstDescendant(node, (child) => child.type === 'function_declarator') ? (insideType ? 'method' : 'function') : 'field';
    default: return '';
  }
}
function nameNode(node) {
  if (['identifier', 'property_identifier', 'field_identifier', 'type_identifier'].includes(node.type)) return node;
  const explicit = field(node, 'name', 'property', 'declarator', 'pattern');
  if (explicit) {
    if (['identifier', 'property_identifier', 'field_identifier', 'type_identifier'].includes(explicit.type)) return explicit;
    const nested = firstDescendant(explicit, (child) => ['identifier', 'property_identifier', 'field_identifier', 'type_identifier'].includes(child.type));
    if (nested) return nested;
  }
  const declarator = firstDescendant(node, (child) => ['function_declarator', 'variable_declarator', 'init_declarator', 'pointer_declarator', 'reference_declarator'].includes(child.type));
  if (declarator) return firstDescendant(declarator, (child) => ['identifier', 'property_identifier', 'field_identifier'].includes(child.type));
  return firstDescendant(node, (child) => ['identifier', 'property_identifier', 'field_identifier', 'type_identifier'].includes(child.type));
}
function typeText(content, node) {
  const type = field(node, 'type', 'return_type');
  if (type) return type.type === 'type_annotation' ? text(content, type).replace(/^:\s*/, '').trim() : text(content, type).trim();
  const annotation = node.namedChildren.find((child) => child.type === 'type_annotation');
  if (annotation) return text(content, annotation).replace(/^:\s*/, '').trim();
  const declaration = node.namedChildren.find((child) => child.type === 'variable_declaration');
  if (declaration) return typeText(content, declaration);
  return null;
}
function modifierTokens(content, node, name) {
  const values = [];
  const collect = (parent, boundary) => {
    for (const child of parent.children) {
      if (child.startIndex >= boundary || attributeNodes.has(child.type)) continue;
      const value = text(content, child).trim();
      if (modifierWords.has(value)) values.push(value);
      if (modifierContainers.has(child.type)) collect(child, boundary);
    }
  };
  const boundary = name ? name.startIndex : node.endIndex;
  collect(node, boundary);
  // Some grammars put `export` around a declaration. Only inspect the wrapper's
  // direct tokens before the declaration; never scan arbitrary prefix source.
  let wrapper = node.parent;
  while (wrapper && ['export_statement', 'declaration'].includes(wrapper.type)) {
    collect(wrapper, node.startIndex);
    wrapper = wrapper.parent;
  }
  return [...new Set(values)];
}
function attributes(content, node) {
  const candidates = [];
  for (const child of node.namedChildren) {
    if (child.type === 'attribute_list') candidates.push(child);
    if (['annotation', 'marker_annotation', 'decorator'].includes(child.type)) candidates.push(child);
    if (child.type === 'modifiers') candidates.push(...descendants(child, (value) => ['annotation', 'marker_annotation'].includes(value.type)));
  }
  let wrapper = node.parent;
  while (wrapper && ['export_statement', 'declaration'].includes(wrapper.type)) {
    candidates.push(...wrapper.namedChildren.filter((child) => ['attribute_list', 'annotation', 'marker_annotation', 'decorator'].includes(child.type)));
    wrapper = wrapper.parent;
  }
  return [...new Set(candidates.map((candidate) => text(content, candidate).trim()).filter(Boolean))];
}
function enclosingAccess(content, node) {
  const siblings = node.parent?.namedChildren || [];
  for (let index = siblings.length - 1; index >= 0; index--) {
    if (siblings[index].endIndex <= node.startIndex && siblings[index].type === 'access_specifier') {
      const access = text(content, siblings[index]).replace(/:$/, '').trim();
      return ['public', 'private', 'protected'].includes(access) ? access : null;
    }
  }
  return null;
}
function parameters(content, node) {
  const declarator = field(node, 'declarator');
  const list = field(node, 'parameters') || node.namedChildren.find((child) => ['formal_parameters', 'parameter_list'].includes(child.type)) || (declarator && firstDescendant(declarator, (child) => ['formal_parameters', 'parameter_list'].includes(child.type)));
  if (!list) return [];
  const values = [];
  for (const parameter of list.namedChildren) {
    const name = nameNode(parameter);
    if (!name) continue;
    values.push({name: text(content, name), type: typeText(content, parameter), modifiers: modifierTokens(content, parameter, name), attributes: attributes(content, parameter), start: utf8Offset(content, parameter.startIndex), end: utf8Offset(content, parameter.endIndex)});
  }
  return values;
}
function symbol(content, node, kind, parentId, id) {
  const name = nameNode(node);
  const modifiers = modifierTokens(content, node, name);
  const access = enclosingAccess(content, node);
  if (access && !modifiers.includes(access)) modifiers.unshift(access);
  const visibility = modifiers.includes('private') && modifiers.includes('protected') ? 'private protected' : modifiers.includes('protected') && modifiers.includes('internal') ? 'protected internal' : modifiers.find((modifier) => ['public', 'private', 'protected', 'internal'].includes(modifier)) || access || 'unknown';
  const value = kind === 'enum_member' ? field(node, 'value') : null;
  const bases = kind === 'enum' ? field(node, 'bases') : null;
  return {id, name: name ? text(content, name) : '', kind, start: utf8Offset(content, node.startIndex), end: utf8Offset(content, node.endIndex), parentId, type: typeText(content, node), underlyingType: bases ? text(content, bases).replace(/^:\s*/, '').trim() : null, visibility, modifiers, attributes: attributes(content, node), parameters: parameters(content, node), enumValue: value ? text(content, value).trim() : null};
}
function importPath(content, node) {
  const source = field(node, 'source');
  if (source) return text(content, source).replace(/^(?:'|")|(?:'|")$/g, '');
  return text(content, node).replace(/^using\s+/, '').replace(/;\s*$/, '').trim();
}
function callTarget(node) {
  const target = field(node, 'function', 'name');
  if (!target) return null;
  if (identifierNodes.has(target.type)) return target;
  // Member calls retain only their syntactic member name. The receiver is not a
  // call target and is deliberately left unresolved: this helper has no type or
  // dispatch semantics.
  const member = field(target, 'name', 'property', 'member', 'field');
  if (member && identifierNodes.has(member.type)) return member;
  let last = null;
  const visit = (value) => {
    for (const child of value.namedChildren) {
      if (identifierNodes.has(child.type)) last = child;
      visit(child);
    }
  };
  visit(target);
  return last;
}

function missingNodes(node, values = []) {
  if (node.isMissing()) values.push(node);
  for (const child of node.children) missingNodes(child, values);
  return values;
}
function reportTree(input, tree) {
  const symbols = [], references = [], imports = [], diagnostics = [];
  let serial = 0;
  function walk(node, parentId, insideType) {
    const kind = declarationKind(node, insideType);
    const id = kind ? `s${++serial}` : parentId;
    if (kind) symbols.push(symbol(input.content, node, kind, parentId, id));
    if (['import_statement', 'using_directive'].includes(node.type)) {
      imports.push({path: importPath(input.content, node), start: utf8Offset(input.content, node.startIndex), end: utf8Offset(input.content, node.endIndex), kind: 'import', resolution: 'unresolved'});
    }
    if (['call_expression', 'invocation_expression'].includes(node.type)) {
      const target = callTarget(node);
      if (target) references.push({name: text(input.content, target), start: utf8Offset(input.content, target.startIndex), end: utf8Offset(input.content, target.endIndex), kind: 'call', resolution: 'unresolved'});
    }
    if (node.type === 'identifier') references.push({name: text(input.content, node), start: utf8Offset(input.content, node.startIndex), end: utf8Offset(input.content, node.endIndex), kind: 'identifier', resolution: 'unresolved'});
    const childInsideType = insideType || ['class', 'interface', 'struct'].includes(kind);
    for (const child of node.namedChildren) walk(child, id, childInsideType);
  }
  walk(tree.rootNode, null, false);
  for (const node of tree.rootNode.descendantsOfType('ERROR')) diagnostics.push({kind: 'parse_error', start: utf8Offset(input.content, node.startIndex), end: utf8Offset(input.content, node.endIndex)});
  for (const node of missingNodes(tree.rootNode)) diagnostics.push({kind: 'missing_token', token: node.type, start: utf8Offset(input.content, node.startIndex), end: utf8Offset(input.content, node.endIndex)});
  return {schema: 1, language: aliases[input.language] || input.language, path: input.path, validSyntax: diagnostics.length === 0, symbols, references, imports, diagnostics, capabilities: {syntax: true, semanticResolution: false}};
}
async function report(input) {
  if (!input || input.schema !== 1 || typeof input.language !== 'string' || typeof input.path !== 'string' || typeof input.content !== 'string') throw new InputError('invalid_input');
  const parser = new Parser();
  try {
    parser.setLanguage(await language(input.language, input.path));
    const tree = parser.parse(input.content);
    try { return reportTree(input, tree); } finally { tree.delete(); }
  } finally { parser.delete(); }
}
function failure(input, error) {
  const code = error instanceof InputError ? error.code : 'parser_unavailable';
  return {schema: 1, language: typeof input?.language === 'string' ? input.language : undefined, path: typeof input?.path === 'string' ? input.path : undefined, validSyntax: false, symbols: [], references: [], imports: [], diagnostics: [{kind: 'invalid_input', code}], capabilities: {syntax: false, semanticResolution: false}};
}
const lines = readline.createInterface({input: process.stdin, crlfDelay: Infinity});
for await (const line of lines) {
  let input;
  try { input = JSON.parse(line); process.stdout.write(`${JSON.stringify(await report(input))}\n`); }
  catch (error) { process.stdout.write(`${JSON.stringify(failure(input, error))}\n`); }
}
