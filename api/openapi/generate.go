// Command openapi generates the HTTP contract from the application's Go DTOs.
// Run: go run ./api/openapi
package main

import (
	"encoding/json"
	"github.com/valio-projects/valio.code/internal/application/queries"
	"github.com/valio-projects/valio.code/internal/application/snapshots"
	"github.com/valio-projects/valio.code/internal/domain"
	"github.com/valio-projects/valio.code/internal/projects"
	"github.com/valio-projects/valio.code/internal/search"
	"os"
	"reflect"
	"regexp"
	"strings"
)

type schemaBuilder struct{ definitions map[string]any }

var invalidName = regexp.MustCompile(`[^A-Za-z0-9._-]`)

func (b *schemaBuilder) schema(t reflect.Type) map[string]any {
	if t.Kind() == reflect.Pointer {
		return map[string]any{"anyOf": []any{b.schema(t.Elem()), map[string]string{"type": "null"}}}
	}
	if t.Name() != "" && t.Kind() == reflect.Struct {
		name := invalidName.ReplaceAllString(t.String(), "_")
		if _, ok := b.definitions[name]; !ok {
			b.definitions[name] = map[string]any{}
			b.definitions[name] = b.object(t)
		}
		return map[string]any{"$ref": "#/components/schemas/" + name}
	}
	switch t.Kind() {
	case reflect.Struct:
		return b.object(t)
	case reflect.String:
		return map[string]any{"type": "string"}
	case reflect.Bool:
		return map[string]any{"type": "boolean"}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return map[string]any{"type": "integer"}
	case reflect.Float32, reflect.Float64:
		return map[string]any{"type": "number"}
	case reflect.Slice, reflect.Array:
		return map[string]any{"type": []string{"array", "null"}, "items": b.schema(t.Elem())}
	case reflect.Map:
		return map[string]any{"type": "object", "additionalProperties": b.schema(t.Elem())}
	default:
		return map[string]any{}
	}
}
func (b *schemaBuilder) object(t reflect.Type) map[string]any {
	props := map[string]any{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if tag == "-" {
			continue
		}
		if f.Anonymous && tag == "" {
			embedded := b.object(f.Type)
			for k, v := range embedded["properties"].(map[string]any) {
				props[k] = v
			}
			continue
		}
		if tag == "" {
			tag = f.Name
		}
		props[tag] = b.schema(f.Type)
	}
	return map[string]any{"type": "object", "properties": props, "additionalProperties": false}
}
func main() {
	b := schemaBuilder{definitions: map[string]any{}}
	schema := func(v any) any { return b.schema(reflect.TypeOf(v)) }
	paths := map[string]any{}
	operation := func(path, method, id string, request, response any, parameters []any, public bool) {
		responses := map[string]any{"200": map[string]any{"description": "Successful response", "content": map[string]any{"application/json": map[string]any{"schema": response}}}}
		for _, code := range []string{"400", "401", "403", "404", "409", "413", "415", "500"} {
			responses[code] = map[string]any{"description": "Sanitized error; SCOPE_TOO_LARGE is never a partial success", "content": map[string]any{"application/json": map[string]any{"schema": map[string]any{"$ref": "#/components/schemas/Error"}}}}
		}
		op := map[string]any{"operationId": id, "responses": responses}
		if parameters != nil {
			op["parameters"] = parameters
		}
		if request != nil {
			op["requestBody"] = map[string]any{"required": true, "content": map[string]any{"application/json": map[string]any{"schema": request}}}
		}
		if public {
			op["security"] = []any{}
		}
		if paths[path] == nil {
			paths[path] = map[string]any{}
		}
		paths[path].(map[string]any)[method] = op
	}
	parameter := func(name, where string, required bool) any {
		return map[string]any{"name": name, "in": where, "required": required, "schema": map[string]string{"type": "string"}}
	}
	list := func(item any) any {
		return map[string]any{"type": "object", "properties": map[string]any{"items": map[string]any{"type": "array", "items": item}}}
	}
	operation("/healthz", "get", "health", nil, map[string]any{"type": "object"}, nil, true)
	operation("/readyz", "get", "ready", nil, map[string]any{"type": "object"}, nil, true)
	operation("/api/v1/session", "post", "createSession", map[string]any{"type": "object", "required": []string{"token"}, "additionalProperties": false, "properties": map[string]any{"token": map[string]any{"type": "string", "minLength": 32, "writeOnly": true}}}, map[string]any{"type": "object"}, []any{parameter("Origin", "header", true)}, true)
	operation("/api/v1/session", "delete", "deleteSession", nil, map[string]any{"type": "object"}, nil, false)
	operation("/api/v1/workspace", "get", "getWorkspace", nil, schema(domain.Workspace{}), nil, false)
	operation("/api/v1/projects", "get", "listProjects", nil, list(schema(projects.Definition{})), nil, false)
	operation("/api/v1/projects", "post", "createProject", schema(projects.Definition{}), schema(projects.Definition{}), nil, false)
	operation("/api/v1/projects/{id}", "get", "getProject", nil, schema(projects.Definition{}), []any{parameter("id", "path", true)}, false)
	operation("/api/v1/projects/{id}", "put", "updateProject", schema(projects.Definition{}), schema(projects.Definition{}), []any{parameter("id", "path", true)}, false)
	operation("/api/v1/repositories", "get", "listRepositories", nil, list(schema(domain.Repository{})), nil, false)
	operation("/api/v1/repositories", "post", "registerRepository", schema(domain.Repository{}), schema(domain.Repository{}), nil, false)
	operation("/api/v1/ingestion", "post", "ingestSnapshot", schema(snapshots.IngestCommand{}), schema(snapshots.IngestResult{}), nil, false)
	operation("/api/v1/views", "get", "getLatestView", nil, schema(snapshots.View{}), nil, false)
	operation("/api/v1/views/{id}", "get", "getView", nil, schema(snapshots.View{}), []any{parameter("id", "path", true)}, false)
	operation("/api/v1/search", "post", "searchCode", schema(queries.SearchQuery{}), schema(queries.SearchResult{}), nil, false)
	operation("/api/v1/types", "get", "queryTypes", nil, schema(queries.TypeResult{}), []any{parameter("name", "query", true), parameter("viewId", "query", false), parameter("projectId", "query", false), parameter("buildProfileId", "query", false)}, false)
	operation("/api/v1/files/{id}", "get", "getSourceFile", nil, schema(search.File{}), []any{parameter("id", "path", true), parameter("viewId", "query", false)}, false)
	operation("/api/v1/capabilities", "get", "getCapabilities", nil, map[string]any{"type": "object", "properties": map[string]any{"features": map[string]any{"type": "array", "items": map[string]any{"type": "object", "properties": map[string]any{"name": map[string]string{"type": "string"}, "status": map[string]string{"type": "string"}}}}, "limits": map[string]any{"type": "object", "additionalProperties": map[string]string{"type": "integer"}}, "authentication": map[string]string{"type": "string"}, "buildProfileId": map[string]string{"type": "string"}}}, nil, false)
	b.definitions["Error"] = map[string]any{"type": "object", "properties": map[string]any{"requestId": map[string]string{"type": "string"}, "error": map[string]any{"type": "object", "properties": map[string]any{"code": map[string]string{"type": "string"}, "message": map[string]string{"type": "string"}}}}}
	doc := map[string]any{"openapi": "3.1.0", "info": map[string]string{"title": "valio.code local bootstrap API", "version": "1.0.0", "description": "One database-bound workspace. Cookie mutations require a trusted Origin. Go type evidence is syntax-only under buildProfileId syntax-default. Latest view resolves once; pin returned viewId for subsequent requests. Maximum JSON body and scanned source are each 16 MiB; maximum 10,000 files. Project/repository IDs are explicitly registered before ingestion. Unsupported projections report unsupported. MCP is separately mounted at /mcp with bearer authentication."}, "servers": []any{map[string]string{"url": "http://127.0.0.1:8090"}}, "security": []any{map[string]any{"bearerAuth": []any{}}, map[string]any{"cookieAuth": []any{}}}, "paths": paths, "components": map[string]any{"schemas": b.definitions, "securitySchemes": map[string]any{"bearerAuth": map[string]string{"type": "http", "scheme": "bearer"}, "cookieAuth": map[string]string{"type": "apiKey", "in": "cookie", "name": "valio_session"}}}}
	encoded, e := json.MarshalIndent(doc, "", "  ")
	if e != nil {
		panic(e)
	}
	encoded = append(encoded, '\n')
	if e = os.WriteFile("api/openapi/openapi.json", encoded, 0644); e != nil {
		panic(e)
	}
}
