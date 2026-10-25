package protocolschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Member is one key of an Object.
type Member struct {
	Key   string
	Value any
}

// Object is a JSON object that keeps the order its members were set in, so the
// schema reads the way the Go types are written and the output never moves
// between runs.
type Object []Member

// Set adds a member, or replaces the value of an existing one in place.
func (o *Object) Set(key string, value any) {
	for i := range *o {
		if (*o)[i].Key == key {
			(*o)[i].Value = value
			return
		}
	}
	*o = append(*o, Member{Key: key, Value: value})
}

// Get returns a member's value.
func (o Object) Get(key string) (any, bool) {
	for _, m := range o {
		if m.Key == key {
			return m.Value, true
		}
	}
	return nil, false
}

// ref is the object {"$ref": "#/$defs/<name>"}.
func ref(name string) Object { return Object{{Key: "$ref", Value: "#/$defs/" + name}} }

// Render writes a value as JSON with two-space indentation and a final newline.
// Scalars and short arrays of scalars stay on one line. Strings are not
// HTML-escaped: the file is read by people.
func Render(v any) ([]byte, error) {
	var b bytes.Buffer
	if err := write(&b, v, ""); err != nil {
		return nil, err
	}
	b.WriteString("\n")
	return b.Bytes(), nil
}

const inlineLimit = 100

func write(b *bytes.Buffer, v any, indent string) error {
	switch x := v.(type) {
	case Object:
		if len(x) == 0 {
			b.WriteString("{}")
			return nil
		}
		if line, ok := inlineObject(x); ok && len(line)+len(indent) <= inlineLimit {
			b.WriteString(line)
			return nil
		}
		inner := indent + "  "
		b.WriteString("{\n")
		for i, m := range x {
			b.WriteString(inner)
			key, err := scalar(m.Key)
			if err != nil {
				return err
			}
			b.WriteString(key + ": ")
			if err := write(b, m.Value, inner); err != nil {
				return err
			}
			if i < len(x)-1 {
				b.WriteString(",")
			}
			b.WriteString("\n")
		}
		b.WriteString(indent + "}")
		return nil
	case []any:
		return writeArray(b, x, indent)
	case []string:
		items := make([]any, len(x))
		for i, s := range x {
			items[i] = s
		}
		return writeArray(b, items, indent)
	default:
		s, err := scalar(v)
		if err != nil {
			return err
		}
		b.WriteString(s)
		return nil
	}
}

// inlineObject renders an object of scalars and arrays of scalars on one line,
// so {"$ref": "#/$defs/Task"} and {"type": "string"} stay short.
func inlineObject(o Object) (string, bool) {
	parts := make([]string, len(o))
	for i, m := range o {
		var value string
		switch v := m.Value.(type) {
		case Object:
			return "", false
		case []any:
			return "", false
		case []string:
			var inner bytes.Buffer
			if err := writeArray(&inner, toAny(v), ""); err != nil || strings.Contains(inner.String(), "\n") {
				return "", false
			}
			value = inner.String()
		default:
			s, err := scalar(v)
			if err != nil {
				return "", false
			}
			value = s
		}
		key, err := scalar(m.Key)
		if err != nil {
			return "", false
		}
		parts[i] = key + ": " + value
	}
	return "{" + strings.Join(parts, ", ") + "}", true
}

func toAny(in []string) []any {
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func writeArray(b *bytes.Buffer, items []any, indent string) error {
	if len(items) == 0 {
		b.WriteString("[]")
		return nil
	}
	allScalar := true
	parts := make([]string, len(items))
	for i, item := range items {
		switch item.(type) {
		case Object, []any, []string:
			allScalar = false
		default:
			s, err := scalar(item)
			if err != nil {
				return err
			}
			parts[i] = s
		}
	}
	if allScalar {
		if line := "[" + strings.Join(parts, ", ") + "]"; len(line)+len(indent) <= inlineLimit {
			b.WriteString(line)
			return nil
		}
	}
	inner := indent + "  "
	b.WriteString("[\n")
	for i, item := range items {
		b.WriteString(inner)
		if err := write(b, item, inner); err != nil {
			return err
		}
		if i < len(items)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(indent + "]")
	return nil
}

func scalar(v any) (string, error) {
	switch v.(type) {
	case nil, string, bool, int, int64, float64:
	default:
		return "", fmt.Errorf("protocolschema: cannot render %T", v)
	}
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimRight(b.String(), "\n"), nil
}
