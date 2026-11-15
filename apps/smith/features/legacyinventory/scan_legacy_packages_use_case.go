package legacyinventory

import (
	"sort"
	"strings"
)

// ScanPackages finds the packages of the retired families that nothing
// remaining imports or runs. The tree should already hold the file changes
// the plan makes, so a package only a deleted config used is not kept by it.
// A package stays (reported in Kept) while a file that stays names it or runs
// its commands, a package that stays needs it as a peer, or a project
// declares it as a peer dependency of its own.
func ScanPackages(t *Tree, retired []Family) Inventory {
	is := map[Family]bool{}
	for _, f := range retired {
		is[f] = true
	}
	ix := BuildIndex(t)
	declaredIn := map[string][]string{} // package -> manifests declaring it as a dependency
	peerIn := map[string]string{}       // package -> a manifest declaring it as a peer
	all := map[string]bool{}            // every declared package
	for _, mf := range t.ManifestPaths() {
		m, ok := ReadManifest(t.Read(mf))
		if !ok {
			continue
		}
		for _, deps := range []map[string]string{m.Dependencies, m.DevDependencies, m.OptionalDependencies} {
			for name := range deps {
				declaredIn[name] = appendOnce(declaredIn[name], mf)
				all[name] = true
			}
		}
		for name := range m.PeerDependencies {
			peerIn[name] = mf
			all[name] = true
		}
	}
	family := map[string]Family{}
	var candidates []string
	for name := range all {
		if f, ok := packageFamily(name); ok && is[f] {
			family[name] = f
			candidates = append(candidates, name)
		}
	}
	sort.Strings(candidates)

	kept := map[string]string{}
	for _, name := range candidates {
		if why := packageStaysBecause(t, ix, name, family[name], peerIn); why != "" {
			kept[name] = why
		}
	}
	// A package that stays may need another as a peer, and that one then
	// stays too; repeat until nothing more is kept.
	for round := 0; round < maxScanRounds; round++ {
		changed := false
		for name := range all {
			if _, goes := family[name]; goes {
				if _, stays := kept[name]; !stays {
					continue
				}
			}
			m, ok := t.installed(name)
			if !ok {
				continue
			}
			for _, peer := range m.RequiredPeers() {
				if _, candidate := family[peer]; !candidate {
					continue
				}
				if _, stays := kept[peer]; !stays {
					kept[peer] = "a required peer of " + name
					changed = true
				}
			}
		}
		if !changed {
			break
		}
	}
	var inv Inventory
	for _, name := range candidates {
		if why, stays := kept[name]; stays {
			inv.Kept = append(inv.Kept, Kept{What: name, Family: family[name], Why: why})
			continue
		}
		if len(declaredIn[name]) == 0 {
			continue
		}
		inv.Packages = append(inv.Packages, LegacyPackage{Name: name, Family: family[name], In: declaredIn[name]})
	}
	return inv
}

// packageStaysBecause says what keeps a package of a retired family, "" when
// nothing does.
func packageStaysBecause(t *Tree, ix *Index, name string, family Family, peerIn map[string]string) string {
	if mf, ok := peerIn[name]; ok {
		return "declared as a peer dependency in " + mf
	}
	if tokens := familyTokens[family]; len(tokens) > 0 {
		if refs := ix.Files(2, tokens...); len(refs) > 0 {
			return string(family) + " is still used by " + strings.Join(refs, ", ")
		}
	}
	words := []string{name}
	if m, ok := t.installed(name); ok {
		words = append(words, m.Bins()...)
	}
	if rest, ok := strings.CutPrefix(name, "@types/"); ok {
		words = append(words, strings.Replace(rest, "__", "/", 1))
		if strings.Contains(rest, "__") {
			words = append(words, "@"+strings.Replace(rest, "__", "/", 1))
		}
	}
	if refs := ix.Files(2, words...); len(refs) > 0 {
		return "used by " + strings.Join(refs, ", ")
	}
	return ""
}

func appendOnce(list []string, item string) []string {
	for _, have := range list {
		if have == item {
			return list
		}
	}
	return append(list, item)
}
