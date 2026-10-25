package protocolschema

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/transport"
)

// SchemaID is where the committed schema is published.
const SchemaID = "https://raw.githubusercontent.com/monosmith-app/monosmith/main/packages/plugin-sdk/schema/protocol.schema.json"

// rootDescription says what the schema is and the rules of the wire that a
// type definition cannot say.
var rootDescription = strings.Join([]string{
	"The messages of the smith plugin protocol: JSON-RPC 2.0 between the smith host and a plugin, a separate process. " +
		"This file is generated from the Go types of packages/plugin-sdk/protocol, which are the source of truth; do not edit it. " +
		"Regenerate it with `go run ./schemagen` in packages/plugin-sdk.",
	"x-methods maps each method name to the definitions of its params and result and to the capability a plugin advertises to be asked it " +
		"(null: always available). x-errorCodes, x-roles and x-transport complete the contract. A definition's required members are the ones " +
		"the Go SDK writes without omitempty.",
	"Framing: each message is the header Content-Length: <bytes of the body>, a blank line (CRLF CRLF) and the body, UTF-8 JSON, as the Language Server Protocol frames. " +
		"The length counts bytes, not characters. The host sends initialize first, and a request with an id gets exactly one response with the same id. " +
		"A request without an id is a notification and gets none. The host has at most as many requests in flight to a plugin as its initialize result declares in concurrency (absent: 1, the next request is sent after the previous is answered); " +
		"responses may come in any order, matched by id. A plugin that advertised the cancel capability is sent the cancel notification for a request the host gives up on, and answers that request with its result or the request-cancelled error; " +
		"a plugin without it is never sent cancel. Stdout carries only frames: a plugin logs to stderr. The host sets SMITH_PLUGIN=1 in the plugin's environment.",
	"Reading: a receiver ignores members it does not know (the protocol grows by adding members), and treats a null array or object as an empty one, " +
		"because the Go SDK writes a nil slice or map as null when the member is required. A member that is absent and a member with its zero value are the same for every optional member.",
	"Paths in messages are relative to the workspace and use forward slashes; only workspaceRoot is absolute and uses the host's separators.",
}, "\n\n")

// Build returns the JSON Schema of the protocol, rendered, for the protocol
// package whose source is in protocolDir. It fails when the methods or
// error-code tables are out of step with that package.
func Build(protocolDir string) ([]byte, error) {
	docs, err := ReadSourceDocs(protocolDir)
	if err != nil {
		return nil, err
	}
	if err := CheckTables(docs); err != nil {
		return nil, err
	}
	m := newTypeMapper(docs)
	for _, t := range []reflect.Type{reflect.TypeFor[protocol.Request](), reflect.TypeFor[protocol.Response](), reflect.TypeFor[protocol.ResponseError]()} {
		if _, err := m.schema(t); err != nil {
			return nil, err
		}
	}
	methods, err := methodsTable(m)
	if err != nil {
		return nil, err
	}
	codes, err := errorCodesTable(docs)
	if err != nil {
		return nil, err
	}

	root := Object{
		{Key: "$schema", Value: "https://json-schema.org/draft/2020-12/schema"},
		{Key: "$id", Value: SchemaID},
		{Key: "title", Value: "smith plugin protocol"},
		{Key: "description", Value: rootDescription},
		{Key: "x-protocolVersion", Value: protocol.Version},
		{Key: "x-transport", Value: Object{
			{Key: "framing", Value: "content-length"},
			{Key: "encoding", Value: "utf-8"},
			{Key: "jsonrpc", Value: "2.0"},
			{Key: "maxMessageBytes", Value: transport.MaxMessageSize},
		}},
		{Key: "x-methods", Value: methods},
		{Key: "x-errorCodes", Value: codes},
		{Key: "x-roles", Value: roleTable(docs)},
		{Key: "$defs", Value: m.definitions()},
	}
	return Render(root)
}

func methodsTable(m *typeMapper) (Object, error) {
	table := Object{}
	capabilities := reflect.TypeFor[protocol.Capabilities]()
	for _, spec := range Methods {
		params, err := m.schema(spec.Params)
		if err != nil {
			return nil, fmt.Errorf("%s params: %w", spec.Method, err)
		}
		var result any
		if !spec.Notification {
			result, err = m.schema(spec.Result)
			if err != nil {
				return nil, fmt.Errorf("%s result: %w", spec.Method, err)
			}
		}
		var capability any
		if spec.Capability != "" {
			field, _ := capabilities.FieldByName(spec.Capability)
			capability, _, _ = parseJSONTag(field)
		}
		table.Set(spec.Method, Object{
			{Key: "description", Value: spec.Description},
			{Key: "params", Value: params},
			{Key: "result", Value: result},
			{Key: "capability", Value: capability},
			{Key: "notification", Value: spec.Notification},
		})
	}
	return table, nil
}

func errorCodesTable(docs *SourceDocs) ([]any, error) {
	rows := make([]any, 0, len(ErrorCodes))
	for _, spec := range ErrorCodes {
		c, ok := docs.Const(spec.Const)
		if !ok {
			return nil, fmt.Errorf("error code %s is not declared", spec.Const)
		}
		var code int
		if _, err := fmt.Sscanf(c.Value, "%d", &code); err != nil {
			return nil, fmt.Errorf("error code %s = %q is not an integer", spec.Const, c.Value)
		}
		rows = append(rows, Object{
			{Key: "name", Value: strings.TrimPrefix(spec.Const, "Code")},
			{Key: "code", Value: code},
			{Key: "description", Value: spec.Description},
		})
	}
	return rows, nil
}
