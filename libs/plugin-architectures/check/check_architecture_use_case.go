package check

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/imports"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
)

// skippedFolders hold no hand-written code, or only test data, in any language.
var skippedFolders = map[string]bool{
	"vendor": true, "node_modules": true, "dist": true, "build": true, "target": true, "bin": true, "obj": true,
	"__pycache__": true, "venv": true, "testdata": true, "fixtures": true, "coverage": true,
	"test": true, "tests": true, "__tests__": true, "spec": true, "specs": true, "e2e": true, "testing": true,
}

// source is one file the check reads.
type source struct {
	rel  string
	segs []string
}

// Check walks the project in dir and reports the imports its language's
// sources make that the architecture forbids.
func Check(dir, language string, rs rules.Ruleset) (Report, error) {
	var r Report
	lang, ok := imports.For(language)
	if !ok {
		return r, fmt.Errorf("the architecture check knows %s, not %q", strings.Join(imports.Names(), ", "), language)
	}
	files, layerFolders, sets, err := survey(dir, lang, rs)
	if err != nil {
		return r, err
	}
	if len(files) == 0 {
		return r, nil
	}
	if len(layerFolders) == 0 && !hasModules(files, rs) {
		r.warn("(project)", fmt.Sprintf("no folder is named for a part of the %s architecture (%s), so none of its rules were applied", rs.Name, strings.Join(rs.FolderNames(), ", ")))
		return r, nil
	}
	project := imports.Detect(dir, language)
	for _, f := range files {
		raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(f.rel)))
		if err != nil {
			return r, err
		}
		from := classify(rs, f.segs[:len(f.segs)-1], true, sets)
		for _, ref := range lang.Extract(project, f.rel, string(raw)) {
			to := classify(rs, ref.Segments, false, sets)
			if why := forbiddenImport(rs, from, to, sets); why != "" {
				r.problem(f.rel, fmt.Sprintf("imports %s: %s", ref.Spec, why))
			}
		}
	}
	return r, nil
}

// LayerFolders lists the folders of the project in dir that are named for a
// layer or, for a modular monolith, hold the modules.
func LayerFolders(dir, language string, rs rules.Ruleset) ([]string, error) {
	lang, ok := imports.For(language)
	if !ok {
		return nil, fmt.Errorf("the architecture check knows %s, not %q", strings.Join(imports.Names(), ", "), language)
	}
	files, folders, _, err := survey(dir, lang, rs)
	if err != nil || len(files) == 0 {
		return folders, err
	}
	if len(folders) == 0 && hasModules(files, rs) {
		return []string{"modules"}, nil
	}
	return folders, nil
}

func hasModules(files []source, rs rules.Ruleset) bool {
	if !rs.Modules {
		return false
	}
	return slices.ContainsFunc(files, func(f source) bool {
		return slices.ContainsFunc(f.segs[:len(f.segs)-1], func(s string) bool { return rules.Normalise(s) == "modules" })
	})
}

// survey lists the source files of the project, the folders named for a
// layer, and the folders that sit below layers and modules.
func survey(dir string, lang imports.Language, rs rules.Ruleset) ([]source, []string, folderSets, error) {
	sets := folderSets{layerChildren: map[string]bool{}, moduleChildren: map[string]bool{}}
	var files []source
	var layerFolders []string
	err := filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel == "." {
				return nil
			}
			name := d.Name()
			if strings.HasPrefix(name, ".") || skippedFolders[name] {
				return filepath.SkipDir
			}
			parent := path.Base(path.Dir(rel))
			if path.Dir(rel) != "." {
				_, inLayer := rs.LayerOf(parent)
				grand := path.Base(path.Dir(path.Dir(rel)))
				if _, ok := rs.LayerOf(grand); !inLayer && directions[rules.Normalise(parent)] && ok {
					inLayer = true
				}
				if inLayer {
					sets.layerChildren[rules.Normalise(name)] = true
				}
				if rs.Modules && path.Base(path.Dir(path.Dir(rel))) != "." && rules.Normalise(path.Base(path.Dir(path.Dir(rel)))) == "modules" {
					sets.moduleChildren[rules.Normalise(name)] = true
				}
			}
			if _, ok := rs.LayerOf(name); ok {
				layerFolders = append(layerFolders, rel)
			}
			return nil
		}
		name := d.Name()
		if !slices.Contains(lang.Extensions, path.Ext(name)) || lang.IsTest(name) {
			return nil
		}
		files = append(files, source{rel: rel, segs: split(rel)})
		return nil
	})
	return files, layerFolders, sets, err
}

func split(rel string) []string { return strings.Split(rel, "/") }
