package languageserver

import (
	"strings"
)

// Complete returns the suggestions for the position in a smith.toml text: the
// table names after "[", the keys of the current table, or the allowed values
// after "=". It returns an empty list where the schema has nothing to offer.
func Complete(text string, pos Position) []CompletionItem {
	items := []CompletionItem{}
	root, err := LoadSchema()
	if err != nil {
		return items
	}
	c := locate(text, pos)
	switch c.kind {
	case cursorHeader, cursorKey:
		parent := root.Resolve(c.path[:len(c.path)-1])
		if parent == nil {
			return items
		}
		for _, name := range parent.PropNames() {
			if !strings.HasPrefix(name, c.prefix) {
				continue
			}
			child := parent.element().Props[name]
			if c.kind == cursorHeader {
				if child.IsTable() {
					items = append(items, keyItem(name, child, CompletionKindModule, name))
				}
				continue
			}
			if child.IsTable() {
				items = append(items, keyItem(name, child, CompletionKindModule, name))
			} else {
				items = append(items, keyItem(name, child, CompletionKindProperty, name+" = "))
			}
		}
	case cursorValue:
		node := root.Resolve(c.path)
		if node == nil {
			return items
		}
		for _, v := range valueChoices(node) {
			items = append(items, CompletionItem{
				Label:      v,
				Kind:       CompletionKindValue,
				Detail:     node.Type,
				InsertText: insertValue(v, c.prefix),
			})
		}
	}
	return items
}

// keyItem builds the suggestion for one key of the schema.
func keyItem(label string, node *SchemaNode, kind int, insert string) CompletionItem {
	return CompletionItem{
		Label:         label,
		Kind:          kind,
		Detail:        node.Type,
		Documentation: node.Description,
		InsertText:    insert,
	}
}

// valueChoices lists the values a key takes: its enum, the enum of its array
// elements, or true and false for a boolean.
func valueChoices(node *SchemaNode) []string {
	switch {
	case len(node.Enum) > 0:
		return node.Enum
	case node.Items != nil && len(node.Items.Enum) > 0:
		return node.Items.Enum
	case node.Type == "boolean":
		return []string{"true", "false"}
	}
	return nil
}

// insertValue quotes a string value unless the person already typed a quote
// or the value is a boolean.
func insertValue(v, prefix string) string {
	if v == "true" || v == "false" || strings.ContainsAny(prefix, `"'`) {
		return v
	}
	return `"` + v + `"`
}
