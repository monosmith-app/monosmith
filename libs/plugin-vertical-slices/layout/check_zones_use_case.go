package layout

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/imports"
	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
)

const (
	features     = "features"
	integrations = "integrations"
	shared       = "shared"
)

// pureRoles hold no I/O: they import no integration and no client,
// repository or store. The type allowance of a handler, mapper, algorithm or
// contract needs type-only imports, which only TypeScript has, so those roles
// are not checked here.
var pureRoles = map[string]bool{"model": true, "policy": true, "validator": true, "error": true, "enum": true}

// ioRoles do I/O.
var ioRoles = map[string]bool{"client": true, "repository": true, "store": true}

// testFolders hold tests, not slices.
var testFolders = map[string]bool{"test": true, "tests": true, "spec": true, "testing": true}

// checkZones holds a zoned project to its zones: every slice under the
// source root sits in features/, integrations/ or shared/; shared and
// integrations import no feature; a model, policy, validator, error or enum
// imports no integration and no I/O role; a listed vendor package is
// imported only in integrations/. A project is zoned when it is an app (under
// an apps folder) or its source root holds a zone folder.
func checkZones(r *Report, dir string, rules languages.Rules, vendors []string) error {
	root := sourceRoot(dir, rules)
	base := filepath.Join(dir, filepath.FromSlash(root))
	if !isApp(dir) && !hasZone(base) {
		return nil
	}
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	lang, readable := imports.For(rules.Name)
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || slices.Contains(languages.Zones, name) || strings.HasPrefix(name, ".") || skippedFolders[name] || testFolders[name] {
			continue
		}
		if holdsSource(filepath.Join(base, name), rules, lang, readable) {
			r.problem(path.Join(root, name)+"/", "an app keeps its slices in features/, integrations/ or shared/: move it into one of them (features for a business outcome, integrations for an external system, shared for a building block with no business knowledge)")
		}
	}
	if !readable {
		return nil
	}
	project := imports.Detect(dir, rules.Name)
	return filepath.WalkDir(base, func(file string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if file != base && (strings.HasPrefix(name, ".") || skippedFolders[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		if !slices.Contains(lang.Extensions, path.Ext(name)) {
			return nil
		}
		fromRoot, _ := filepath.Rel(base, file)
		parts := strings.Split(filepath.ToSlash(fromRoot), "/")
		zone := ""
		if len(parts) > 1 && slices.Contains(languages.Zones, parts[0]) {
			zone = parts[0]
		}
		raw, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		text := string(raw)
		fromProject, _ := filepath.Rel(dir, file)
		shown := filepath.ToSlash(fromProject)
		test := lang.IsTest(name) || slices.ContainsFunc(parts[:len(parts)-1], func(p string) bool { return testFolders[p] })
		role := languages.RoleOf(name)
		for _, ref := range lang.Extract(project, shown, text) {
			target := zoneIn(ref.Segments)
			if target == features && (zone == shared || zone == integrations) {
				r.problem(shown, fmt.Sprintf("%s/ holds no business code, so it never imports a feature (%s): move what both need into shared/ or integrations/", zone, ref.Spec))
				continue
			}
			if test || !pureRoles[role] {
				continue
			}
			if target == integrations && zone != integrations {
				r.problem(shown, fmt.Sprintf("a %s holds no I/O, so it imports no integration (%s): move the call into a use case", role, ref.Spec))
			} else if len(ref.Segments) > 0 && ioRoles[languages.RoleOf(ref.Segments[len(ref.Segments)-1]+".x")] {
				r.problem(shown, fmt.Sprintf("a %s holds no I/O, so it imports no client, repository or store (%s): move the call into a use case", role, ref.Spec))
			}
		}
		if test || zone == integrations || len(vendors) == 0 {
			return nil
		}
		for _, spec := range imports.All(rules.Name, text) {
			if isVendor(spec, vendors) {
				r.problem(shown, fmt.Sprintf("%q is a vendor package: only a slice in integrations/ imports it; call that integration from here", spec))
			}
		}
		return nil
	})
}

// sourceRoot is where a project's sources start, from the project root: the
// first of the language's roots that exists, then down a chain of single
// folders that hold no source of their own (a JVM package path).
func sourceRoot(dir string, rules languages.Rules) string {
	roots := rules.Roots
	if len(roots) == 0 {
		roots = []string{""}
	}
	lang, _ := imports.For(rules.Name)
	extensions := lang.Extensions
	for _, root := range roots {
		if info, err := os.Stat(filepath.Join(dir, filepath.FromSlash(root))); err != nil || !info.IsDir() {
			continue
		}
		for {
			entries, err := os.ReadDir(filepath.Join(dir, filepath.FromSlash(root)))
			if err != nil {
				return root
			}
			var folders []string
			files := false
			for _, e := range entries {
				switch {
				case e.IsDir() && !strings.HasPrefix(e.Name(), ".") && !skippedFolders[e.Name()] && !testFolders[e.Name()]:
					folders = append(folders, e.Name())
				case !e.IsDir() && slices.Contains(extensions, path.Ext(e.Name())):
					files = true
				}
			}
			if files || len(folders) != 1 || slices.Contains(languages.Zones, folders[0]) {
				return root
			}
			root = path.Join(root, folders[0])
		}
	}
	return ""
}

// isApp reports whether a project lives under an apps folder.
func isApp(dir string) bool {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return false
	}
	parts := strings.Split(filepath.ToSlash(abs), "/")
	return slices.Contains(parts[max(len(parts)-3, 0):len(parts)-1], "apps")
}

// hasZone reports whether a source root holds a zone folder.
func hasZone(base string) bool {
	for _, zone := range languages.Zones {
		if info, err := os.Stat(filepath.Join(base, zone)); err == nil && info.IsDir() {
			return true
		}
	}
	return false
}

// holdsSource reports whether a folder holds source files at any depth.
func holdsSource(dir string, rules languages.Rules, lang imports.Language, readable bool) bool {
	found := false
	_ = filepath.WalkDir(dir, func(_ string, d os.DirEntry, err error) error {
		if err != nil || found {
			return filepath.SkipAll
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") || skippedFolders[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if readable && slices.Contains(lang.Extensions, path.Ext(d.Name())) {
			found = true
		}
		return nil
	})
	return found
}

// zoneIn is the zone an import names: its first part that is a zone.
func zoneIn(segments []string) string {
	for _, s := range segments {
		if slices.Contains(languages.Zones, s) {
			return s
		}
	}
	return ""
}

// isVendor reports whether an import names a listed vendor package: the name
// itself, or something under it (after / . :: \ or :), or for a name ending
// in *, anything that starts with the rest.
func isVendor(spec string, vendors []string) bool {
	for _, v := range vendors {
		if prefix, ok := strings.CutSuffix(v, "*"); ok {
			if strings.HasPrefix(spec, prefix) {
				return true
			}
			continue
		}
		if spec == v {
			return true
		}
		for _, sep := range []string{"/", ".", "::", `\`, ":"} {
			if strings.HasPrefix(spec, v+sep) {
				return true
			}
		}
	}
	return false
}
