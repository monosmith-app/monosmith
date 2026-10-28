package agentfiles

import (
	"errors"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Finding is smith doctor's check of the agent files, for the clients the
// workspace has. A missing file, or a stale one smith wrote, is a warning with
// a fix that writes it; a file the person edited is theirs and passes.
func Finding(root string) protocol.Finding {
	check := "The MonoSmith agent files are installed and current"
	claude, vscode := PresentClients(root, nil)
	files, err := Plan(Options{Root: root, Claude: claude, VSCode: vscode, Agent: true})
	if err != nil {
		detail := err.Error()
		remedy := "fix the JSON by hand, or write the smith entry yourself (see docs/manual/agent.md)"
		if errors.Is(err, ErrNotJSON) {
			detail = "an MCP file is not plain JSON (comments?): " + err.Error()
		}
		return protocol.Finding{Check: check, Warning: true, Detail: detail, Remedy: remedy}
	}
	var todo []File
	var names []string
	for _, f := range files {
		if f.Status.Writable() {
			todo = append(todo, f)
			names = append(names, f.Path+f.problem())
		}
	}
	if len(todo) == 0 {
		return protocol.Finding{Check: check, OK: true, Warning: true}
	}
	fix := &protocol.Fix{Description: "write the MonoSmith agent files (smith agent install)", Safe: true}
	for _, f := range todo {
		fix.Files = append(fix.Files, protocol.FileEdit{Path: f.Path, Content: f.New})
	}
	return protocol.Finding{
		Check: check, Warning: true, Fix: fix,
		Detail: strings.Join(names, "; "),
		Remedy: "run `smith agent install`, or set `enabled = false` under [agent] in smith.toml if you do not want them",
	}
}

// Refresh brings the agent file up to date when smith wrote it and nobody
// edited it (smith upgrade). It creates nothing: a workspace is offered the
// files with smith agent install. With check it only lists what would change.
// kept is the agent file when the person edited it, which smith left alone.
func Refresh(root string, check bool) (changed []string, kept bool, err error) {
	f, err := planAgent(root, CatalogTools())
	if err != nil {
		return nil, false, err
	}
	switch f.Status {
	case Stale:
		changed = append(changed, f.Path)
		if !check {
			_, err = Write(root, []File{f})
		}
	case Edited:
		kept = true
	}
	return changed, kept, err
}
