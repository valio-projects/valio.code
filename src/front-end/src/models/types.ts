export interface TypeScope {
  workspaceId: string;
  projectId: string;
  buildProfileId: string;
  versionId: string;
}
export interface SourceRange {
  fileId?: string;
  path?: string;
  startByte?: number;
  endByte?: number;
  start?: number;
  end?: number;
  [key: string]: unknown;
}
export interface Evidence {
  id: string;
  producer: string;
  version?: string;
  range?: SourceRange;
  [key: string]: unknown;
}
export interface Fact<T> {
  status: "known" | "unresolved" | "unsupported";
  value?: T;
  reason?: string;
  evidence?: Evidence[];
  scope?: TypeScope;
}
export interface SymbolReference {
  resolution: string;
  exact?: { id: string; scope: TypeScope };
  candidates?: { id: string; scope: TypeScope }[];
  reason?: string;
}
export interface Occurrence {
  id: string;
  range: SourceRange;
  role: Fact<string>;
  symbol: SymbolReference;
}
export interface TypeReference {
  name: Fact<string>;
  symbol: SymbolReference;
  attributes?: AttributeUse[];
  array?: ArrayShape;
}
export interface AttributeValue {
  kind: string;
  type: TypeReference;
  literal: Fact<string>;
  expression: Fact<string>;
  redactionReason?: string;
}
export interface AttributeUse {
  id: string;
  name: Fact<string>;
  attributeClass: SymbolReference;
  positionalArguments?: AttributeValue[];
  namedArguments?: { name: Fact<string>; value: AttributeValue }[];
  occurrence?: Occurrence;
}
export interface Parameter {
  id: string;
  name: Fact<string>;
  position: Fact<number>;
  type: TypeReference;
  modifiers?: Fact<string[]>;
  defaultValue?: Fact<string>;
  attributes?: AttributeUse[];
}
export interface Method {
  id: string;
  name: Fact<string>;
  signature: Fact<string>;
  visibility: Fact<string>;
  modifiers: Fact<string[]>;
  receiver?: Parameter;
  parameters?: Parameter[];
  returns?: Parameter[];
  genericParameters?: GenericParameter[];
  attributes?: AttributeUse[];
  declaration?: Occurrence;
}
export interface Field {
  id: string;
  name: Fact<string>;
  type: TypeReference;
  visibility: Fact<string>;
  modifiers: Fact<string[]>;
  tag?: Fact<string>;
  attributes?: AttributeUse[];
  declaration?: Occurrence;
}
export interface Property extends Field {
  parameters?: Parameter[];
  getter?: Method;
  setter?: Method;
}
export interface GenericParameter {
  id: string;
  name: Fact<string>;
  constraints?: TypeReference[];
  attributes?: AttributeUse[];
}
export interface Constant {
  id: string;
  groupId?: string;
  name: Fact<string>;
  expression: Fact<string>;
  value: { type: TypeReference; value: Fact<string> };
  attributes?: AttributeUse[];
  occurrences?: Occurrence[];
  declaration?: Occurrence;
}
export interface ArrayShape {
  elementType: TypeReference;
  rank: Fact<number>;
  dimensions?: { length: Fact<number>; lowerBound: Fact<number> }[];
}
export interface TypeLayout {
  targetArchitecture: Fact<string>;
  abi: Fact<string>;
  compiler: Fact<string>;
  compilerVersion: Fact<string>;
  kind: Fact<string>;
  sizeBytes: Fact<number>;
  alignmentBytes: Fact<number>;
  packingBytes: Fact<number>;
  fields?: {
    fieldId: string;
    offsetBytes: Fact<number>;
    bitOffset: Fact<number>;
    bitWidth: Fact<number>;
  }[];
  unknownReasons?: string[];
}
export interface TypeDescriptor {
  id: string;
  scope: TypeScope;
  name: Fact<string>;
  fullyQualifiedName: Fact<string>;
  language: Fact<string>;
  kind: Fact<string>;
  visibility: Fact<string>;
  modifiers: Fact<string[]>;
  symbol: SymbolReference;
  declaration?: Occurrence;
  fields?: Field[];
  properties?: Property[];
  methods?: Method[];
  constructors?: Method[];
  genericParameters?: GenericParameter[];
  attributes?: AttributeUse[];
  underlyingType?: TypeReference;
  enum?: { underlyingType: TypeReference; members: Constant[] };
  constants?: Constant[];
  embeddedTypes?: TypeReference[];
  array?: ArrayShape;
  layout: TypeLayout;
}
export interface TypeCandidate {
  type: TypeDescriptor;
  counts: Record<string, number>;
  reference: SymbolReference;
}
export interface TypeResult {
  viewId: string;
  status: "exact" | "ambiguous" | "not_found";
  candidates: TypeCandidate[];
  reference?: SymbolReference;
}
