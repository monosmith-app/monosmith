package agentfiles

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Status says how one of the files stands against what smith would write.
type Status int

// The statuses.
const (
	// Missing: the file, or smith's entry in it, is absent. Writing adds it.
	Missing Status = iota
	// Current: it is what smith writes now.
	Current
	// Stale: smith wrote it and nobody edited it, but smith's text changed.
	// Writing brings it up to date.
	Stale
	// Edited: it is the person's. smith never overwrites it.
	Edited
)

// problem says what is wrong with a file that writing would fix.
func (f File) problem() string {
	switch {
	case f.Status == Stale:
		return " is out of date"
	case f.Old != "":
		return " has no smith server"
	}
	return " is missing"
}

// action says what writing the file does, for the preview.
func (f File) action() string {
	switch {
	case f.Status == Stale:
		return "Update " + f.Path
	case f.Old != "":
		return "Add the smith server to " + f.Path
	}
	return "Add " + f.Path
}

// Writable reports whether writing the file changes it.
func (s Status) Writable() bool { return s == Missing || s == Stale }

// File is one of the files and what smith would make of it.
type File struct {
	Path   string
	Status Status
	Old    string
	// New is the whole file smith would make of it. For an Edited file that is
	// the file with smith's text put back, which smith shows and never writes.
	New string
}

// Options choose the files.
type Options struct {
	Root string
	// Claude selects .mcp.json and, with Agent, .claude/agents/smith.md.
	Claude bool
	// VSCode selects .vscode/mcp.json.
	VSCode bool
	// Agent includes the agent file (smith agent install, not smith mcp install).
	Agent bool
	// Launcher is how a new entry starts smith; the zero value is DirectLauncher.
	Launcher Launcher
}

// Plan reads the files and says what writing them would do. An MCP file that
// is not plain JSON is an error that wraps ErrNotJSON.
func Plan(o Options) ([]File, error) {
	if o.Launcher.Command == "" {
		o.Launcher = DirectLauncher
	}
	var files []File
	if o.Claude && o.Agent {
		f, err := planAgent(o.Root, CatalogTools())
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	for _, c := range []struct {
		on   bool
		form MCPForm
	}{{o.Claude, ClaudeForm}, {o.VSCode, VSCodeForm}} {
		if !c.on {
			continue
		}
		f, err := planMCP(o.Root, c.form, o.Launcher)
		if err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, nil
}

func readOrEmpty(root, path string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	return string(raw), err
}

func planAgent(root string, tools []Tool) (File, error) {
	old, err := readOrEmpty(root, AgentPath)
	if err != nil {
		return File{}, err
	}
	want := RenderAgent(tools)
	f := File{Path: AgentPath, Old: old, New: want}
	if old == "" {
		return f, nil
	}
	plain, recorded, ok := unmarked(old)
	switch {
	case !ok || digest(plain) != recorded:
		f.Status = Edited
	case strings.ReplaceAll(old, "\r\n", "\n") == want:
		f.Status = Current
	default:
		f.Status = Stale
	}
	return f, nil
}

func planMCP(root string, form MCPForm, l Launcher) (File, error) {
	old, err := readOrEmpty(root, form.Path)
	if err != nil {
		return File{}, err
	}
	merged, state, err := MergeMCP([]byte(old), form, l)
	if err != nil {
		return File{}, err
	}
	f := File{Path: form.Path, Old: old, New: string(merged)}
	switch state {
	case MCPCurrent:
		f.Status = Current
	case MCPEdited:
		f.Status = Edited
	}
	return f, nil
}

// Write writes the files that are Missing or Stale and returns their paths.
func Write(root string, files []File) ([]string, error) {
	var written []string
	for _, f := range files {
		if !f.Status.Writable() {
			continue
		}
		p := filepath.Join(root, filepath.FromSlash(f.Path))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return written, err
		}
		if err := os.WriteFile(p, []byte(f.New), 0o644); err != nil {
			return written, err
		}
		written = append(written, f.Path)
	}
	return written, nil
}
