package projectgraph

import "testing"

func TestSchemaKind(t *testing.T) {
	for file, want := range map[string]string{
		"api/openapi.yaml": "openapi", "api/Swagger.JSON": "openapi", "x/asyncapi.yml": "asyncapi",
		"proto/user/v1/user.proto": "protobuf", "gql/schema.graphql": "graphql", "q.gql": "graphql",
		"api/package.json": "", "openapi.md": "", "api/openapi-gen.yaml": "",
	} {
		if got := SchemaKind(file); got != want {
			t.Errorf("SchemaKind(%q) = %q, want %q", file, got, want)
		}
	}
}
