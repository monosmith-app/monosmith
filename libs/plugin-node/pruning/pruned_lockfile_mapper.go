package pruning

import (
	"encoding/json"
	"strings"
)

// ModulesDir is where the deploy output keeps its copies of the workspace
// packages it uses.
const ModulesDir = "workspace_modules"

type lockfile struct {
	Packages map[string]json.RawMessage `json:"packages"`
}

type lockEntry struct {
	Link     bool   `json:"link"`
	Resolved string `json:"resolved"`
}

// WorkspacePackages lists the workspace's npm packages, by name, with their
// directories, from the links npm records for them in the root lockfile
// ("node_modules/@acme/core": {"resolved": "packages/core", "link": true}).
func WorkspacePackages(lock []byte) (map[string]string, error) {
	var l lockfile
	if err := json.Unmarshal(lock, &l); err != nil {
		return nil, err
	}
	out := map[string]string{}
	for key, raw := range l.Packages {
		name, ok := strings.CutPrefix(key, "node_modules/")
		if !ok || strings.Contains(name, "/node_modules/") {
			continue
		}
		var e lockEntry
		if json.Unmarshal(raw, &e) != nil || !e.Link || strings.HasPrefix(e.Resolved, "node_modules/") {
			continue
		}
		out[name] = e.Resolved
	}
	return out, nil
}

// SeedLockfile is the deploy output's starting lockfile: every package the
// workspace installed, at the version it installed, with the app as the root
// and each copied workspace package under workspace_modules/. Dependencies
// nested under the app's directory become top-level (the app is the root
// now), and those nested under a copied package move with it. npm then
// prunes it to the app's production tree (npm install --package-lock-only),
// so the versions are the workspace's, not newly resolved ones. modules maps
// each copied package's name to its workspace directory.
func SeedLockfile(lock []byte, name, version, appRoot string, modules map[string]string) ([]byte, error) {
	var l lockfile
	if err := json.Unmarshal(lock, &l); err != nil {
		return nil, err
	}
	packages := map[string]json.RawMessage{}
	root, err := json.Marshal(map[string]string{"name": name, "version": version})
	if err != nil {
		return nil, err
	}
	packages[""] = root
	var relocated [][2]string
	for key, raw := range l.Packages {
		if key == "" {
			continue
		}
		if rest, ok := strings.CutPrefix(key, appRoot+"/node_modules/"); ok {
			relocated = append(relocated, [2]string{"node_modules/" + rest, key})
			continue
		}
		moved := false
		for pkg, dir := range modules {
			if rest, ok := strings.CutPrefix(key, dir+"/node_modules/"); ok {
				relocated = append(relocated, [2]string{ModulesDir + "/" + pkg + "/node_modules/" + rest, key})
				moved = true
				break
			}
		}
		// Workspace directories and their own node_modules are not installed
		// by the deploy; only the workspace's node_modules/ is.
		if moved || !strings.HasPrefix(key, "node_modules/") {
			continue
		}
		var e lockEntry
		if json.Unmarshal(raw, &e) == nil && e.Link && !strings.HasPrefix(e.Resolved, "node_modules/") {
			continue // a workspace package: npm links its copy instead
		}
		packages[key] = raw
	}
	// Last, so the app's own versions win over the workspace's hoisted ones.
	for _, r := range relocated {
		packages[r[0]] = l.Packages[r[1]]
	}
	return json.MarshalIndent(map[string]any{
		"name":            name,
		"version":         version,
		"lockfileVersion": 3,
		"requires":        true,
		"packages":        packages,
	}, "", "  ")
}
