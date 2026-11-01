package release

import "github.com/monosmith-app/monosmith/apps/smith/projectgraph"

// BuildsAssetsFromCommand reports whether a project's release files come from
// its [release] build command, run on a runner per platform and uploaded by
// smith ci assets, rather than from its plugin.
func BuildsAssetsFromCommand(p *projectgraph.Project) bool {
	return p.Release != nil && p.Release.Build != ""
}

// SkipsAssetsInstall reports whether a project opted out of being installed by
// the assets job ([release] install = false): its build command needs none of
// the project's dependencies, so smith ci setup --assets leaves it out of what
// the plugins install.
func SkipsAssetsInstall(p *projectgraph.Project) bool {
	return BuildsAssetsFromCommand(p) && p.Release.Install != nil && !*p.Release.Install
}
