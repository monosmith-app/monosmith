package standardfiles

import (
	"bytes"
	"encoding/json"
	"errors"
)

// member is one member of a JSON object, its value kept as written.
type member struct {
	Key   string
	Value json.RawMessage
}

// object is a JSON object whose members keep their order: in a package.json
// "exports" the order of the conditions is meaning.
type object []member

// parseObject reads a JSON object, keeping the order of its members.
//
// Errors: raw is not a JSON object.
func parseObject(raw []byte) (object, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, errors.New("not a JSON object")
	}
	var out object
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := tok.(string)
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, err
		}
		out = append(out, member{Key: key, Value: v})
	}
	return out, nil
}

// get returns the value of a member and whether it is there.
func (o object) get(key string) (json.RawMessage, bool) {
	for _, m := range o {
		if m.Key == key {
			return m.Value, true
		}
	}
	return nil, false
}

// without is the object less one member.
func (o object) without(key string) object {
	var out object
	for _, m := range o {
		if m.Key != key {
			out = append(out, m)
		}
	}
	return out
}

// set replaces the value of a member, or adds the member last.
func (o object) set(key string, v json.RawMessage) object {
	for i, m := range o {
		if m.Key == key {
			out := append(object(nil), o...)
			out[i].Value = v
			return out
		}
	}
	return append(append(object(nil), o...), member{Key: key, Value: v})
}

// MarshalJSON writes the members in their order.
func (o object) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, m := range o {
		if i > 0 {
			b.WriteByte(',')
		}
		k, err := json.Marshal(m.Key)
		if err != nil {
			return nil, err
		}
		b.Write(k)
		b.WriteByte(':')
		b.Write(m.Value)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

// str is a JSON string value.
func str(s string) json.RawMessage {
	raw, _ := json.Marshal(s)
	return raw
}
