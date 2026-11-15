package workspace

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// The schema describes exactly the keys smith.toml accepts, at every level.
func TestSchemaMatchesTheConfig(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal(Schema, &schema); err != nil {
		t.Fatal(err)
	}
	if schema["$id"] != SchemaURL {
		t.Fatalf("$id %v", schema["$id"])
	}
	definitions, _ := schema["definitions"].(map[string]any)
	var walk func(typ reflect.Type, node map[string]any, at string)
	walk = func(typ reflect.Type, node map[string]any, at string) {
		if ref, ok := node["$ref"].(string); ok {
			node = definitions[strings.TrimPrefix(ref, "#/definitions/")].(map[string]any)
		}
		for typ.Kind() == reflect.Pointer {
			typ = typ.Elem()
		}
		switch typ.Kind() {
		case reflect.Struct:
			if node["additionalProperties"] != false {
				t.Errorf("%s: a table must reject unknown keys", at)
			}
			props, _ := node["properties"].(map[string]any)
			var want, have []string
			for i := 0; i < typ.NumField(); i++ {
				f := typ.Field(i)
				key := f.Tag.Get("toml")
				want = append(want, key)
				if sub, ok := props[key].(map[string]any); ok {
					walk(f.Type, sub, at+"."+key)
				}
			}
			for k := range props {
				have = append(have, k)
			}
			sort.Strings(want)
			sort.Strings(have)
			if strings.Join(want, ",") != strings.Join(have, ",") {
				t.Errorf("%s: config has %v, schema %v", at, want, have)
			}
		case reflect.Map:
			sub, ok := node["additionalProperties"].(map[string]any)
			if !ok {
				t.Errorf("%s: a map needs additionalProperties", at)
				return
			}
			walk(typ.Elem(), sub, at+".*")
		}
	}
	walk(reflect.TypeOf(Config{}), schema, "smith.toml")
}
