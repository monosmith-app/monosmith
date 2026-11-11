package electronapp

import (
	"path"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// builderShipped are what electron-builder writes into its output folder that
// a project ships: the installers and archives, the unpacked app of --dir
// (win-unpacked, linux-unpacked) and the macOS app (mac/, mac-arm64/).
var builderShipped = []string{
	"*-unpacked", "mac*/*.app",
	"*.exe", "*.msi", "*.dmg", "*.pkg", "*.zip", "*.7z", "*.AppImage", "*.deb", "*.rpm", "*.snap", "*.tar.gz",
}

// ApplyTasks gives the tasks of an Electron project their two folders (ADR
// 0092). A task that packs the app ships what the tool packed into the shipped
// folder dist/<project dir> (Task.Ship) and lists that folder as its output.
// For the app smith makes, the build writes the folder start runs,
// dist/electron/<project dir>/app, and caches it.
//
// Packing tasks are the scripts that run electron-forge package or make, or
// electron-builder; for the app smith makes, its package script. A setup whose
// folder smith cannot read ships nothing.
//
// Parameters: tasks are the project's tasks, changed in place; root the
// workspace-relative project folder; s the setup; scripts the package.json
// scripts.
func ApplyTasks(tasks map[string]protocol.Task, root string, s Setup, scripts map[string]string) {
	switch s.Tool {
	case SmithForge:
		if t, ok := tasks["build"]; ok {
			t.Outputs = []string{protocol.FromProject(root, WorkDir(root)+"/app")}
			tasks["build"] = t
		}
		if _, ok := tasks["package"]; ok {
			ship(tasks, "package", root, protocol.FromProject(root, WantedOutDir(root, Forge))+"/*-*-*")
		}
	case Forge:
		if s.OutDir == "" {
			return
		}
		for name, text := range scripts {
			switch {
			case strings.Contains(text, "electron-forge make"):
				ship(tasks, name, root, path.Join(s.OutDir, "*-*-*"), path.Join(s.OutDir, "make", "*"))
			case strings.Contains(text, "electron-forge package"):
				ship(tasks, name, root, path.Join(s.OutDir, "*-*-*"))
			}
		}
	case Builder:
		if s.OutDir == "" {
			return
		}
		globs := make([]string, len(builderShipped))
		for i, g := range builderShipped {
			globs[i] = path.Join(s.OutDir, g)
		}
		for name, text := range scripts {
			if strings.Contains(text, "electron-builder") {
				ship(tasks, name, root, globs...)
			}
		}
	}
}

// ship makes the task link globs into the shipped folder and list that folder
// as its output.
func ship(tasks map[string]protocol.Task, name, root string, globs ...string) {
	t, ok := tasks[name]
	if !ok {
		return
	}
	t.Ship = globs
	t.Outputs = []string{protocol.BuildDirFromProject(root)}
	tasks[name] = t
}
