package protocol

import "testing"

func TestValidateFactChecksKeyAndType(t *testing.T) {
	for name, tc := range map[string]struct {
		key   string
		value any
		ok    bool
	}{
		"aliases":               {FactModuleAliases, map[string]any{"@app": "src"}, true},
		"aliases not strings":   {FactModuleAliases, map[string]any{"@app": 1}, false},
		"aliases as a list":     {FactModuleAliases, []any{"src"}, false},
		"roots":                 {FactSourceRoots, []any{"src"}, true},
		"roots as a string":     {FactSourceRoots, "src", false},
		"node runtime":          {FactRuntimeNode, "22", true},
		"node runtime number":   {FactRuntimeNode, int64(22), false},
		"plugin key, any type":  {"acme.thing", map[string]any{"a": 1}, true},
		"key without namespace": {"aliases", "x", false},
		"key in capitals":       {"Module.Aliases", "x", false},
	} {
		if err := ValidateFact(tc.key, tc.value); (err == nil) != tc.ok {
			t.Errorf("%s: ValidateFact(%q) = %v, want ok=%v", name, tc.key, err, tc.ok)
		}
	}
}

func TestValidateFactsNamesTheFirstBadKey(t *testing.T) {
	err := ValidateFacts(map[string]any{FactSourceRoots: "src", FactRuntimeNode: "22"})
	if err == nil {
		t.Fatal("expected an error")
	}
}
