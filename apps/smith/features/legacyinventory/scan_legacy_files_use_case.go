package legacyinventory

import (
	"path"
	"regexp"
	"strings"
)

var swcrcDisabled = regexp.MustCompile(`swcrc["']?\s*[:=]\s*false`)

// maxScanRounds bounds the loop that puts kept files back in the index.
const maxScanRounds = 8

// ScanFiles finds the files of the retired families that nothing remaining
// needs. A file stays (and is reported in Kept) while a file that stays
// imports it, runs its tool, or loads it by convention.
func ScanFiles(t *Tree, retired []Family) Inventory {
	is := map[Family]bool{}
	for _, f := range retired {
		is[f] = true
	}
	rules := map[string]fileRule{}
	var candidates []string
	for _, f := range t.Files() {
		if r, ok := fileFamily(f); ok && is[r.family] {
			rules[f] = r
			candidates = append(candidates, f)
		}
	}
	kept := map[string]string{}
	for round := 0; round < maxScanRounds; round++ {
		work := t.Clone()
		for _, f := range candidates {
			if _, stays := kept[f]; !stays {
				work.Delete(f)
			}
		}
		ix := BuildIndex(work)
		changed := false
		for _, f := range candidates {
			if _, stays := kept[f]; stays {
				continue
			}
			if why := fileStaysBecause(work, ix, f, rules[f]); why != "" {
				kept[f] = why
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	var inv Inventory
	for _, f := range candidates {
		if why, stays := kept[f]; stays {
			inv.Kept = append(inv.Kept, Kept{What: f, Family: rules[f].family, Why: why})
		} else {
			inv.Files = append(inv.Files, LegacyFile{Path: f, Family: rules[f].family})
		}
	}
	return inv
}

// fileStaysBecause says what keeps a file of a retired family, "" when
// nothing does.
func fileStaysBecause(work *Tree, ix *Index, f string, r fileRule) string {
	if tokens := familyTokens[r.family]; len(tokens) > 0 {
		if refs := ix.Files(2, tokens...); len(refs) > 0 {
			return string(r.family) + " is still used by " + strings.Join(refs, ", ")
		}
	}
	if r.byName {
		if refs := ix.Files(2, path.Base(f)); len(refs) > 0 {
			return "mentioned by " + strings.Join(refs, ", ")
		}
	}
	switch r.tool {
	case "":
	case "swc":
		if ix.Runs("swc") {
			return "a script runs swc"
		}
		if consumer := swcrcConsumer(work); consumer != "" {
			return "loaded by @swc/jest in " + consumer
		}
	case "babel":
		if refs := ix.Files(2, "babel", "babel-jest", "babel-loader", "@babel"); len(refs) > 0 {
			return "babel is still used by " + strings.Join(refs, ", ")
		}
		if consumer := jestWithoutTransform(work, ix); consumer != "" {
			return "jest loads it (babel-jest is its default transform) in " + consumer
		}
	default:
		if refs := ix.Files(2, r.tool); len(refs) > 0 {
			return r.tool + " is still used by " + strings.Join(refs, ", ")
		}
	}
	return ""
}

// swcrcConsumer is a Jest config that uses @swc/jest and does not turn off
// its search for a .swcrc.
func swcrcConsumer(t *Tree) string {
	for _, f := range t.Files() {
		if !strings.HasPrefix(path.Base(f), "jest.config.") {
			continue
		}
		text := t.Read(f)
		if strings.Contains(text, "@swc/jest") && !swcrcDisabled.MatchString(text) {
			return f
		}
	}
	return ""
}

// jestWithoutTransform is a Jest config that sets no transform, so Jest uses
// babel-jest and the Babel config next to it.
func jestWithoutTransform(t *Tree, ix *Index) string {
	if !ix.Mentions("jest") {
		return ""
	}
	for _, f := range t.Files() {
		if strings.HasPrefix(path.Base(f), "jest.config.") && !strings.Contains(t.Read(f), "transform") && !strings.Contains(t.Read(f), "preset") {
			return f
		}
	}
	return ""
}
