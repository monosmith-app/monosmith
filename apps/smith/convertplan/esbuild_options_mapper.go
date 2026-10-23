package convertplan

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nxworkspace"
)

// EsbuildOptions are an @nx/esbuild build's esbuildOptions as esbuild CLI
// flags. Sourcemap says whether the build writes source maps (on unless
// esbuildOptions turns them off); Unmapped names the options no flag stands
// for, so convert can report them instead of dropping them silently.
type EsbuildOptions struct {
	Flags     []string
	Sourcemap string
	Unmapped  []string
}

// MapEsbuildOptions reads a target's esbuildOptions. An identity
// outExtension (".js" to ".js") needs no flag.
func MapEsbuildOptions(t nxworkspace.Target) EsbuildOptions {
	res := EsbuildOptions{Sourcemap: "--sourcemap"}
	var opts map[string]json.RawMessage
	if !t.Option("esbuildOptions", &opts) {
		return res
	}
	keys := make([]string, 0, len(opts))
	for k := range opts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		flags, ok := esbuildOptionFlags(k, opts[k], &res)
		if !ok {
			res.Unmapped = append(res.Unmapped, k)
			continue
		}
		res.Flags = append(res.Flags, flags...)
	}
	return res
}

func esbuildOptionFlags(key string, raw json.RawMessage, res *EsbuildOptions) ([]string, bool) {
	switch key {
	case "sourcemap":
		var on bool
		if json.Unmarshal(raw, &on) == nil {
			res.Sourcemap = ""
			if on {
				res.Sourcemap = "--sourcemap"
			}
			return nil, true
		}
		var kind string
		if json.Unmarshal(raw, &kind) == nil && plainFlagValue(kind) {
			res.Sourcemap = "--sourcemap=" + kind
			return nil, true
		}
	case "outExtension", "loader", "define", "logOverride":
		var m map[string]string
		if json.Unmarshal(raw, &m) != nil {
			return nil, false
		}
		name := map[string]string{"outExtension": "out-extension", "loader": "loader", "define": "define", "logOverride": "log-override"}[key]
		return mapFlags(name, m, key == "outExtension")
	case "external", "mainFields", "conditions", "target", "resolveExtensions":
		var list []string
		if json.Unmarshal(raw, &list) != nil {
			var one string
			if json.Unmarshal(raw, &one) != nil {
				return nil, false
			}
			list = []string{one}
		}
		for _, v := range list {
			if !plainFlagValue(v) {
				return nil, false
			}
		}
		if key == "external" {
			flags := make([]string, len(list))
			for i, v := range list {
				flags[i] = "--external:" + v
			}
			return flags, true
		}
		name := map[string]string{"mainFields": "main-fields", "conditions": "conditions", "target": "target", "resolveExtensions": "resolve-extensions"}[key]
		return []string{"--" + name + "=" + strings.Join(list, ",")}, true
	case "minify", "keepNames", "treeShaking":
		var on bool
		if json.Unmarshal(raw, &on) != nil {
			return nil, false
		}
		name := map[string]string{"minify": "minify", "keepNames": "keep-names", "treeShaking": "tree-shaking"}[key]
		return []string{fmt.Sprintf("--%s=%t", name, on)}, true
	case "legalComments", "charset", "jsx", "logLevel":
		var v string
		if json.Unmarshal(raw, &v) != nil || !plainFlagValue(v) {
			return nil, false
		}
		name := map[string]string{"legalComments": "legal-comments", "charset": "charset", "jsx": "jsx", "logLevel": "log-level"}[key]
		return []string{"--" + name + "=" + v}, true
	}
	return nil, false
}

// mapFlags is esbuild's --name:key=value form, in key order. skipIdentity
// leaves out a key mapped to itself.
func mapFlags(name string, m map[string]string, skipIdentity bool) ([]string, bool) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var flags []string
	for _, k := range keys {
		if skipIdentity && k == m[k] {
			continue
		}
		if !plainFlagValue(k) || !plainFlagValue(m[k]) {
			return nil, false
		}
		flags = append(flags, "--"+name+":"+k+"="+m[k])
	}
	return flags, true
}

// plainFlagValue reports whether a value can sit in a script unquoted, in
// every shell npm runs scripts with.
func plainFlagValue(v string) bool {
	return v != "" && !strings.ContainsAny(v, " \t'\"`$%\\!&|<>;()*?[]{}^~")
}
