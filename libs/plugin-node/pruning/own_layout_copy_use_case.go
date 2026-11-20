package pruning

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
)

// sourceExport is an exports entry that points at a TypeScript source.
var sourceExport = regexp.MustCompile(`^\./src/(.+)\.ts$`)

// copyOwnLayout copies a package that builds inside its own folder (ADR 0102)
// as npm packs it from that folder: its package.json at the root and the
// entries of its files list (dist/) where they are. Its main and exports
// paths, and a read of ../package.json from dist/, keep working. Its scripts
// go, and an exports entry that names a TypeScript source names its build.
func copyOwnLayout(project, dest string) error {
	raw, err := os.ReadFile(filepath.Join(project, "package.json"))
	if err != nil {
		return err
	}
	doc, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return err
	}
	var include []string
	doc.RewriteStrings([]string{"files"}, func(v string) string {
		if !strings.HasPrefix(v, "!") {
			include = append(include, v)
		}
		return v
	})
	if err := copyIncluded(project, dest, include); err != nil {
		return err
	}
	doc.Delete("scripts")
	if exports, ok := doc.String("exports"); ok {
		if m := sourceExport.FindStringSubmatch(exports); m != nil {
			if err := doc.SetJSON("exports", map[string]string{"types": "./dist/" + m[1] + ".d.ts", "default": "./dist/" + m[1] + ".js"}); err != nil {
				return err
			}
		}
	}
	return os.WriteFile(filepath.Join(dest, "package.json"), doc.Bytes(), 0o644)
}

// copyIncluded copies the package's folder, or only the entries of its files
// list and the README and LICENSE when it has one. A glob entry (dist/**/*.js)
// copies the folder before its first wildcard.
func copyIncluded(project, dest string, include []string) error {
	if len(include) == 0 {
		return copyTree(project, dest)
	}
	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return err
	}
	for _, entry := range include {
		rel := strings.TrimPrefix(entry, "./")
		if i := strings.IndexAny(rel, "*?[{"); i >= 0 {
			rel = rel[:i]
		}
		rel = strings.TrimSuffix(rel, "/")
		if rel == "" {
			return copyTree(project, dest)
		}
		if err := copyEntry(filepath.Join(project, filepath.FromSlash(rel)), filepath.Join(dest, filepath.FromSlash(rel))); err != nil {
			return err
		}
	}
	entries, err := os.ReadDir(project)
	if err != nil {
		return err
	}
	for _, e := range entries {
		upper := strings.ToUpper(e.Name())
		if !e.IsDir() && (strings.HasPrefix(upper, "README") || strings.HasPrefix(upper, "LICENSE") || strings.HasPrefix(upper, "LICENCE")) {
			if err := copyEntry(filepath.Join(project, e.Name()), filepath.Join(dest, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// copyEntry copies a file or a folder; a missing one is left out, as npm
// leaves it out of the package.
func copyEntry(from, to string) error {
	info, err := os.Stat(from)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	if info.IsDir() {
		return copyTree(from, to)
	}
	if err := os.MkdirAll(filepath.Dir(to), 0o755); err != nil {
		return err
	}
	content, err := os.ReadFile(from)
	if err != nil {
		return err
	}
	return os.WriteFile(to, content, info.Mode().Perm())
}
