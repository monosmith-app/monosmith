package adoptmove

import (
	"path"
	"sort"
	"strings"
)

// Role says where a top-level entry of the repository goes.
type Role int

const (
	// RoleProject is an entry that belongs to the project and moves with it.
	RoleProject Role = iota
	// RoleRepository is an entry that belongs to the whole repository and stays.
	RoleRepository
)

// repositoryEntries are the top-level names that describe the repository, not
// the project: version control, CI, editor and agent settings, smith's own
// files, repository-wide docs and policy files, and the tools that run on
// commits. They stay at the root whatever the ecosystem. Matching is
// case-insensitive.
var repositoryEntries = map[string]bool{
	".git": true, ".gitignore": true, ".gitattributes": true, ".gitmodules": true,
	".github": true, ".gitlab": true, ".gitlab-ci.yml": true, ".circleci": true, ".travis.yml": true,
	".azure-pipelines": true, "azure-pipelines.yml": true, "jenkinsfile": true,
	".devcontainer": true, ".vscode": true, ".idea": true, ".editorconfig": true,
	".claude": true, ".mcp.json": true, "claude.md": true, "agents.md": true,
	".smith": true, "smith.toml": true,
	".husky": true, ".pre-commit-config.yaml": true,
	".commitlintrc": true, ".commitlintrc.json": true, ".commitlintrc.yml": true, ".commitlintrc.yaml": true,
	"commitlint.config.js": true, "commitlint.config.mjs": true, "commitlint.config.cjs": true, "commitlint.config.ts": true,
	".npmrc": true, ".nvmrc": true, ".node-version": true, ".tool-versions": true,
	"renovate.json": true, ".renovaterc": true, ".renovaterc.json": true,
	"codeowners": true, "docs": true, "docker-compose.yml": true, "docker-compose.yaml": true,
	"compose.yml": true, "compose.yaml": true,
}

// repositoryPrefixes are repository-wide documents, whatever their extension
// (README.md, LICENSE-MIT, CHANGELOG.md): a change log of one project is made
// again per project by smith, and the documents here describe the repository.
var repositoryPrefixes = []string{"readme", "license", "licence", "copying", "notice", "contributing", "code_of_conduct", "security", "changelog", "authors", "docker-compose."}

// foreignLockfiles are the lockfiles of the package managers smith does not use
// for Node: they stay at the root and the move tells the person to write
// package-lock.json. package-lock.json stays too: with npm workspaces the one
// lockfile lives at the root, and npm install rewrites it for the new layout.
var foreignLockfiles = map[string]bool{"yarn.lock": true, "pnpm-lock.yaml": true, "bun.lockb": true, "bun.lock": true, "package-lock.json": true, "npm-shrinkwrap.json": true}

// Classify says where the top-level entry name goes for the ecosystem.
// Everything not named as the repository's moves, because a project's own
// folders and config files have no fixed names (src, lib, app, tests, .env,
// tsconfig.json, jest.config.js, Dockerfile, ...). The Node lockfiles stay.
func Classify(ecosystem, name string) Role {
	lower := strings.ToLower(name)
	if repositoryEntries[lower] {
		return RoleRepository
	}
	for _, prefix := range repositoryPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return RoleRepository
		}
	}
	if ecosystem == EcosystemNode && foreignLockfiles[lower] {
		return RoleRepository
	}
	return RoleProject
}

// TopLevel lists the first path segment of each file, sorted and without
// repeats. files are repository-relative with forward slashes.
func TopLevel(files []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range files {
		first, _, _ := strings.Cut(path.Clean(f), "/")
		if first != "" && !seen[first] {
			seen[first] = true
			out = append(out, first)
		}
	}
	sort.Strings(out)
	return out
}
