package pipeline

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/libs/plugin-node/manifestfixes"
	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckTsconfigChain checks that the tsconfig.json of a project reaches the
// workspace's tsconfig.base.json through its "extends" chain, so one edit of the
// base reaches every project. It has nothing to say when the workspace has no
// base, the project has no tsconfig.json, the file cannot be read as JSON, or it
// is a solution file (no files, only references). It is a warning: a project can
// be standalone on purpose, and the fix asks before it edits anything.
func CheckTsconfigChain(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	if _, err := os.Stat(filepath.Join(workspaceRoot, tsconfigchain.Base)); err != nil {
		return protocol.Finding{}, false
	}
	file := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "tsconfig.json")
	raw, err := os.ReadFile(file)
	if err != nil || isSolutionFile(raw) {
		return protocol.Finding{}, false
	}
	reaches, err := tsconfigchain.ReachesBase(workspaceRoot, file)
	if err != nil {
		return protocol.Finding{}, false
	}
	finding := protocol.Finding{
		Check:   p.Name + " tsconfig.json reaches tsconfig.base.json",
		Project: p.Name,
		OK:      reaches,
		Detail:  p.Root + "/tsconfig.json does not extend tsconfig.base.json, directly or through another file, so a change of the base (a compiler option, a paths entry) never reaches this project",
		Remedy:  "set \"extends\" in " + p.Root + "/tsconfig.json to \"" + tsconfigchain.Extends(p.Root, tsconfigchain.Base) + "\" (smith doctor --fix shows the edit and asks first), or leave the project standalone on purpose",
		Warning: true,
	}
	if !reaches {
		finding.Problem = "tsconfig.json does not extend tsconfig.base.json"
		finding.Fix = manifestfixes.TsconfigChainFix(workspaceRoot, p)
	}
	return finding, true
}

// isSolutionFile reports a tsconfig that only lists other tsconfig files
// ("files": [] and "references"), which compiles nothing itself.
func isSolutionFile(raw []byte) bool {
	var doc struct {
		Files      *[]string        `json:"files"`
		References []map[string]any `json:"references"`
	}
	if json.Unmarshal(jsondoc.StripComments(raw), &doc) != nil {
		return false
	}
	return doc.Files != nil && len(*doc.Files) == 0 && len(doc.References) > 0
}
