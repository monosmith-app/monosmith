package protocol

import (
	"fmt"
	"regexp"
	"sort"
)

// factKeyPattern is a namespaced key: two or more lowercase dot-separated
// parts, the first starting with a letter.
var factKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9-]*(\.[a-z0-9][a-z0-9-]*)+$`)

// ValidateFact checks one fact: the key must be namespaced, and a well-known
// key must hold the type its constant documents. A key that is not
// well-known may hold any value.
//
// key is the fact's key. value is a decoded JSON or TOML value (string,
// []any or []string, map[string]any or map[string]string, number, bool).
// It returns an error that names the key when the key or the value is
// invalid, and nil otherwise.
func ValidateFact(key string, value any) error {
	if !factKeyPattern.MatchString(key) {
		return fmt.Errorf("fact %q: a key is a namespace and a name, such as module.aliases", key)
	}
	var ok bool
	var want string
	switch key {
	case FactModuleAliases:
		want, ok = "an object of strings", isStringMap(value)
	case FactSourceRoots, FactTestPatterns, FactBuildOutputs:
		want, ok = "an array of strings", isStringList(value)
	case FactRuntimeNode:
		_, ok = value.(string)
		want = "a string"
	default:
		return nil
	}
	if !ok {
		return fmt.Errorf("fact %q must be %s", key, want)
	}
	return nil
}

// ValidateFacts checks every fact of a set; see ValidateFact. It returns the
// error of the first invalid key in key order.
func ValidateFacts(facts map[string]any) error {
	keys := make([]string, 0, len(facts))
	for k := range facts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := ValidateFact(key, facts[key]); err != nil {
			return err
		}
	}
	return nil
}

func isStringMap(v any) bool {
	switch m := v.(type) {
	case map[string]string:
		return true
	case map[string]any:
		for _, e := range m {
			if _, ok := e.(string); !ok {
				return false
			}
		}
		return true
	}
	return false
}

func isStringList(v any) bool {
	switch l := v.(type) {
	case []string:
		return true
	case []any:
		for _, e := range l {
			if _, ok := e.(string); !ok {
				return false
			}
		}
		return true
	}
	return false
}
