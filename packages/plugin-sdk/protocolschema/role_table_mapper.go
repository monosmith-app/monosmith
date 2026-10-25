package protocolschema

import (
	"slices"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Cardinalities of a role in one project, as the schema names them.
const (
	cardinalityExclusive = "exclusive" // at most one plugin per project
	cardinalityAdditive  = "additive"  // any number of plugins, each only adding
	cardinalityPerSlot   = "perSlot"   // one plugin per task name it provides
)

// roleTable lists the roles in the order a project's make-up lists them, with
// how many plugins may play each in one project and which kinds of plugin play
// it when the plugin names none. The values come from the protocol package's
// own functions, so the schema cannot disagree with the Go SDK.
func roleTable(docs *SourceDocs) []any {
	described := map[string]string{}
	for _, c := range docs.Enum("Role") {
		described[c.Value] = c.Doc
	}
	kinds := []protocol.PluginKind{protocol.KindLanguage, protocol.KindConvention}
	rows := make([]any, 0, len(protocol.Roles))
	for _, role := range protocol.Roles {
		cardinality := cardinalityPerSlot
		switch {
		case role.Exclusive():
			cardinality = cardinalityExclusive
		case role.Additive():
			cardinality = cardinalityAdditive
		}
		var defaultFor []string
		for _, kind := range kinds {
			if slices.Contains(protocol.DefaultRoles(kind), role) {
				defaultFor = append(defaultFor, string(kind))
			}
		}
		row := Object{{Key: "name", Value: string(role)}}
		if text := described[string(role)]; text != "" {
			row.Set("description", text)
		}
		row.Set("cardinality", cardinality)
		row.Set("defaultFor", defaultFor)
		rows = append(rows, row)
	}
	return rows
}
