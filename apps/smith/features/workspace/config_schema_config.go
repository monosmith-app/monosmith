package workspace

import _ "embed"

// Schema is the JSON Schema of smith.toml, for editors (Even Better TOML reads
// the #:schema line smith new writes) and `smith schema`.
//
//go:embed smith.schema.json
var Schema []byte

// SchemaURL is where the schema is served; it is the file in this repository.
const SchemaURL = "https://raw.githubusercontent.com/monosmith-app/monosmith/main/apps/smith/workspace/smith.schema.json"
