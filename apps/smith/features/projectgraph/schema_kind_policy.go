package projectgraph

import (
	"path"
	"strings"
)

// SchemaKind names the kind of shared schema a file is, from its name:
// "openapi" (openapi/swagger .yaml, .yml, .json), "asyncapi", "protobuf"
// (*.proto) or "graphql" (*.graphql, *.gql). It is "" for any other file.
func SchemaKind(file string) string {
	base := strings.ToLower(path.Base(file))
	ext := path.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	switch ext {
	case ".proto":
		return "protobuf"
	case ".graphql", ".gql":
		return "graphql"
	case ".yaml", ".yml", ".json":
		switch stem {
		case "openapi", "swagger":
			return "openapi"
		case "asyncapi":
			return "asyncapi"
		}
	}
	return ""
}
