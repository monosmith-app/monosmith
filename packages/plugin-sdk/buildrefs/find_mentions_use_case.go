package buildrefs

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// workspaceFiles are the files at the workspace root that run or describe
// builds. A task of smith.toml, a root script, a Dockerfile or a pipeline can
// name the folder a tool builds into, and would break when the build moves.
var workspaceFiles = []string{"smith.toml", "package.json", "Makefile", "justfile", "Dockerfile", "compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}

// projectFileSuffixes are the files at the top of a project that run a build:
// a Dockerfile, a Makefile, a script, a manifest of scripts.
var projectFileSuffixes = []string{"Dockerfile", ".Dockerfile", "Makefile", ".sh", ".ps1", ".cmd", ".bat", ".yml", ".yaml", "package.json"}

// Mentions lists the files that name one of the old build folders of a
// project, so a move that would break them can be refused with the file named.
//
// folders are relative to the project ("target", "bin", "build"). A file
// counts when it names a folder (with a slash after it: target/release) from the workspace root (src/App/bin) or, for
// the files at the top of the project, from the project (bin/Release). Source
// files are not read: a tool's own configuration is read by its own check.
//
// Parameters: workspaceRoot is the absolute workspace folder; projectRoot the
// workspace-relative project folder ("." for the root); folders the old build
// folders. Returns the workspace-relative files, sorted, without duplicates.
func Mentions(workspaceRoot, projectRoot string, folders ...string) []string {
	found := map[string]bool{}
	named := func(text string, refs []string) bool {
		for _, ref := range refs {
			if regexp.MustCompile(`(?:^|[^\w./@\\-])` + regexp.QuoteMeta(ref) + `[/\\]`).MatchString(text) {
				return true
			}
		}
		return false
	}
	fromRoot := make([]string, 0, len(folders))
	for _, f := range folders {
		fromRoot = append(fromRoot, path.Clean(path.Join(projectRoot, f)))
	}
	scan := func(rel string, refs []string) {
		raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
		if err == nil && named(string(raw), refs) {
			found[rel] = true
		}
	}
	for _, f := range workspaceFiles {
		scan(f, fromRoot)
	}
	if workflows, err := os.ReadDir(filepath.Join(workspaceRoot, ".github", "workflows")); err == nil {
		for _, e := range workflows {
			scan(path.Join(".github/workflows", e.Name()), fromRoot)
		}
	}
	entries, err := os.ReadDir(filepath.Join(workspaceRoot, filepath.FromSlash(projectRoot)))
	if err != nil {
		return sorted(found)
	}
	for _, e := range entries {
		if e.IsDir() || !isProjectFile(e.Name()) {
			continue
		}
		rel := path.Join(projectRoot, e.Name())
		scan(rel, folders)
		scan(rel, fromRoot)
	}
	return sorted(found)
}

func isProjectFile(name string) bool {
	for _, s := range projectFileSuffixes {
		if strings.HasSuffix(name, s) || strings.HasPrefix(name, "Dockerfile") {
			return true
		}
	}
	return false
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
