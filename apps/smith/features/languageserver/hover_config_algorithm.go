package languageserver

import (
	"strings"
)

// HoverAt returns the description of the table or key under pos in a smith.toml
// text, or nil when the cursor is not on a key the schema knows.
func HoverAt(text string, pos Position) *Hover {
	root, err := LoadSchema()
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if pos.Line < 0 || pos.Line >= len(lines) {
		return nil
	}
	runes := []rune(lines[pos.Line])
	ch := max(0, min(pos.Character, len(runes)))
	onWord := (ch < len(runes) && isKeyChar(runes[ch])) || (ch > 0 && isKeyChar(runes[ch-1]))
	if !onWord {
		return nil
	}
	for ch < len(runes) && isKeyChar(runes[ch]) {
		ch++
	}
	c := locate(text, Position{Line: pos.Line, Character: ch})
	if c.kind != cursorHeader && c.kind != cursorKey {
		return nil
	}
	node := root.Resolve(c.path)
	if node == nil || (node.Description == "" && node.Type == "") {
		return nil
	}
	var b strings.Builder
	b.WriteString("**" + strings.Join(c.path, ".") + "**")
	if node.Type != "" {
		b.WriteString(" (" + node.Type + ")")
	}
	if node.Description != "" {
		b.WriteString("\n\n" + node.Description)
	}
	if len(node.Enum) > 0 {
		b.WriteString("\n\nOne of: " + strings.Join(node.Enum, ", "))
	}
	return &Hover{Contents: MarkupContent{Kind: "markdown", Value: b.String()}}
}
