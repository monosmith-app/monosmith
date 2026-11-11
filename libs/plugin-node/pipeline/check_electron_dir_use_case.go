package pipeline

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-node/electronapp"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckElectronDir checks where an adopted Electron Forge or electron-builder
// setup packs the app: in its work folder dist/electron/<project dir> (ADR
// 0092), not in the project, where the packed app and the tool's work files
// sit among the sources. The packing tasks link the packed app into the
// shipped folder dist/<project dir> wherever the tool packs, so this is a
// warning. When smith can edit the configuration safely, the finding carries
// the move as a presented choice; else the remedy says what to set by hand.
// The folder Forge's Vite plugin writes (.vite/) is fixed by the plugin and
// is not checked.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project.
// Returns the finding, and false when the project is not an adopted Electron
// setup.
func CheckElectronDir(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	s := electronapp.Detect(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root)))
	if s.Tool != electronapp.Forge && s.Tool != electronapp.Builder {
		return protocol.Finding{}, false
	}
	want := electronapp.WantedOutDir(p.Root, s.Tool)
	setting, tool := "outDir", "Electron Forge"
	if s.Tool == electronapp.Builder {
		setting, tool = "directories.output", "electron-builder"
	}
	where := "the Forge configuration"
	if s.Config != "" {
		where = p.Root + "/" + s.Config
	}
	resolved := ""
	if s.OutDir != "" {
		resolved = path.Clean(path.Join(p.Root, s.OutDir))
	}
	ok := resolved == want || strings.HasPrefix(resolved, want+"/")
	f := protocol.Finding{
		Check:   p.Name + " packs into " + want,
		Project: p.Name,
		OK:      ok,
		Warning: true,
		Remedy:  "set " + setting + " to \"" + protocol.FromProject(p.Root, want) + "\" in " + where,
	}
	if ok {
		return f, true
	}
	if resolved == "" {
		f.Problem = "smith cannot read where " + tool + " packs"
		f.Detail = where + " sets " + setting + " in a way smith does not read, so smith cannot tell where the packed app goes or link it into " + protocol.ProjectBuildDir(p.Root)
		return f, true
	}
	f.Problem = "packs into " + resolved + ", not " + want
	f.Detail = tool + " packs " + p.Name + " into " + resolved + ". smith keeps a tool's work files in " + electronapp.WorkDir(p.Root) + " and links the packed app into " + protocol.ProjectBuildDir(p.Root)
	f.Fix = electronapp.MoveOutDirFix(workspaceRoot, p, s)
	if f.Fix == nil {
		f.Remedy = "smith does not edit " + where + " for you, because it cannot change that file safely. To move it by hand: " + f.Remedy
	}
	return f, true
}
