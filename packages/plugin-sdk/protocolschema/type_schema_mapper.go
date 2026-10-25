package protocolschema

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// emptyDef is the name of the definition of an object with no members: the
// params and result of shutdown.
const emptyDef = "Empty"

var rawMessageType = reflect.TypeFor[json.RawMessage]()

// typeMapper turns Go types into schema nodes, collecting a definition for
// every named struct and enumeration it meets.
type typeMapper struct {
	docs *SourceDocs
	defs map[string]Object
}

func newTypeMapper(docs *SourceDocs) *typeMapper {
	return &typeMapper{docs: docs, defs: map[string]Object{}}
}

// schema returns the schema of t as a value: a $ref for a named struct or
// enumeration, an inline schema otherwise.
func (m *typeMapper) schema(t reflect.Type) (Object, error) {
	if t == rawMessageType {
		return Object{}, nil // any JSON value
	}
	switch t.Kind() {
	case reflect.Pointer:
		return m.schema(t.Elem())
	case reflect.Interface:
		return Object{}, nil
	case reflect.Bool:
		return Object{{Key: "type", Value: "boolean"}}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return Object{{Key: "type", Value: "integer"}}, nil
	case reflect.Float32, reflect.Float64:
		return Object{{Key: "type", Value: "number"}}, nil
	case reflect.String:
		return m.stringSchema(t), nil
	case reflect.Slice:
		items, err := m.schema(t.Elem())
		if err != nil {
			return nil, err
		}
		return Object{{Key: "type", Value: "array"}, {Key: "items", Value: items}}, nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("protocolschema: %s has non-string keys", t)
		}
		out := Object{{Key: "type", Value: "object"}}
		if t.Elem().Kind() != reflect.Interface {
			values, err := m.schema(t.Elem())
			if err != nil {
				return nil, err
			}
			out.Set("additionalProperties", values)
		}
		return out, nil
	case reflect.Struct:
		return m.structSchema(t)
	}
	return nil, fmt.Errorf("protocolschema: cannot map %s", t)
}

// stringSchema is a plain string, or, for a named type with typed constants, a
// reference to its enumeration.
func (m *typeMapper) stringSchema(t reflect.Type) Object {
	values := m.docs.Enum(t.Name())
	if t.Name() == "string" || len(values) == 0 {
		return Object{{Key: "type", Value: "string"}}
	}
	if _, done := m.defs[t.Name()]; !done {
		def := Object{}
		described := false
		descriptions := Object{}
		names := make([]string, len(values))
		for i, v := range values {
			names[i] = v.Value
			if v.Doc != "" {
				descriptions.Set(v.Value, v.Doc)
				described = true
			}
		}
		text := m.docs.TypeDoc(t.Name())
		if group := m.docs.EnumGroupDoc(t.Name()); group != "" && !described {
			text = strings.TrimSpace(text + "\n\n" + group)
		}
		if text != "" {
			def.Set("description", text)
		}
		def.Set("type", "string")
		def.Set("enum", names)
		if described {
			def.Set("x-enumDescriptions", descriptions)
		}
		m.defs[t.Name()] = def
	}
	return ref(t.Name())
}

func (m *typeMapper) structSchema(t reflect.Type) (Object, error) {
	name := t.Name()
	if name == "" {
		if t.NumField() != 0 {
			return nil, fmt.Errorf("protocolschema: anonymous struct %s", t)
		}
		name = emptyDef
		if _, done := m.defs[name]; !done {
			m.defs[name] = Object{
				{Key: "description", Value: "An object with no members."},
				{Key: "type", Value: "object"},
			}
		}
		return ref(name), nil
	}
	if _, done := m.defs[name]; done {
		return ref(name), nil
	}
	m.defs[name] = nil // reserve the name: a type may refer to itself
	properties := Object{}
	var required []string
	for i := range t.NumField() {
		f := t.Field(i)
		jsonName, omitEmpty, ok := parseJSONTag(f)
		if !ok {
			continue
		}
		prop, err := m.schema(f.Type)
		if err != nil {
			return nil, fmt.Errorf("%s.%s: %w", name, f.Name, err)
		}
		if f.Type.Kind() == reflect.Pointer && !omitEmpty {
			prop = nullable(prop)
		}
		if doc := m.docs.FieldDoc(name, f.Name); doc != "" {
			prop.Set("description", doc)
		}
		properties.Set(jsonName, prop)
		if !omitEmpty {
			required = append(required, jsonName)
		}
	}
	def := Object{}
	if doc := m.docs.TypeDoc(name); doc != "" {
		def.Set("description", doc)
	}
	def.Set("type", "object")
	def.Set("properties", properties)
	if len(required) > 0 {
		def.Set("required", required)
	}
	m.defs[name] = def
	return ref(name), nil
}

// parseJSONTag reads a field's json tag: its wire name and whether omitempty
// makes it optional. ok is false for a field that is not on the wire.
func parseJSONTag(f reflect.StructField) (name string, omitEmpty, ok bool) {
	if !f.IsExported() {
		return "", false, false
	}
	tag := f.Tag.Get("json")
	if tag == "-" {
		return "", false, false
	}
	parts := strings.Split(tag, ",")
	name = parts[0]
	if name == "" {
		name = f.Name
	}
	for _, option := range parts[1:] {
		if option == "omitempty" || option == "omitzero" {
			omitEmpty = true
		}
	}
	return name, omitEmpty, true
}

// nullable allows null as well: a required pointer is null on the wire when
// it points at nothing.
func nullable(prop Object) Object {
	if kind, ok := prop.Get("type"); ok {
		if s, isString := kind.(string); isString {
			prop.Set("type", []string{s, "null"})
			return prop
		}
	}
	return Object{{Key: "anyOf", Value: []any{prop, Object{{Key: "type", Value: "null"}}}}}
}

// definitions returns every definition collected so far, sorted by name.
func (m *typeMapper) definitions() Object {
	names := make([]string, 0, len(m.defs))
	for name := range m.defs {
		names = append(names, name)
	}
	sort.Strings(names)
	out := make(Object, 0, len(names))
	for _, name := range names {
		out = append(out, Member{Key: name, Value: m.defs[name]})
	}
	return out
}
