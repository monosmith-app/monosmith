package adoptmove

import (
	"path"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/adoptsplit"
)

// Role says where a top-level entry of the repository goes.
type Role int

const (
	// RoleProject is an entry that belongs to the project and moves with it.
	RoleProject Role = iota
	// RoleRepository is an entry that belongs to the whole repository, or is
	// workspace tool configuration, and stays at the root.
	RoleRepository
	// RoleDocs is a document of the project that goes to the workspace's doc
	// folder.
	RoleDocs
)

// DocsDir is the workspace folder the project's documents go to.
const DocsDir = "doc"

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
	"codeowners": true, "docs": true, "doc": true, "docker-compose.yml": true, "docker-compose.yaml": true,
	"compose.yml": true, "compose.yaml": true,
}

// repositoryPrefixes are repository-wide documents, whatever their extension
// (README.md, LICENSE-MIT, CHANGELOG.md): a change log of one project is made
// again per project by smith, and the documents here describe the repository.
var repositoryPrefixes = []string{"readme", "license", "licence", "copying", "notice", "contributing", "code_of_conduct", "security", "changelog", "authors", "docker-compose."}

// documentExtensions are the files of a project that are documents: they go to
// the doc folder. README, LICENSE and the other repository documents keep
// their names at the root (repositoryPrefixes).
var documentExtensions = map[string]bool{".md": true, ".markdown": true, ".mdx": true, ".rst": true, ".adoc": true}

// documentFolders are folders of documents that go to the doc folder whole.
var documentFolders = map[string]bool{"documentation": true, "wiki": true}

// Classify says where the top-level entry name goes for the ecosystem. What
// the repository or the language names as the workspace's stays at the root:
// version control, CI, editor settings, the .code-workspace file, repository
// documents, and the tool configuration and lockfiles the language declares
// (adoptsplit.Rules). Documents go to the doc folder. Everything else moves,
// because a project's own folders and config files have no fixed names (src,
// lib, app, tests, .env, jest.config.js, Dockerfile, ...). The language's
// Split then cuts the manifest and tsconfig.json that move.
func Classify(ecosystem, name string) Role {
	lower := strings.ToLower(name)
	if repositoryEntries[lower] || strings.HasSuffix(lower, ".code-workspace") {
		return RoleRepository
	}
	for _, prefix := range repositoryPrefixes {
		if strings.HasPrefix(lower, prefix) {
			return RoleRepository
		}
	}
	if adoptsplit.For(ecosystem).StaysAtRoot(lower) {
		return RoleRepository
	}
	if documentExtensions[path.Ext(lower)] || documentFolders[lower] {
		return RoleDocs
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
