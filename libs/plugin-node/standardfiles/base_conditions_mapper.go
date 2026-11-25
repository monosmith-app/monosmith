package standardfiles

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// BaseCompilerOptions is the "compilerOptions" of the workspace's
// tsconfig.base.json with the source condition in "customConditions", the
// other options kept in their order.
//
// Parameters: raw is the tsconfig.base.json.
// Returns: the new "compilerOptions", and false when the base cannot carry
// the condition: it has comments smith cannot keep, it is not a JSON object,
// or it resolves modules the node10 or classic way, which ignores conditions.
func BaseCompilerOptions(raw []byte) (json.RawMessage, bool) {
	if !bytes.Equal(jsondoc.StripComments(raw), raw) {
		return nil, false
	}
	doc, err := parseObject(raw)
	if err != nil {
		return nil, false
	}
	options := object{}
	if value, ok := doc.get("compilerOptions"); ok {
		if options, err = parseObject(value); err != nil {
			return nil, false
		}
	}
	if !resolvesConditions(options) {
		return nil, false
	}
	var conditions []string
	if value, ok := options.get("customConditions"); ok {
		if json.Unmarshal(value, &conditions) != nil {
			return nil, false
		}
	}
	for _, c := range conditions {
		if c == Condition {
			return marshal(options)
		}
	}
	list, err := json.Marshal(append(conditions, Condition))
	if err != nil {
		return nil, false
	}
	return marshal(options.set("customConditions", list))
}

// resolvesConditions says whether the options let tsc read custom conditions:
// the node16, nodenext and bundler module resolutions do. Options that name
// neither moduleResolution nor module take the preset's (nodenext).
func resolvesConditions(options object) bool {
	read := func(key string) string {
		var s string
		if v, ok := options.get(key); ok {
			_ = json.Unmarshal(v, &s)
		}
		return strings.ToLower(s)
	}
	switch read("moduleResolution") {
	case "node16", "nodenext", "bundler":
		return true
	case "":
	default:
		return false
	}
	switch read("module") {
	case "", "node16", "node18", "node20", "nodenext", "preserve":
		return true
	}
	return false
}
