package agentfiles

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Tool is one MCP tool as the agent text lists it.
type Tool struct {
	Name, Summary string
	ReadOnly      bool
}

// markerPrefix starts the marker line of the agent file.
const markerPrefix = "<!-- smith:generated "

// RenderAgent is the agent file for these tools: the frontmatter, the marker,
// the text, and the two lists of tools in the order given.
func RenderAgent(tools []Tool) string {
	var b strings.Builder
	b.WriteString(agentIntro)
	list := func(title string, readOnly bool) {
		b.WriteString("\n" + title + "\n\n")
		for _, t := range tools {
			if t.ReadOnly == readOnly {
				b.WriteString("- " + t.Name + ": " + t.Summary + "\n")
			}
		}
	}
	list("Read-only tools, use them freely:", true)
	list("Tools that change things, ask first (smith_run is routine after a change: run lint, typecheck and test without asking, and ask before any other task):", false)
	b.WriteString(agentHabits)
	return withMarker(b.String())
}

// withMarker puts the marker after the frontmatter. The hash covers the text
// without the marker line.
func withMarker(body string) string {
	plain := agentFrontmatter + body
	return agentFrontmatter + markerPrefix + "sha256=" + digest(plain) +
		" by smith agent install. Delete this line to keep your own edits; smith then leaves the file alone. -->\n" + body
}

func digest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])[:12]
}

// unmarked returns a file's text without its marker line and the hash the
// marker recorded; ok is false when there is no marker.
func unmarked(file string) (plain, recorded string, ok bool) {
	file = strings.ReplaceAll(file, "\r\n", "\n")
	lines := strings.SplitAfter(file, "\n")
	for i, l := range lines {
		if !strings.HasPrefix(l, markerPrefix+"sha256=") {
			continue
		}
		recorded, _, _ = strings.Cut(strings.TrimPrefix(l, markerPrefix+"sha256="), " ")
		return strings.Join(lines[:i], "") + strings.Join(lines[i+1:], ""), recorded, true
	}
	return file, "", false
}
