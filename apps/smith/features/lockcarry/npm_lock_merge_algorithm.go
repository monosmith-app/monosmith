package lockcarry

import (
	"bytes"
	"encoding/json"
	"errors"
	"sort"
	"strings"
)

// Locked is one direct dependency of the project and the version its
// lockfile locked.
type Locked struct {
	// Name is the package name.
	Name string
	// Version is the locked version.
	Version string
}

// npmLock is the top of a package-lock.json, in npm's order of keys.
type npmLock struct {
	Name            string                     `json:"name,omitempty"`
	Version         string                     `json:"version,omitempty"`
	LockfileVersion int                        `json:"lockfileVersion"`
	Requires        bool                       `json:"requires,omitempty"`
	Packages        map[string]json.RawMessage `json:"packages,omitempty"`
	Dependencies    json.RawMessage            `json:"dependencies,omitempty"`
}

// v1Dep is one entry of the nested "dependencies" tree of lockfileVersion 1.
type v1Dep struct {
	Version      string            `json:"version"`
	Resolved     string            `json:"resolved,omitempty"`
	Integrity    string            `json:"integrity,omitempty"`
	Dev          bool              `json:"dev,omitempty"`
	Optional     bool              `json:"optional,omitempty"`
	Requires     map[string]string `json:"requires,omitempty"`
	Dependencies map[string]v1Dep  `json:"dependencies,omitempty"`
}

// packageEntry is the part of a "packages" entry that the merge reads.
type packageEntry struct {
	Version         string            `json:"version"`
	Link            bool              `json:"link"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// ErrNoPackages is returned when the workspace lockfile has no "packages"
// section (lockfileVersion 1), which npm 7 and later always write.
var ErrNoPackages = errors.New(`the workspace package-lock.json has no "packages" section`)

// MergeNpm adds the installed packages of a project's own package-lock.json
// to the workspace's package-lock.json, under the project's folder
// (<folder>/node_modules/<name>). npm keeps an entry of the lockfile that
// satisfies the range, so the next install puts the same versions in place.
// An entry the workspace has already, a workspace link and a local folder
// are not copied.
//
// Parameters: workspace is the workspace's package-lock.json; source is the
// project's (lockfileVersion 1, 2 or 3); manifest is the project's
// package.json, which names its direct dependencies; folder is the project's
// folder from the workspace root, with forward slashes.
// Returns: the new workspace lockfile; the project's direct dependencies and
// their locked versions, sorted by name; the number of entries added; and an
// error when a file is not a lockfile npm reads.
func MergeNpm(workspace, source, manifest []byte, folder string) ([]byte, []Locked, int, error) {
	var ws, src npmLock
	if err := json.Unmarshal(workspace, &ws); err != nil {
		return nil, nil, 0, err
	}
	if ws.Packages == nil {
		return nil, nil, 0, ErrNoPackages
	}
	if err := json.Unmarshal(source, &src); err != nil {
		return nil, nil, 0, err
	}
	entries, err := sourcePackages(src)
	if err != nil {
		return nil, nil, 0, err
	}
	added := 0
	folder = strings.TrimSuffix(folder, "/")
	for key, raw := range entries {
		if !strings.HasPrefix(key, "node_modules/") {
			continue
		}
		var e packageEntry
		if json.Unmarshal(raw, &e) != nil || e.Link || isLocal(e.Version) {
			continue
		}
		target := folder + "/" + key
		if _, ok := ws.Packages[target]; ok {
			continue
		}
		ws.Packages[target] = raw
		added++
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(ws); err != nil {
		return nil, nil, 0, err
	}
	return out.Bytes(), direct(manifest, entries), added, nil
}

// sourcePackages is the "packages" section of a lockfile, converted from the
// nested "dependencies" tree when the lockfile is version 1.
func sourcePackages(src npmLock) (map[string]json.RawMessage, error) {
	if src.Packages != nil {
		return src.Packages, nil
	}
	var tree map[string]v1Dep
	if len(src.Dependencies) > 0 {
		if err := json.Unmarshal(src.Dependencies, &tree); err != nil {
			return nil, err
		}
	}
	out := map[string]json.RawMessage{}
	var walk func(prefix string, deps map[string]v1Dep)
	walk = func(prefix string, deps map[string]v1Dep) {
		for name, d := range deps {
			key := prefix + "node_modules/" + name
			entry := map[string]any{"version": d.Version}
			if d.Resolved != "" {
				entry["resolved"] = d.Resolved
			}
			if d.Integrity != "" {
				entry["integrity"] = d.Integrity
			}
			if d.Dev {
				entry["dev"] = true
			}
			if d.Optional {
				entry["optional"] = true
			}
			if len(d.Requires) > 0 {
				entry["dependencies"] = d.Requires
			}
			raw, _ := json.Marshal(entry)
			out[key] = raw
			walk(key+"/", d.Dependencies)
		}
	}
	walk("", tree)
	return out, nil
}

// direct lists the project's direct dependencies and their locked versions.
func direct(manifest []byte, entries map[string]json.RawMessage) []Locked {
	names := map[string]bool{}
	var root packageEntry
	if json.Unmarshal(manifest, &root) == nil {
		for n := range root.Dependencies {
			names[n] = true
		}
		for n := range root.DevDependencies {
			names[n] = true
		}
	}
	var out []Locked
	for n := range names {
		var e packageEntry
		if raw, ok := entries["node_modules/"+n]; ok && json.Unmarshal(raw, &e) == nil && !e.Link && !isLocal(e.Version) && e.Version != "" {
			out = append(out, Locked{Name: n, Version: e.Version})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// isLocal is true for a version that points at a folder, not a registry.
func isLocal(version string) bool {
	return strings.HasPrefix(version, "file:") || strings.HasPrefix(version, "link:")
}
