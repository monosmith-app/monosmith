package release

import "github.com/monosmith-app/monosmith/apps/smith/projectgraph"

// BuildsAssetsFromCommand reports whether a project's release files come from
// its [release] build command, run on a runner per platform and uploaded by
// smith ci assets, rather than from its plugin.
func BuildsAssetsFromCommand(p *projectgraph.Project) bool {
	return p.Release != nil && p.Release.Build != ""
}
