package electronapp

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// MoveDescription is the description of the fix, the same for every project
// so that doctor folds them.
const MoveDescription = "Pack the Electron app in its work folder"

// MoveOutDirFix is the presented fix that moves where an adopted Forge or
// electron-builder setup packs to its work folder (WantedOutDir). It is a
// choice between the move and leaving the project as it is: smith never moves
// it unasked. It edits one file, the setup's Config, and nothing else.
//
// Parameters: workspaceRoot is the absolute workspace folder; p the project;
// s its setup.
// Returns the fix, or nil when the setup is not an adopted one or smith
// cannot edit its configuration safely.
func MoveOutDirFix(workspaceRoot string, p protocol.Project, s Setup) *protocol.Fix {
	if (s.Tool != Forge && s.Tool != Builder) || !s.Editable || s.Config == "" {
		return nil
	}
	want := WantedOutDir(p.Root, s.Tool)
	value := protocol.FromProject(p.Root, want)
	file := path.Join(p.Root, s.Config)
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(file)))
	// A Builder setup without a configuration file takes a new "build" key in
	// package.json, which always exists.
	if err != nil && (s.Tool != Builder || !os.IsNotExist(err)) {
		return nil
	}
	content, ok := edited(string(raw), s, value)
	if !ok {
		return nil
	}
	setting := "outDir"
	if s.Tool == Builder {
		setting = "directories.output"
	}
	from := "the project folder"
	if s.OutDir != "" {
		from = path.Join(p.Root, s.OutDir)
	}
	move := protocol.Fix{
		Description: MoveDescription,
		Files:       []protocol.FileEdit{{Path: file, Content: content}},
	}
	return &protocol.Fix{
		Description: MoveDescription,
		Choices: []protocol.FixChoice{{
			Question: p.Name + " packs its app into " + from + ". smith keeps the work files of a tool in " + WorkDir(p.Root) +
				" and links the packed app into " + protocol.ProjectBuildDir(p.Root) + ". Moving it sets " + setting + " to " + value + " in " + file +
				". The next screen shows the changed lines. The old folder stays on disk and nothing is committed.",
			Options: []protocol.FixChoiceOption{
				{Label: "pack " + p.Name + " into " + want, Fix: move},
				{Label: "leave " + p.Name + " as it is (nothing changes)"},
			},
		}},
	}
}

// edited is the configuration text with the output folder set to value, and
// false when smith cannot set it safely.
func edited(text string, s Setup, value string) (string, bool) {
	switch {
	case strings.HasSuffix(s.Config, ".json"):
		if strings.TrimSpace(text) == "" {
			text = "{}"
		}
		doc, err := jsondoc.ParseDocument([]byte(text))
		if err != nil {
			return "", false
		}
		target := doc
		if s.Config == "package.json" {
			if s.Tool == Forge {
				target = doc.Object("config", true).Object("forge", true)
			} else {
				target = doc.Object("build", true)
			}
		}
		if s.Tool == Forge {
			target.SetString("outDir", value)
		} else {
			target.Object("directories", true).SetString("output", value)
		}
		return lineendings.Match(text, string(doc.Bytes())), true
	case s.Tool == Builder && (strings.HasSuffix(s.Config, ".yml") || strings.HasSuffix(s.Config, ".yaml")):
		if yamlDirectories.MatchString(text) {
			return "", false
		}
		if text != "" && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		return lineendings.Match(text, text+"directories:\n  output: "+value+"\n"), true
	case s.Tool == Forge:
		if m := outDirString.FindStringSubmatchIndex(text); m != nil {
			return text[:m[2]] + value + text[m[3]:], true
		}
		loc := exportedObject.FindStringIndex(text)
		if loc == nil {
			return "", false
		}
		return text[:loc[1]] + lineendings.Match(text, "\n  outDir: '"+value+"',") + text[loc[1]:], true
	}
	return "", false
}
