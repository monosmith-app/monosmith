package swapchoice

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node-toolchains/tooling"
)

// testModules are the modules a test file imports its API from, for each test
// runner plugin that smith can swap by itself.
var testModules = map[string]string{"jest": "@jest/globals", "vitest": "vitest"}

// commonTestNames are the names both runners export, so an import of only
// these moves from one module to the other without any other edit.
var commonTestNames = map[string]bool{
	"describe": true, "it": true, "test": true, "expect": true,
	"beforeEach": true, "afterEach": true, "beforeAll": true, "afterAll": true,
}

// Migration is what moving a project from one toolchain plugin to another
// changes, worked out before anything is written.
type Migration struct {
	// Dependencies leave the root package.json devDependencies. Kept are those
	// the old plugin added that stay because another project still uses it.
	Dependencies, Kept []string
	// Scripts and Fields leave the project's package.json. Files are
	// project-relative files that leave: they hold exactly what the old plugin
	// wrote.
	Scripts, Fields, Files []string
	// Rewrites are the test files whose import moves to the new module, with
	// the new content (project-relative path).
	Rewrites map[string]string
	// ByHand says what smith leaves to the person, one line each.
	ByHand []string
}

// Empty reports whether the migration changes nothing.
func (m Migration) Empty() bool {
	return len(m.Dependencies) == 0 && len(m.Scripts) == 0 && len(m.Fields) == 0 && len(m.Files) == 0 && len(m.Rewrites) == 0
}

// RewriteNames are the paths of the rewritten test files, sorted.
func (m Migration) RewriteNames() []string {
	names := make([]string, 0, len(m.Rewrites))
	for name := range m.Rewrites {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PlanMigration works out how to take a project from the old toolchain plugin's
// setup to the new one's. It reads the project and writes nothing. Only what
// the old plugin wrote and the project still has as it was written is removed:
// a script, field or file the person edited stays and is named in ByHand.
//
// Parameters: root is the workspace root; dir is the project's directory
// (workspace-relative); old and next are the two plugins' setups for the
// project; fromModule and toModule are the modules test files import from;
// shared is true when another project still uses the old plugin, which keeps
// the root dependencies in place.
// Returns: the migration, or the error that stopped reading the project.
func PlanMigration(root, dir string, old, next tooling.Setup, fromModule, toModule string, shared bool) (Migration, error) {
	m := Migration{Rewrites: map[string]string{}}
	project := filepath.Join(root, filepath.FromSlash(dir))
	manifest, err := readJSON(filepath.Join(project, "package.json"))
	if err != nil {
		return m, err
	}
	rootManifest, err := readJSON(filepath.Join(root, "package.json"))
	if err != nil {
		return m, err
	}
	have, _ := rootManifest["devDependencies"].(map[string]any)
	for _, name := range sortedKeys(old.DevDependencies) {
		if _, kept := next.DevDependencies[name]; kept {
			continue
		}
		if _, there := have[name]; !there {
			continue
		}
		if shared {
			m.Kept = append(m.Kept, name)
		} else {
			m.Dependencies = append(m.Dependencies, name)
		}
	}
	scripts, _ := manifest["scripts"].(map[string]any)
	for _, name := range sortedKeys(old.Scripts) {
		value, there := scripts[name]
		switch {
		case !there:
		case value == old.Scripts[name]:
			m.Scripts = append(m.Scripts, name)
		default:
			m.ByHand = append(m.ByHand, "package.json script "+name+" is "+quote(value)+", not what the old runner wrote: set it to "+quote(next.Scripts[name])+" yourself")
		}
	}
	for _, name := range sortedKeys(old.Fields) {
		value, there := manifest[name]
		if !there {
			continue
		}
		if _, replaced := next.Fields[name]; replaced {
			continue
		}
		if sameJSON(value, old.Fields[name]) {
			m.Fields = append(m.Fields, name)
		} else {
			m.ByHand = append(m.ByHand, "package.json field "+name+" has your edits, so it stays: move them into the new runner's configuration")
		}
	}
	for _, name := range sortedKeys(old.Files) {
		if _, again := next.Files[name]; again {
			continue
		}
		body, err := os.ReadFile(filepath.Join(project, filepath.FromSlash(name)))
		switch {
		case err != nil:
		case string(body) == old.Files[name]:
			m.Files = append(m.Files, name)
		default:
			m.ByHand = append(m.ByHand, name+" has your edits, so it stays: move them into the new runner's configuration")
		}
	}
	err = planTestFiles(&m, project, fromModule, toModule)
	return m, err
}

// planTestFiles finds the test files that use the old runner's API: an import
// of only common names moves to the new module; anything else is left to the
// person.
func planTestFiles(m *Migration, project, fromModule, toModule string) error {
	importOf := regexp.MustCompile(`import\s*\{([^}]*)\}\s*from\s*(['"])` + regexp.QuoteMeta(fromModule) + `(['"])`)
	apiUse := regexp.MustCompile(`\bjest\.`)
	if fromModule == "vitest" {
		apiUse = regexp.MustCompile(`\bvi\.`)
	}
	api := map[bool]string{true: "jest.", false: "vi."}[fromModule != "vitest"]
	return filepath.WalkDir(project, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case "node_modules", "dist", "build", "coverage", ".git":
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".ts", ".tsx", ".js", ".jsx", ".mts", ".cts", ".mjs", ".cjs":
		default:
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		body := string(raw)
		rel, _ := filepath.Rel(project, path)
		rel = filepath.ToSlash(rel)
		imported := importOf.FindStringSubmatch(body)
		if imported == nil && !apiUse.MatchString(body) {
			return nil
		}
		movable := imported != nil && !apiUse.MatchString(body)
		if imported != nil {
			for _, name := range strings.Split(imported[1], ",") {
				if name = strings.TrimSpace(name); name != "" && !commonTestNames[strings.Fields(name)[0]] {
					movable = false
				}
			}
		}
		if movable {
			m.Rewrites[rel] = importOf.ReplaceAllString(body, "import {$1} from "+"${2}"+toModule+"${3}")
			return nil
		}
		m.ByHand = append(m.ByHand, rel+" uses "+api+" or helpers beyond describe, it and expect: change it to the new runner's API (jest.fn is vi.fn, jest.mock is vi.mock)")
		return nil
	})
}

func readJSON(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// sameJSON compares two values as JSON documents.
func sameJSON(a, b any) bool {
	var x, y any
	ra, _ := json.Marshal(a)
	rb, _ := json.Marshal(b)
	return json.Unmarshal(ra, &x) == nil && json.Unmarshal(rb, &y) == nil && reflect.DeepEqual(x, y)
}

func quote(v any) string {
	raw, _ := json.Marshal(v)
	return string(raw)
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
