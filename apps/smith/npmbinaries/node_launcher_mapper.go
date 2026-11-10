package npmbinaries

import (
	_ "embed"
	"encoding/json"
	"strings"
)

// shimTemplate is the wrapper's bin, with __BINARY__ and __PLATFORMS__ to fill.
//
//go:embed shim_template.js
var shimTemplate string

// Shim is the wrapper's bin: it finds the platform package npm installed
// and runs its binary with the same arguments, exit code and signal. When the
// package is missing it says which of three situations that is, and offers the
// fix without applying it.
func Shim(binary string, platforms map[string]string) string {
	raw, _ := json.MarshalIndent(platforms, "", "  ")
	return strings.NewReplacer("__BINARY__", binary, "__PLATFORMS__", string(raw)).Replace(shimTemplate)
}
