package protocolschema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type shade string

type sample struct {
	// Name is required.
	Name   string            `json:"name"`
	Nick   string            `json:"nick,omitempty"`
	Shade  shade             `json:"shade"`
	Tags   []string          `json:"tags,omitempty"`
	Extra  map[string]any    `json:"extra,omitempty"`
	Env    map[string]string `json:"env,omitempty"`
	Count  *int64            `json:"count"`
	Child  *sample           `json:"child,omitempty"`
	Raw    json.RawMessage   `json:"raw,omitempty"`
	hidden string
	Skip   string `json:"-"`
}

func docsFor(t *testing.T, source string) *SourceDocs {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "sample.go"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	docs, err := ReadSourceDocs(dir)
	if err != nil {
		t.Fatal(err)
	}
	return docs
}

func TestAStructBecomesADefinitionInTagOrder(t *testing.T) {
	docs := docsFor(t, `package protocol

// sample is a thing.
type sample struct {
	// Name is required.
	Name string
}

type shade string

// Shades of grey.
const (
	// Light is pale.
	Light shade = "light"
	Dark  shade = "dark"
)
`)
	_ = sample{hidden: "an unexported field is not on the wire"}
	m := newTypeMapper(docs)
	if _, err := m.schema(reflect.TypeFor[sample]()); err != nil {
		t.Fatal(err)
	}
	raw, err := Render(m.definitions())
	if err != nil {
		t.Fatal(err)
	}
	var defs map[string]struct {
		Description string
		Properties  map[string]map[string]any
		Required    []string
		Enum        []string
	}
	if err := json.Unmarshal(raw, &defs); err != nil {
		t.Fatalf("%v\n%s", err, raw)
	}
	s := defs["sample"]
	if s.Description != "sample is a thing." || s.Properties["name"]["description"] != "Name is required." {
		t.Errorf("comments are not carried: %+v", s)
	}
	if got := strings.Join(s.Required, ","); got != "name,shade,count" {
		t.Errorf("required = %s (omitempty members are optional)", got)
	}
	if _, ok := s.Properties["hidden"]; ok || s.Properties["Skip"] != nil || s.Properties["-"] != nil {
		t.Errorf("unexported and skipped fields are on the wire: %v", s.Properties)
	}
	if got := s.Properties["count"]["type"]; !reflect.DeepEqual(got, []any{"integer", "null"}) {
		t.Errorf("a required pointer is nullable, got %v", got)
	}
	if got := s.Properties["child"]["$ref"]; got != "#/$defs/sample" {
		t.Errorf("a self reference is %v", got)
	}
	if _, typed := s.Properties["raw"]["type"]; typed {
		t.Errorf("a raw message is any JSON: %v", s.Properties["raw"])
	}
	if got := defs["shade"].Enum; !reflect.DeepEqual(got, []string{"light", "dark"}) {
		t.Errorf("enum = %v", got)
	}
	if extra := s.Properties["extra"]; extra["type"] != "object" || extra["additionalProperties"] != nil {
		t.Errorf("map[string]any = %v", extra)
	}
}
