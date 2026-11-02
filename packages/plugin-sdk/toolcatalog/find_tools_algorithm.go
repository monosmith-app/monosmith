package toolcatalog

import (
	"fmt"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// For returns the catalogue entries with the given names, in the order given.
// A plugin lists its tools with it: Tools: toolcatalog.For("python").
//
// Parameters: names are entry names ("python", "git").
// Returns: copies of the entries.
// Panics when a name is not in the catalogue: the plugin's author wrote it, so
// it is a programming error found by the first test run.
func For(names ...string) []protocol.ToolNeed {
	out := make([]protocol.ToolNeed, 0, len(names))
	for _, name := range names {
		t, ok := Find(name)
		if !ok {
			panic(fmt.Sprintf("toolcatalog: no tool named %q", name))
		}
		out = append(out, t)
	}
	return out
}

// Find returns the catalogue entry with a name.
//
// Parameters: name is an entry name.
// Returns: a copy of the entry and whether it exists.
func Find(name string) (protocol.ToolNeed, bool) {
	for _, t := range entries {
		if t.Name == name {
			return t, true
		}
	}
	return protocol.ToolNeed{}, false
}

// All returns a copy of every entry.
func All() []protocol.ToolNeed { return append([]protocol.ToolNeed(nil), entries...) }
