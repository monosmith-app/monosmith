package conformance

import (
	"path"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// checkRepair checks what every plugin with the repair capability does the
// same: a plan names a kind the protocol has, at most one plan has each kind,
// every plan does something, and every path it deletes stays inside the
// workspace. The plugin is asked about a project whose folder does not exist,
// so nothing of a real workspace is named.
func checkRepair(c *host.Client, r *Report) {
	if !c.Info.Capabilities.Repair {
		return
	}
	language := ""
	if len(c.Info.Languages) > 0 {
		language = c.Info.Languages[0]
	}
	gone := protocol.Project{Name: "conformance-gone", Root: "conformance/no-such-folder", Language: language, Kind: protocol.ProjectLib}
	var res protocol.RepairResult
	err := c.Call(protocol.MethodRepair, protocol.RepairParams{Projects: []protocol.Project{gone}}, &res, time.Minute)
	r.add("repair plans are given for a project with nothing installed", err == nil, "dependencies/repair only describes: answer with the plans (or none), never with an error for a folder that is not there")
	if err != nil {
		return
	}
	known := map[protocol.RepairKind]bool{}
	for _, k := range protocol.RepairKinds() {
		known[k] = true
	}
	seen := map[protocol.RepairKind]bool{}
	kinds, once, acts, inside := true, true, true, true
	for _, p := range res.Plans {
		kinds = kinds && known[p.Kind]
		once = once && !seen[p.Kind]
		seen[p.Kind] = true
		acts = acts && p.Summary != "" && (len(p.Delete) > 0 || len(p.Commands) > 0 || p.Install)
		for _, d := range p.Delete {
			inside = inside && InsideWorkspace(d)
		}
		for _, cmd := range p.Commands {
			acts = acts && len(cmd.Argv) > 0
			inside = inside && (cmd.Dir == "" || InsideWorkspace(cmd.Dir))
		}
	}
	r.add("every repair plan names a repair kind of the protocol", kinds, "use prune, reinstall, lockfile or cache")
	r.add("a repair kind has one plan", once, "merge what the kind deletes and runs into one plan")
	r.add("every repair plan has a summary and does something", acts, "give Summary, and at least one of Delete, Commands or Install; leave out a kind the tool has no equivalent for")
	r.add("a repair plan deletes only inside the workspace", inside, "give workspace-relative paths with forward slashes, never the workspace root, an absolute path or one with ..")
}

// InsideWorkspace reports whether a path of a repair plan is a clean
// workspace-relative path below the workspace root: not empty, not the root,
// not absolute, with forward slashes and no ".." step.
//
// Parameters: rel is the path as the plugin gave it.
//
// Returns: true when the host may act on the path.
func InsideWorkspace(rel string) bool {
	if rel == "" || strings.ContainsAny(rel, "\\:") || strings.HasPrefix(rel, "/") {
		return false
	}
	clean := path.Clean(rel)
	return clean == rel && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}
