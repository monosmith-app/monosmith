package doctorchecks

// The areas a workspace-wide finding is filed under, when it is about no one
// project. Each is one lowercase word.
const (
	// AreaToolchain is tools on PATH, versions, node_modules, venvs, runtimes.
	AreaToolchain = "toolchain"
	// AreaCI is pipelines, CI pins and workflows.
	AreaCI = "ci"
	// AreaLint is workspace-wide lint, format, test and build task and config checks.
	AreaLint = "lint"
	// AreaDependencies is lockfiles, overrides, audit and one-version checks.
	AreaDependencies = "dependencies"
	// AreaRelease is registries, tags, lineage, stable_only and release tooling.
	AreaRelease = "release"
	// AreaGit is ignore files, tracked files, commits and shallow clones.
	AreaGit = "git"
	// AreaConfig is smith.toml, defaults, cache, plugins, agent files and the rest.
	AreaConfig = "config"
)

// Areas is the closed set of areas a workspace-wide finding may carry. A
// finding that has a Project has none of them.
var Areas = []string{AreaToolchain, AreaCI, AreaLint, AreaDependencies, AreaRelease, AreaGit, AreaConfig}

// ValidArea reports whether area is one of Areas.
//
// Parameters: area is the word a finding carries.
// Returns: true when the word is in the closed set.
func ValidArea(area string) bool {
	for _, a := range Areas {
		if a == area {
			return true
		}
	}
	return false
}
