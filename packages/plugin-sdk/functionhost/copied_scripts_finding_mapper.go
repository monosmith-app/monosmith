package functionhost

import (
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CopiedScriptsFix is the description of the fix, the same on every project so
// that the host folds them into one.
const CopiedScriptsFix = "run smith's commands in place of the copied scripts"

// CopiedScriptsFinding checks one function or Electron app for the scripts
// that older smith copied into it (CopiedScripts). Its fix removes the copies
// and points the package.json scripts at smith's commands. A copy that is not
// as smith wrote it was edited by hand: the fix then asks, and keeps the copy
// unless the person says to remove it.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project;
// host is a host choice name or Electron.
// Returns the finding, and false when the project holds no copy.
func CopiedScriptsFinding(workspaceRoot string, p protocol.Project, host string) (protocol.Finding, bool) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	manifest, _ := os.ReadFile(filepath.Join(dir, "package.json"))
	detail := "a fix to a copied script never reaches the app. smith runs the same steps itself now (" + AssembleCommand(host) + ", " + packCommand(host) + ")"
	remedy := "smith doctor --fix removes the copies and points the scripts of package.json at smith"
	return copiesFinding(workspaceRoot, p, CopiedScripts, detail, remedy, func(remove []string, copies map[string]string) []protocol.FileEdit {
		return removeCopies(p, host, manifest, copies, remove)
	})
}

// CopiedPythonScriptsFinding checks one Python function app for package.py and
// start.py, which older smith copied into it. The tasks of the app run `smith
// func stage`, `smith func zip` and the tool of the host now, so the fix only
// removes the copies. A copy that was edited is asked about, as in
// CopiedScriptsFinding.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project.
// Returns the finding, and false when the project holds no copy.
func CopiedPythonScriptsFinding(workspaceRoot string, p protocol.Project) (protocol.Finding, bool) {
	detail := "a fix to a copied script never reaches the app. smith runs the same steps itself now (smith func stage, smith func zip) and the tool of the host starts the function"
	remedy := "smith doctor --fix removes the copies"
	return copiesFinding(workspaceRoot, p, CopiedPythonScripts, detail, remedy, func(remove []string, _ map[string]string) []protocol.FileEdit {
		return deleteEdits(p, CopiedPythonScripts, remove)
	})
}

// copiesFinding is the finding for the copied files of a project, with the
// fix that removes them. edits makes the file edits that remove the listed
// copies; copies holds the content of every copy found.
func copiesFinding(workspaceRoot string, p protocol.Project, files []string, detail, remedy string, edits func(remove []string, copies map[string]string) []protocol.FileEdit) (protocol.Finding, bool) {
	dir := filepath.Join(workspaceRoot, filepath.FromSlash(p.Root))
	copies := map[string]string{}
	for _, rel := range files {
		if raw, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel))); err == nil {
			copies[rel] = string(raw)
		}
	}
	if len(copies) == 0 {
		return protocol.Finding{}, false
	}
	var names, edited, scaffolded []string
	for _, rel := range files {
		text, ok := copies[rel]
		if !ok {
			continue
		}
		names = append(names, rel)
		if IsScaffolded(text) {
			scaffolded = append(scaffolded, rel)
		} else {
			edited = append(edited, rel)
		}
	}
	f := protocol.Finding{
		Check:   p.Name + " runs smith's commands, not copied scripts",
		Project: p.Name,
		Problem: strings.Join(names, ", ") + " in " + p.Root + " are copies of smith scripts",
		Detail:  detail,
		Remedy:  remedy,
		Warning: true,
	}
	fix := protocol.Fix{Description: CopiedScriptsFix, Safe: len(edited) == 0, AppliesTo: []string{p.Name}}
	if len(edited) == 0 {
		fix.Files = edits(scaffolded, copies)
		f.Fix = &fix
		return f, true
	}
	f.Detail += ". " + strings.Join(edited, ", ") + " was edited by hand"
	keep := protocol.Fix{Description: "keep the edited copies", Files: edits(scaffolded, copies)}
	drop := protocol.Fix{Description: "remove every copy, the edits too", Files: edits(names, copies)}
	fix.Choices = []protocol.FixChoice{{
		Question: p.Name + ": " + strings.Join(edited, ", ") + " was changed by hand. Keep it, or remove it and let smith run the command?",
		Options: []protocol.FixChoiceOption{
			{Label: "Keep the copies I edited", Fix: keep},
			{Label: "Remove them all: the edits are lost", Fix: drop},
		},
	}}
	f.Fix = &fix
	return f, true
}

// deleteEdits are the file edits that delete the listed copies of a project.
func deleteEdits(p protocol.Project, files, remove []string) []protocol.FileEdit {
	var edits []protocol.FileEdit
	for _, rel := range files {
		if slices.Contains(remove, rel) {
			edits = append(edits, protocol.FileEdit{Path: path.Join(p.Root, rel), Delete: true})
		}
	}
	return edits
}

// removeCopies is the file edits that remove the listed copies and point the
// scripts of package.json that ran them at smith.
func removeCopies(p protocol.Project, host string, manifest []byte, copies map[string]string, remove []string) []protocol.FileEdit {
	var edits []protocol.FileEdit
	if len(remove) == 0 {
		return edits
	}
	trigger := triggerOf(host, copies["scripts/start.mjs"])
	replacement := map[string]string{
		"scripts/assemble.mjs": AssembleCommand(host),
		"scripts/package.mjs":  PackageTask(host, p.Root),
		"scripts/start.mjs":    StartTask(host, p.Root, trigger),
	}
	if doc, err := jsondoc.ParseDocument(manifest); err == nil {
		if scripts := doc.Object("scripts", false); scripts != nil {
			changed := scripts.RewriteStrings(scripts.Keys(), func(s string) string {
				for _, rel := range remove {
					s = strings.ReplaceAll(s, "node "+rel, replacement[rel])
				}
				return s
			})
			if changed {
				edits = append(edits, protocol.FileEdit{Path: path.Join(p.Root, "package.json"), Content: string(doc.Bytes())})
			}
		}
	}
	return append(edits, deleteEdits(p, CopiedScripts, remove)...)
}

// triggerOf reads the trigger of a function from the copy of its start script.
// A host whose start command has no trigger, or a missing copy, gives http.
func triggerOf(host, start string) string {
	switch {
	case host == AWS && start != "" && !strings.Contains(start, "start-api"):
		return "timer"
	case host == GCP && strings.Contains(start, "processMessage"):
		return "queue"
	case host == GCP && strings.Contains(start, "processObject"):
		return "blob"
	}
	return TriggerHTTP
}

// packCommand is the command that packs an app, for a message.
func packCommand(host string) string {
	if host == Electron {
		return "npx electron-forge package"
	}
	return "smith func package"
}
