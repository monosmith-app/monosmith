package nxworkspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// ErrNotNx is returned when the directory has no nx.json.
var ErrNotNx = problem.New("there is no nx.json here: convert runs at the root of an Nx workspace", "cd to the folder that holds nx.json, then run this again")

// ErrNxNotInstalled is returned when Nx cannot describe the workspace.
var ErrNxNotInstalled = problem.New("smith needs to read your projects before it can move them, and Nx describes them, so its packages must be installed first", "npm ci, then run this again")

// Read loads an Nx workspace: nx.json and the root package.json from disk,
// and every project with its fully resolved targets - inferred ones
// included - from Nx itself (`nx graph --file`). files are the workspace's
// files (workspace-relative, forward slashes), used to give each project its
// own files.
func Read(root string, files []string) (*Workspace, error) {
	w := &Workspace{Root: root}
	if err := readJSON(root, "nx.json", &w.Nx); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ErrNotNx
		}
		return nil, err
	}
	if err := readJSON(root, "package.json", &w.RootPackage); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	w.NxVersion = w.RootPackage.DevDependencies["nx"]
	if w.NxVersion == "" {
		w.NxVersion = w.RootPackage.Dependencies["nx"]
	}
	graph, err := describe(root)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(graph.Graph.Nodes))
	for name := range graph.Graph.Nodes {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		node := graph.Graph.Nodes[name]
		p := &Project{
			Name:        name,
			Root:        filepath.ToSlash(node.Data.Root),
			Type:        node.Type,
			ProjectType: node.Data.ProjectType,
			Tags:        node.Data.Tags,
			Targets:     node.Data.Targets,
			Release:     node.Data.Release,
		}
		if p.Targets == nil {
			p.Targets = map[string]Target{}
		}
		if p.Root == "." || p.Root == "" {
			w.RootTargets = p.Targets
			continue
		}
		for _, f := range files {
			if strings.HasPrefix(f, p.Root+"/") {
				p.Files = append(p.Files, f)
			}
		}
		var pkg PackageJSON
		if err := readJSON(root, p.Root+"/package.json", &pkg); err == nil {
			p.Package = &pkg
		}
		p.HasProjectJSON = p.HasFile("project.json")
		w.Projects = append(w.Projects, p)
	}
	return w, nil
}

type graphFile struct {
	Graph struct {
		Nodes map[string]struct {
			Type string `json:"type"`
			Data struct {
				Root        string                     `json:"root"`
				ProjectType string                     `json:"projectType"`
				Tags        []string                   `json:"tags"`
				Targets     map[string]Target          `json:"targets"`
				Release     map[string]json.RawMessage `json:"release"`
			} `json:"data"`
		} `json:"nodes"`
	} `json:"graph"`
}

// describe runs the workspace's own Nx to write its project graph.
func describe(root string) (graphFile, error) {
	var g graphFile
	entry := filepath.Join(root, "node_modules", "nx", "dist", "bin", "nx.js")
	if _, err := os.Stat(entry); err != nil {
		entry = filepath.Join(root, "node_modules", "nx", "bin", "nx.js")
		if _, err := os.Stat(entry); err != nil {
			return g, ErrNxNotInstalled
		}
	}
	dir, err := os.MkdirTemp("", "smith-convert-")
	if err != nil {
		return g, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	out := filepath.Join(dir, "graph.json")
	cmd := exec.Command("node", entry, "graph", "--file="+out)
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "NX_DAEMON=false", "NX_NO_CLOUD=true", "NX_TUI=false")
	if msg, err := cmd.CombinedOutput(); err != nil {
		return g, fmt.Errorf("nx graph failed: %w\n%s", err, strings.TrimSpace(string(msg)))
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		return g, err
	}
	if err := json.Unmarshal(raw, &g); err != nil {
		return g, fmt.Errorf("reading the Nx graph: %w", err)
	}
	return g, nil
}

func readJSON(root, rel string, out any) error {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("%s: %w", rel, err)
	}
	return nil
}
