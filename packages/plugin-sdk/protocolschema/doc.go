// Package protocolschema derives the JSON Schema (draft 2020-12) of the plugin
// protocol from the Go types in the protocol package. The Go types are the
// source of truth; the schema, committed as schema/protocol.schema.json, is the
// contract the SDKs of other languages generate their types from.
//
// The schema is built by reflection over the types' json tags (a field with
// omitempty is optional) and by reading the protocol package's source for doc
// comments and for the typed string constants that make an enumeration. The
// methods table, the error codes and the roles are added next to the
// definitions. Build refuses to run when the protocol has a method, an error
// code or a capability the tables do not know, so a protocol change cannot
// reach a release without the schema being looked at.
//
// A test fails when the committed schema differs from what Build produces.
// Regenerate it from the module's root with
//
//	go run ./schemagen
//
//go:generate go run ../schemagen
package protocolschema
