package standardfiles

import (
	"encoding/json"
	"path"
	"strings"
)

// Condition is the custom resolution condition that leads tsc and editors to
// a workspace library's TypeScript sources; Node, bundlers and npm ignore it,
// so what runs and what is published still read the build.
const Condition = "source"

// manifest is the part of a package.json the entry points come from.
type manifest struct {
	Main    string          `json:"main"`
	Module  string          `json:"module"`
	Types   string          `json:"types"`
	Typings string          `json:"typings"`
	Exports json.RawMessage `json:"exports"`
}

// SourceExports is the "exports" of a library whose entry points lead to its
// build, with the source condition first in its main entry ("."), so that
// Ctrl+click on an import opens the library's source. The other conditions and
// subpaths keep their values and their order. A library with no "exports"
// gets one from main, module and types, with "./*" so a deep import still
// resolves as before.
//
// Parameters: raw is the package.json; source is the entry source file,
// relative to the package ("./src/index.ts").
// Returns: the new "exports" value, and false when the entry points do not
// lead to a build (the library already exports its sources) or the manifest
// cannot be read.
func SourceExports(raw []byte, source string) (json.RawMessage, bool) {
	var m manifest
	if json.Unmarshal(raw, &m) != nil {
		return nil, false
	}
	if len(m.Exports) == 0 || string(m.Exports) == "null" {
		return fromFields(m, source)
	}
	var s string
	if json.Unmarshal(m.Exports, &s) == nil {
		if !built(s) {
			return nil, false
		}
		main := object{{Key: Condition, Value: str(source)}}
		if t := typesOf(m); t != "" {
			main = main.set("types", str(relative(t)))
		}
		return marshal(object{{Key: ".", Value: mustMarshal(main.set("default", str(s)))}})
	}
	exports, err := parseObject(m.Exports)
	if err != nil || len(exports) == 0 {
		return nil, false
	}
	if !strings.HasPrefix(exports[0].Key, ".") {
		// Conditions only: they are the main entry.
		entry, ok := withSource(m.Exports, source)
		if !ok {
			return nil, false
		}
		return entry, true
	}
	value, ok := exports.get(".")
	if !ok {
		return nil, false
	}
	entry, ok := withSource(value, source)
	if !ok {
		return nil, false
	}
	return marshal(exports.set(".", entry))
}

// withSource puts the source condition first in an entry: a path string
// becomes {"source", "default"}, a conditions object gets it before its other
// members. ok is false when the entry does not lead to a build.
func withSource(entry json.RawMessage, source string) (json.RawMessage, bool) {
	var s string
	if json.Unmarshal(entry, &s) == nil {
		if !built(s) {
			return nil, false
		}
		return marshal(object{{Key: Condition, Value: str(source)}, {Key: "default", Value: str(s)}})
	}
	conds, err := parseObject(entry)
	if err != nil || !anyBuilt(conds) {
		return nil, false
	}
	return marshal(append(object{{Key: Condition, Value: str(source)}}, conds.without(Condition)...))
}

// fromFields is the "exports" of a package that has none: its main entry from
// types and main (or module), and "./*" for every other path.
func fromFields(m manifest, source string) (json.RawMessage, bool) {
	runtime := m.Main
	if runtime == "" {
		runtime = m.Module
	}
	types := typesOf(m)
	if runtime == "" || !built(runtime) && !built(types) {
		return nil, false
	}
	entry := object{{Key: Condition, Value: str(source)}}
	if types != "" {
		entry = entry.set("types", str(relative(types)))
	}
	entry = entry.set("default", str(relative(runtime)))
	return marshal(object{{Key: ".", Value: mustMarshal(entry)}, {Key: "./*", Value: str("./*")}})
}

func typesOf(m manifest) string {
	if m.Types != "" {
		return m.Types
	}
	return m.Typings
}

// anyBuilt says whether a conditions object leads to a build somewhere.
func anyBuilt(conds object) bool {
	for _, c := range conds {
		var s string
		if json.Unmarshal(c.Value, &s) == nil && built(s) {
			return true
		}
		if nested, err := parseObject(c.Value); err == nil && anyBuilt(nested) {
			return true
		}
	}
	return false
}

// built says whether an entry path names built code: JavaScript or a
// declaration file, not a TypeScript source.
func built(p string) bool {
	p = path.Clean(strings.TrimPrefix(p, "./"))
	if p == "" || p == "." {
		return false
	}
	if strings.HasSuffix(p, ".d.ts") || strings.HasSuffix(p, ".d.mts") || strings.HasSuffix(p, ".d.cts") {
		return true
	}
	switch path.Ext(p) {
	case ".js", ".mjs", ".cjs":
		return true
	}
	return false
}

// relative writes a package path the way "exports" requires: from "./".
func relative(p string) string {
	if strings.HasPrefix(p, "./") {
		return p
	}
	return "./" + strings.TrimPrefix(p, "/")
}

func marshal(o object) (json.RawMessage, bool) {
	raw, err := json.Marshal(o)
	return raw, err == nil
}

func mustMarshal(o object) json.RawMessage {
	raw, _ := json.Marshal(o)
	return raw
}
