package apisurface

import (
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/gitrevision"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/toolrun"
)

// Source reads the files of a workspace at one revision. Paths are
// workspace-relative and use forward slashes.
type Source interface {
	// Read returns the text of a file, and false when the revision has none.
	Read(rel string) (string, bool)
}

// NewSource returns the source for a workspace root: the working tree when
// ref is empty, else the files git holds at that revision.
func NewSource(root, ref string) Source {
	if ref == "" {
		return worktreeSource{root: root}
	}
	return &gitSource{root: root, ref: ref, batch: gitrevision.NewReader(root, ref)}
}

// worktreeSource reads the files on disk.
type worktreeSource struct{ root string }

// Read implements Source.
func (w worktreeSource) Read(rel string) (string, bool) {
	body, err := os.ReadFile(filepath.Join(w.root, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	return string(body), true
}

// gitSource reads the files with git show, so no checkout is needed.
type gitSource struct {
	root, ref string
	batch     *gitrevision.Reader
}

// Read implements Source. The "./" makes the path relative to the workspace
// root, which need not be the top of the repository.
func (g *gitSource) Read(rel string) (string, bool) {
	if body, ok := g.batch.Read(rel); ok || !g.batch.Failed() {
		return body, ok
	}
	out, _, err := toolrun.Capture(g.root, []string{"git", "show", g.ref + ":./" + rel})
	if err != nil {
		return "", false
	}
	return out, true
}

// cachedSource remembers what it read, because a re-export chain reads the
// same file more than once.
type cachedSource struct {
	inner Source
	seen  map[string]cachedFile
}

type cachedFile struct {
	body string
	ok   bool
}

// Read implements Source.
func (c *cachedSource) Read(rel string) (string, bool) {
	if f, ok := c.seen[rel]; ok {
		return f.body, f.ok
	}
	body, ok := c.inner.Read(rel)
	c.seen[rel] = cachedFile{body, ok}
	return body, ok
}

// Close ends the git process the source reads through.
func (g *gitSource) Close() error { return g.batch.Close() }

// Close ends what the source read through, when it holds a process.
func (c *cachedSource) Close() error {
	if closer, ok := c.inner.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
