package turboworkspace

import "strings"

// DetectManager decides which package manager a workspace uses: the root
// package.json `packageManager` field, else the lockfile, else
// pnpm-workspace.yaml, else npm.
func DetectManager(packageManager string, files []string) Manager {
	name, _, _ := strings.Cut(packageManager, "@")
	switch Manager(name) {
	case PNPM, Yarn, NPM, Bun:
		return Manager(name)
	}
	has := map[string]bool{}
	for _, f := range files {
		has[f] = true
	}
	switch {
	case has["pnpm-lock.yaml"]:
		return PNPM
	case has["yarn.lock"]:
		return Yarn
	case has["bun.lock"] || has["bun.lockb"]:
		return Bun
	case has["package-lock.json"]:
		return NPM
	case has["pnpm-workspace.yaml"]:
		return PNPM
	}
	return NPM
}
