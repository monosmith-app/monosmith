package jsondoc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
)

// Document is a JSON object that keeps its key order, so editing a
// package.json changes only what was edited.
type Document struct {
	keys   []string
	values map[string]any // json.RawMessage for leaves and arrays, *Document for objects
	crlf   bool           // the parsed text ended its lines with CRLF; Bytes does the same
}

// ParseDocument reads a JSON object.
func ParseDocument(raw []byte) (*Document, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("not a JSON object")
	}
	doc, err := readObject(dec)
	if err != nil {
		return nil, err
	}
	doc.crlf, _ = lineendings.IsCRLF(string(raw))
	return doc, nil
}

func readObject(dec *json.Decoder) (*Document, error) {
	doc := &Document{values: map[string]any{}}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := tok.(string)
		if !ok {
			return nil, fmt.Errorf("object key is %v", tok)
		}
		var value json.RawMessage
		if err := dec.Decode(&value); err != nil {
			return nil, err
		}
		trimmed := bytes.TrimSpace(value)
		if len(trimmed) > 0 && trimmed[0] == '{' {
			sub := json.NewDecoder(bytes.NewReader(trimmed))
			sub.UseNumber()
			if _, err := sub.Token(); err != nil {
				return nil, err
			}
			child, err := readObject(sub)
			if err != nil {
				return nil, err
			}
			doc.set(key, child)
			continue
		}
		doc.set(key, json.RawMessage(trimmed))
	}
	if _, err := dec.Token(); err != nil && err != io.EOF {
		return nil, err
	}
	return doc, nil
}

func (d *Document) set(key string, v any) {
	if _, ok := d.values[key]; !ok {
		d.keys = append(d.keys, key)
	}
	d.values[key] = v
}

// Has reports whether the key exists.
func (d *Document) Has(key string) bool { _, ok := d.values[key]; return ok }

// Object returns the object under key, creating it (at the end) when create
// is true.
func (d *Document) Object(key string, create bool) *Document {
	if v, ok := d.values[key].(*Document); ok {
		return v
	}
	if !create {
		return nil
	}
	child := &Document{values: map[string]any{}}
	d.set(key, child)
	return child
}

// SetString sets a string value, keeping the key's position.
func (d *Document) SetString(key, value string) {
	d.set(key, json.RawMessage(quote(value)))
}

// SetJSON sets any JSON-encodable value, keeping the key's position. An
// object becomes an editable Document.
func (d *Document) SetJSON(key string, value any) error {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		return err
	}
	raw := bytes.TrimSpace(b.Bytes())
	if len(raw) > 0 && raw[0] == '{' {
		child, err := ParseDocument(raw)
		if err != nil {
			return err
		}
		d.set(key, child)
		return nil
	}
	d.set(key, json.RawMessage(raw))
	return nil
}

// String returns a string value; ok is false when the key is missing or not
// a string.
func (d *Document) String(key string) (string, bool) {
	raw, isLeaf := d.values[key].(json.RawMessage)
	if !isLeaf {
		return "", false
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return "", false
	}
	return s, true
}

// Delete removes a key.
func (d *Document) Delete(key string) {
	if _, ok := d.values[key]; !ok {
		return
	}
	delete(d.values, key)
	for i, k := range d.keys {
		if k == key {
			d.keys = append(d.keys[:i], d.keys[i+1:]...)
			break
		}
	}
}

// Keys lists the keys in order.
func (d *Document) Keys() []string { return append([]string(nil), d.keys...) }

// Len is the number of keys.
func (d *Document) Len() int { return len(d.keys) }

// Bytes renders the document with two-space indentation and a final newline,
// the way npm writes package.json. Lines end with CRLF when the parsed text
// did, so editing a CRLF file does not turn it into LF.
func (d *Document) Bytes() []byte {
	var b strings.Builder
	d.write(&b, "")
	b.WriteString("\n")
	return []byte(lineendings.Convert(b.String(), d.crlf))
}

func (d *Document) write(b *strings.Builder, indent string) {
	if len(d.keys) == 0 {
		b.WriteString("{}")
		return
	}
	b.WriteString("{\n")
	inner := indent + "  "
	for i, k := range d.keys {
		b.WriteString(inner + quote(k) + ": ")
		switch v := d.values[k].(type) {
		case *Document:
			v.write(b, inner)
		case json.RawMessage:
			b.WriteString(reindent(v, inner))
		}
		if i < len(d.keys)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(indent + "}")
}

// reindent formats an array (or leaf) at the given indentation.
func reindent(raw json.RawMessage, indent string) string {
	if len(raw) == 0 || raw[0] != '[' {
		return string(raw)
	}
	var out bytes.Buffer
	if err := json.Indent(&out, raw, indent, "  "); err != nil {
		return string(raw)
	}
	return out.String()
}

func quote(s string) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	return strings.TrimRight(b.String(), "\n")
}
