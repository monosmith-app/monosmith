package workspacejoin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// workspaceWords names the npm workspace in a finding.
const workspaceWords = `the root package.json "workspaces"`

// rootManifest is the workspace's root package.json as Join reads it.
type rootManifest struct {
	raw []byte
	// globs are the "workspaces" entries: the array, or the object's packages.
	globs []string
	// object is true when "workspaces" is the object form ({"packages": [...]}),
	// which Join does not rewrite.
	object bool
}

// readRoot reads the root package.json. A missing file is an empty manifest.
func readRoot(workspaceRoot string) (rootManifest, error) {
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, "package.json"))
	if os.IsNotExist(err) {
		return rootManifest{}, nil
	}
	if err != nil {
		return rootManifest{}, err
	}
	m := rootManifest{raw: raw}
	var pkg struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return m, fmt.Errorf("the root package.json is not valid JSON: %w", err)
	}
	if len(pkg.Workspaces) == 0 || string(pkg.Workspaces) == "null" {
		return m, nil
	}
	if json.Unmarshal(pkg.Workspaces, &m.globs) == nil {
		return m, nil
	}
	var object struct {
		Packages []string `json:"packages"`
	}
	if err := json.Unmarshal(pkg.Workspaces, &object); err != nil {
		return m, fmt.Errorf(`the root package.json "workspaces" is neither a list nor {"packages": [...]}`)
	}
	m.globs, m.object = object.Packages, true
	return m, nil
}

// Unjoined names the projects the root package.json "workspaces" do not
// cover, so npm install does not reach them. The project at the workspace
// root is the workspace itself and always reached.
//
// Parameters: workspaceRoot is the workspace folder; projects are the Node
// projects.
// Returns: the names of the projects npm install does not reach, in order.
func Unjoined(workspaceRoot string, projects []protocol.Project) []string {
	m, err := readRoot(workspaceRoot)
	var out []string
	for _, p := range projects {
		if p.Root != "." && (err != nil || !Covers(m.globs, p.Root)) {
			out = append(out, p.Name)
		}
	}
	return out
}

// Join says, for each Node project, whether the root package.json
// "workspaces" cover it. The projects they do not cover share one fix: the
// root package.json with the parent folder's glob added for each of them
// (a new private manifest when there is none), then npm install. Everything
// else in the manifest stays as written. It writes nothing.
//
// Parameters: workspaceRoot is the workspace folder; params are the projects.
// Returns: one state for each project not at the workspace root, or an error
// when the root package.json cannot be read.
func Join(workspaceRoot string, params protocol.JoinParams) (protocol.JoinResult, error) {
	m, err := readRoot(workspaceRoot)
	if err != nil {
		return protocol.JoinResult{}, err
	}
	res := protocol.JoinResult{Projects: []protocol.JoinState{}}
	var missing, names []string
	for _, p := range params.Projects {
		if p.Root == "." {
			continue
		}
		joined := Covers(m.globs, p.Root)
		state := protocol.JoinState{Project: p.Name, Joined: joined, Workspace: workspaceWords}
		if !joined {
			state.Problem = `is not in the root package.json "workspaces": npm install does not reach it, so its dependencies are not installed`
			names = append(names, p.Name)
			if pattern := Pattern(p.Root); !slices.Contains(missing, pattern) && !Covers(missing, p.Root) {
				missing = append(missing, pattern)
			}
		}
		res.Projects = append(res.Projects, state)
	}
	if len(missing) == 0 || m.object {
		return res, nil
	}
	content, err := withGlobs(m, missing, filepath.Base(workspaceRoot))
	if err != nil {
		return protocol.JoinResult{}, err
	}
	fix := &protocol.Fix{
		Description: `add ` + strings.Join(quoteAll(missing), ", ") + ` to the root package.json "workspaces", then run npm install`,
		Safe:        true,
		Files:       []protocol.FileEdit{{Path: "package.json", Content: content}},
		Steps:       []protocol.FixStep{{Description: "install the workspace's dependencies", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}}},
		AppliesTo:   names,
	}
	for i := range res.Projects {
		if !res.Projects[i].Joined {
			res.Projects[i].Fix = fix
		}
	}
	return res, nil
}

// withGlobs is the root package.json with the globs added to "workspaces",
// or a new private manifest named after the workspace folder.
func withGlobs(m rootManifest, add []string, folder string) (string, error) {
	globs := append(slices.Clone(m.globs), add...)
	if len(strings.TrimSpace(string(m.raw))) == 0 {
		out, err := json.MarshalIndent(struct {
			Name       string   `json:"name"`
			Private    bool     `json:"private"`
			Workspaces []string `json:"workspaces"`
		}{strings.ToLower(folder) + "-workspace", true, globs}, "", "  ")
		return string(out) + "\n", err
	}
	out, err := jsondoc.SetKeysInPlace(m.raw, map[string]any{"workspaces": globs})
	return string(out), err
}

func quoteAll(items []string) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = `"` + s + `"`
	}
	return out
}
