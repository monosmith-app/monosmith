package pipeline

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/libs/plugin-rust/releasing"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// releaseRunner is the platform of the runner that releases: GitHub's
// ubuntu-latest, and the one an Azure or GitLab pipeline uses.
const releaseRunner = "linux/amd64"

// releaseTargets are the Rust targets the release cross-compiles apps for,
// and the packages of the linkers it needs for them: the platforms of every
// app that is not native, other than the runner's own, that the runner can
// build. A native app is built on its own runner, which has its own target.
func releaseTargets(projects []protocol.Project) (targets, linkerPackages []string) {
	for _, p := range projects {
		if p.ReleaseAssets == nil || slices.Contains(p.Tags, discovery.NativeTag) {
			continue
		}
		for _, platform := range p.ReleaseAssets.Platforms {
			triple, err := releasing.Target(platform)
			reach := releasing.CrossReach(releaseRunner, platform)
			if err != nil || platform == releaseRunner || reach == releasing.OnItsOwnSystem {
				continue
			}
			if !slices.Contains(targets, triple) {
				targets = append(targets, triple)
			}
			if reach == releasing.WithLinker {
				if _, pkg := releasing.Linker(platform); !slices.Contains(linkerPackages, pkg) {
					linkerPackages = append(linkerPackages, pkg)
				}
			}
		}
	}
	slices.Sort(targets)
	slices.Sort(linkerPackages)
	return targets, linkerPackages
}
