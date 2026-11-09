package removeproject

import (
	"sort"
	"strings"
	"unicode"
)

// Mention kinds, in the order the report groups them.
const (
	MentionWorkflow = "workflow"
	MentionEditor   = "editor"
	MentionOther    = "other"
)

// Mention is the first line of a tracked file that still names a removed project.
type Mention struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Kind string `json:"kind"`
}

// hasPathBoundary reports whether line holds root not followed by a name character,
// so that apps/web does not match apps/web-admin.
func hasPathBoundary(line, root string) bool {
	rest := line
	for {
		i := strings.Index(rest, root)
		if i < 0 {
			return false
		}
		rest = rest[i+len(root):]
		if rest == "" {
			return true
		}
		c := rune(rest[0])
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' {
			return true
		}
	}
}

// mentionKind sorts a path into workflow, editor or other.
func mentionKind(path string) string {
	p := "/" + strings.ToLower(path)
	base := p[strings.LastIndex(p, "/")+1:]
	switch {
	case strings.Contains(p, "/.github/workflows/"), strings.HasPrefix(base, "azure-pipelines"), strings.HasPrefix(base, ".gitlab-ci"):
		return MentionWorkflow
	case strings.HasSuffix(base, ".code-workspace"), strings.Contains(p, "/.vscode/"), strings.Contains(p, "/.idea/"):
		return MentionEditor
	}
	return MentionOther
}

// GroupMentions orders mentions by kind (workflow, editor, other), then by path.
func GroupMentions(in []Mention) []Mention {
	rank := map[string]int{MentionWorkflow: 0, MentionEditor: 1, MentionOther: 2}
	out := append([]Mention(nil), in...)
	sort.Slice(out, func(i, j int) bool {
		if rank[out[i].Kind] != rank[out[j].Kind] {
			return rank[out[i].Kind] < rank[out[j].Kind]
		}
		return out[i].File < out[j].File
	})
	return out
}
