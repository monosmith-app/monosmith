package catalog

import "embed"

// Files is the catalogue: one folder per locale, each holding JSON files that map
// a key to its text.
//
//go:embed */*.json
var Files embed.FS
