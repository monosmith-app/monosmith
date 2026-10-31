package doctor

import (
	"encoding/json"
	"sort"
	"strings"
)

// lockfileEntry is one package in package-lock.json "packages".
type lockfileEntry struct {
	InBundle             bool              `json:"inBundle"`
	Dependencies         map[string]string `json:"dependencies"`
	OptionalDependencies map[string]string `json:"optionalDependencies"`
	BundleDependencies   json.RawMessage   `json:"bundleDependencies"`
}

// missingLockfileDeps returns the sorted names of packages that an entry of the
// lockfile depends on (dependencies or optionalDependencies) and that the
// lockfile does not hold anywhere npm would look for them. npm on one OS can
// leave out the optional packages of other platforms, and npm ci on another OS
// then fails. It returns an error when raw is not valid JSON.
func missingLockfileDeps(raw []byte) ([]string, error) {
	var lock struct {
		Packages map[string]lockfileEntry `json:"packages"`
	}
	if err := json.Unmarshal(raw, &lock); err != nil {
		return nil, err
	}
	missing := map[string]bool{}
	for path, entry := range lock.Packages {
		if path == "" || entry.InBundle {
			continue
		}
		bundled := bundledNames(entry.BundleDependencies)
		for _, deps := range []map[string]string{entry.Dependencies, entry.OptionalDependencies} {
			for name := range deps {
				if !bundled[name] && !lockfileHolds(lock.Packages, path, name) {
					missing[name] = true
				}
			}
		}
	}
	names := make([]string, 0, len(missing))
	for name := range missing {
		names = append(names, name)
	}
	sort.Strings(names)
	return names, nil
}

// bundledNames reads bundleDependencies, a list of names (or true, which names
// nothing here).
func bundledNames(raw json.RawMessage) map[string]bool {
	var list []string
	_ = json.Unmarshal(raw, &list)
	out := make(map[string]bool, len(list))
	for _, name := range list {
		out[name] = true
	}
	return out
}

// lockfileHolds resolves name from the package at path the way node does: in
// its own node_modules, then in each parent's, up to the root.
func lockfileHolds(packages map[string]lockfileEntry, path, name string) bool {
	dir := path
	for {
		key := "node_modules/" + name
		if dir != "" {
			key = dir + "/" + key
		}
		if _, ok := packages[key]; ok {
			return true
		}
		if dir == "" {
			return false
		}
		if i := strings.LastIndex(dir, "node_modules/"); i >= 0 {
			dir = strings.TrimSuffix(dir[:i], "/")
		} else {
			dir = ""
		}
	}
}
