package layout

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-vertical-slices/languages"
)

// Check walks a project and reports what breaks the vertical-slice rules of
// its language, the zones of an app included.
//
// Parameters: dir is the project root (where its manifest is); language is
// its language; vendors are the packages only integrations/ may import.
// Returns the report, or an error for an unknown language or an unreadable
// folder.
func Check(dir, language string, vendors []string) (Report, error) {
	var r Report
	rules, ok := languages.For(language)
	if !ok {
		return r, fmt.Errorf("the vertical-slices check knows %s, not %q", strings.Join(languages.Names(), ", "), language)
	}
	if err := walk(&r, dir, rules, "", nil, sourceRoot(dir, rules)); err != nil {
		return r, err
	}
	return r, checkZones(&r, dir, rules, vendors)
}

func walk(r *Report, root string, rules languages.Rules, rel string, parent *languages.Dir, zoneRoot string) error {
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	here := languages.Dir{Rel: rel, Files: map[string]bool{}, Folders: map[string]bool{}, Parent: parent}
	for _, e := range entries {
		if e.IsDir() {
			here.Folders[e.Name()] = true
		} else {
			here.Files[e.Name()] = true
		}
	}
	shown := func(name string) string { return path.Join(rel, name) }
	files := 0
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			if strings.HasPrefix(name, ".") || skippedFolders[name] || (rules.Skip != nil && rules.Skip(name, here)) {
				continue
			}
			// shared is a zone in the source root, or beside features/ or
			// integrations/ (a test tree that mirrors the zones); a junk drawer
			// anywhere else.
			zone := name == "shared" && (rel == zoneRoot || here.Folders["features"] || here.Folders["integrations"])
			if languages.ForbiddenFolders[strings.ToLower(name)] && !zone {
				r.problem(shown(name)+"/", fmt.Sprintf("%q is a junk drawer; name the folder for the outcome it serves", name))
				continue
			}
			if rules.Folder != nil {
				if why := rules.Folder(name, here); why != "" {
					r.problem(shown(name)+"/", why)
				}
			}
			if err := walk(r, root, rules, shown(name), &here, zoneRoot); err != nil {
				return err
			}
			continue
		}
		if rules.Source != nil && !rules.Source(name, here) {
			continue
		}
		if why := rules.File(name, here); why != "" {
			r.problem(shown(name), why)
		}
		if word := languages.BannedWordIn(name); word != "" {
			r.problem(shown(name), fmt.Sprintf("%q is in the name of this file and names no outcome; name the file for what it does", word))
		}
		if rules.Production == nil || rules.Production(name, here) {
			files++
		}
	}
	if files > SliceSize {
		r.warn(strings.TrimPrefix(rel+"/", "/"), fmt.Sprintf("%d production files (more than %d): it usually holds two outcomes, so split it into two slices or write the exception down", files, SliceSize))
	}
	return nil
}
