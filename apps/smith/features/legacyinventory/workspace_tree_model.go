package legacyinventory

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Tree is the workspace as a plan sees it: the files that exist, read from
// disk, with the writes and deletions the plan has made but not applied. A
// plan is computed step by step on a Tree, each step seeing what the earlier
// ones would leave.
type Tree struct {
	// Root is the workspace directory.
	Root string
	// Modules is the directory whose node_modules describe the installed
	// packages (the workspace itself unless it is a copy without them).
	Modules string

	files   map[string]bool
	written map[string]string
}

// NewTree makes a Tree of the given workspace-relative files.
func NewTree(root string, files []string) *Tree {
	t := &Tree{Root: root, Modules: root, files: map[string]bool{}, written: map[string]string{}}
	for _, f := range files {
		t.files[filepath.ToSlash(f)] = true
	}
	return t
}

// Clone copies the Tree so a step can change it without changing this one.
func (t *Tree) Clone() *Tree {
	c := &Tree{Root: t.Root, Modules: t.Modules, files: map[string]bool{}, written: map[string]string{}}
	for f := range t.files {
		c.files[f] = true
	}
	for f, s := range t.written {
		c.written[f] = s
	}
	return c
}

// Files lists the files, sorted.
func (t *Tree) Files() []string {
	out := make([]string, 0, len(t.files))
	for f := range t.files {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// Has reports whether a file exists.
func (t *Tree) Has(rel string) bool { return t.files[rel] }

// Read returns a file's content, "" when it is missing.
func (t *Tree) Read(rel string) string {
	if !t.files[rel] {
		return ""
	}
	if s, ok := t.written[rel]; ok {
		return s
	}
	raw, err := os.ReadFile(filepath.Join(t.Root, filepath.FromSlash(rel)))
	if err != nil {
		return ""
	}
	return string(raw)
}

// Write sets a file's content.
func (t *Tree) Write(rel, content string) {
	t.files[rel] = true
	t.written[rel] = content
}

// Delete removes a file; a folder path removes the files under it.
func (t *Tree) Delete(rel string) {
	delete(t.files, rel)
	delete(t.written, rel)
	prefix := rel + "/"
	for f := range t.files {
		if strings.HasPrefix(f, prefix) {
			delete(t.files, f)
			delete(t.written, f)
		}
	}
}

// Apply makes a fix's file edits.
func (t *Tree) Apply(fix protocol.Fix) {
	for _, e := range fix.Files {
		if e.Delete {
			t.Delete(e.Path)
		} else {
			t.Write(e.Path, e.Content)
		}
	}
}

// ManifestPaths lists every package.json outside node_modules, the root's
// first.
func (t *Tree) ManifestPaths() []string {
	var out []string
	for _, f := range t.Files() {
		if f == "package.json" || strings.HasSuffix(f, "/package.json") {
			if !strings.Contains("/"+f, "/node_modules/") {
				out = append(out, f)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i] == "package.json" && out[j] != "package.json" })
	return out
}
