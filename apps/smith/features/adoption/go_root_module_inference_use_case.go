package adoption

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// DeclaredProject is a project adopt declares in smith.toml: a folder of the
// root Go module that has no go.mod of its own (ADR 0052).
type DeclaredProject struct {
	Name string `json:"name"`
	Root string `json:"root"`
	Kind string `json:"kind"`
}

var (
	packageMain = regexp.MustCompile(`(?m)^package\s+main\b`)
	projectName = regexp.MustCompile(`[^A-Za-z0-9._-]+`)
	// libraryParents are the top-level folders whose children are libraries.
	libraryParents = []string{"libs", "pkg", "internal"}
	// skippedFolders hold no project of the repository's.
	skippedFolders = map[string]bool{"vendor": true, "node_modules": true, "testdata": true, "dist": true, "third_party": true}
)

// InferGoProjects finds the programs and libraries of a repository whose Go
// code is one module at the root: every folder with a main package is an app
// named after it, and every child of libs/, pkg/ and internal/ that holds Go
// code and no program is a lib. A folder with a go.mod of its own, and what is
// under it, is a project smith finds itself and is left alone. root is the
// repository's folder; programs and libraries are sorted by root. rootProgram
// is true when the module's folder itself holds a main package, which cannot be
// declared (move it to cmd/<name>).
func InferGoProjects(root string) (projects []DeclaredProject, rootProgram bool) {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return nil, false
	}
	mains := map[string]bool{}
	hasGo := map[string]bool{}
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			name := d.Name()
			if rel != "." && (strings.HasPrefix(name, ".") || skippedFolders[name]) {
				return filepath.SkipDir
			}
			if rel != "." {
				if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(rel, ".go") || strings.HasSuffix(rel, "_test.go") {
			return nil
		}
		dir := path.Dir(rel)
		hasGo[dir] = true
		if head, err := os.ReadFile(p); err == nil {
			if len(head) > 4096 {
				head = head[:4096]
			}
			if packageMain.Match(head) {
				mains[dir] = true
			}
		}
		return nil
	})
	taken := map[string]bool{}
	var apps []string
	for dir := range mains {
		if dir == "." {
			rootProgram = true
			continue
		}
		apps = append(apps, dir)
	}
	sort.Strings(apps)
	for _, dir := range apps {
		projects = append(projects, DeclaredProject{Name: uniqueName(dir, taken), Root: dir, Kind: "app"})
	}
	libs := map[string]bool{}
	for dir := range hasGo {
		parts := strings.Split(dir, "/")
		if len(parts) < 2 || !slicesContain(libraryParents, parts[0]) {
			continue
		}
		libs[parts[0]+"/"+parts[1]] = true
	}
	var libRoots []string
	for lib := range libs {
		libRoots = append(libRoots, lib)
	}
	sort.Strings(libRoots)
	for _, lib := range libRoots {
		if holdsProgram(lib, mains) {
			continue
		}
		projects = append(projects, DeclaredProject{Name: uniqueName(lib, taken), Root: lib, Kind: "lib"})
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Root < projects[j].Root })
	return projects, rootProgram
}

// holdsProgram says whether a main package is in the folder or below it.
func holdsProgram(dir string, mains map[string]bool) bool {
	for m := range mains {
		if m == dir || strings.HasPrefix(m, dir+"/") {
			return true
		}
	}
	return false
}

// uniqueName is the project's name: its folder's name, or when another project
// has taken it, the whole path joined with dashes.
func uniqueName(dir string, taken map[string]bool) string {
	name := projectName.ReplaceAllString(path.Base(dir), "-")
	if taken[name] {
		name = projectName.ReplaceAllString(strings.ReplaceAll(dir, "/", "-"), "-")
	}
	taken[name] = true
	return name
}

func slicesContain(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// declarationsText is the [projects.<name>] tables that declare the projects.
func declarationsText(projects []DeclaredProject) string {
	var b strings.Builder
	for _, p := range projects {
		b.WriteString("\n[projects." + p.Name + "]\nroot = \"" + p.Root + "\"\nlanguage = \"go\"\nkind = \"" + p.Kind + "\"\n")
	}
	return b.String()
}
