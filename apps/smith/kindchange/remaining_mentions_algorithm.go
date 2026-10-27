package kindchange

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cipipeline"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/pathrefs"
)

// maxMentionBytes is the largest file looked through: a bigger one is data.
const maxMentionBytes = 1 << 20

// remainingMentions lists the workspace's files that still name the project's
// old folder after the plan's edits: a Dockerfile, a script, a document.
// smith does not know how to rewrite them, so it names them for the person.
// Files inside the folder move with it and are not listed.
func remainingMentions(root string, files []string, plan Plan) []string {
	edited := map[string]bool{}
	for _, e := range plan.Edits {
		edited[e.Path] = true
	}
	var found []string
	for _, f := range files {
		if edited[f] || f == plan.Root || strings.HasPrefix(f, plan.Root+"/") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(f)))
		if err != nil || len(raw) > maxMentionBytes || bytes.IndexByte(raw, 0) >= 0 {
			continue
		}
		if slices.Contains(cipipeline.OwnedPaths, f) && cipipeline.IsGenerated(string(raw)) {
			continue // smith regenerates its own pipelines
		}
		if _, n := pathrefs.ReplaceReference(string(raw), plan.Root, plan.NewRoot); n > 0 {
			found = append(found, f)
		}
	}
	sort.Strings(found)
	return found
}
