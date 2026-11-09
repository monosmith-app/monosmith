package why

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/agentdoc"
	"github.com/monosmith-app/monosmith/apps/smith/agentfiles"
	"github.com/monosmith-app/monosmith/apps/smith/derive"
	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Inputs are what Explain reads. Plan is nil when derived files are off.
type Inputs struct {
	// Record is smith.lock.
	Record derivelock.Record
	// Plan is the derive plan, for the state of each derived file.
	Plan *derive.Plan
	// Graph is the project graph.
	Graph *projectgraph.Graph
	// Stacks are the stacks of smith.toml.
	Stacks map[string]workspace.StackConfig
	// Head returns the first lines of a workspace file, or "" when it is
	// missing or unreadable.
	Head func(path string) string
}

// Explain says who owns the workspace-relative path (forward slashes), what
// produced it and how to change it. A path smith knows nothing about gets an
// Explanation with Known false: it is not an error.
func Explain(in Inputs, path string) Explanation {
	e := Explanation{Path: path}
	if in.Graph != nil {
		if p := in.Graph.Owner(path); p != nil {
			e.Project = p.Name
			e.Stack = p.Stack
			if st, ok := in.Stacks[p.Stack]; ok {
				e.Choices = st.Choices
			}
		}
	}
	switch {
	case len(in.Record.EntriesFor(path)) > 0:
		explainDerived(in, &e)
	case path == workspace.FileName:
		e.Known, e.Kind, e.Owner = true, KindWorkspace, OwnerYou
		e.Summary = "You own smith.toml. smith reads it, and smith add, remove and config edit it for you."
		e.Source = "your decisions about the workspace"
		e.Change = []string{"edit smith.toml, or run smith config set <key> <value>"}
	case path == derivelock.FileName:
		e.Known, e.Kind, e.Owner = true, KindGenerated, OwnerSmith
		e.Summary = "smith owns smith.lock. It records which derived files smith owns, with a hash of each."
		e.Rule = "smith derive"
		e.Change = []string{"do not edit it", "commit it with the files it records", "smith derive --eject <file> changes one owner"}
	default:
		if !explainGenerated(in, &e) {
			explainPlain(&e)
		}
	}
	if e.Change == nil {
		e.Change = []string{}
	}
	return e
}

// explainDerived fills e from the record entries of the path.
func explainDerived(in Inputs, e *Explanation) {
	e.Known = true
	entries := in.Record.EntriesFor(e.Path)
	owned := 0
	var rules, modes []string
	for _, en := range entries {
		part := Part{Renderer: en.Renderer, Version: en.Version, Mode: en.Mode, Owner: en.Owner}
		if in.Plan != nil {
			for _, it := range in.Plan.Items {
				if it.File.Path == e.Path && it.File.Renderer == en.Renderer {
					part.State, part.Detail = string(it.Status), it.Detail
					if e.Project == "" {
						e.Project = it.File.Project
					}
				}
			}
		}
		e.Parts = append(e.Parts, part)
		rules = append(rules, fmt.Sprintf("%s v%d", en.Renderer, en.Version))
		modes = append(modes, en.Mode)
		if en.Owner != derivelock.OwnerYou {
			owned++
		}
	}
	e.Rule = "renderer " + strings.Join(rules, ", ")
	if e.Project != "" {
		e.Source = fmt.Sprintf("the facts and the project graph of %s", e.Project)
		if in.Graph != nil {
			if p := in.Graph.Get(e.Project); p != nil && len(p.Deps) > 0 {
				e.Source += " (it depends on " + strings.Join(p.Deps, ", ") + ")"
			}
		}
	} else {
		e.Source = "the facts and the project graph"
	}
	if owned == 0 {
		e.Kind, e.Owner = KindEjected, OwnerYou
		e.Summary = "You own this file. It was derived once (" + e.Rule + ") and was ejected, so smith does not touch it."
		e.Change = []string{"edit the file by hand", "smith derive --replace " + e.Path + " gives it back to smith"}
	} else {
		e.Kind, e.Owner = KindDerived, OwnerSmith
		e.Summary = "smith owns " + ownedText(modes) + " of this file. " + e.Rule + " writes it from the facts and the project graph."
		e.Change = []string{
			"change the input (a dependency, a fact or smith.toml), then run smith derive",
			"to keep your own version, run smith derive --eject " + e.Path,
		}
	}
}

// ownedText says which part of the file smith owns, from the modes.
func ownedText(modes []string) string {
	var parts []string
	for _, m := range modes {
		switch m {
		case "file":
			parts = append(parts, "the whole")
		case "region":
			parts = append(parts, "the lines between the markers")
		case "keys":
			parts = append(parts, "the derived keys")
		default:
			parts = append(parts, m)
		}
	}
	return strings.Join(parts, " and ")
}

// explainGenerated recognises a file a smith command wrote, from its path or
// its header. It reports whether it did.
func explainGenerated(in Inputs, e *Explanation) bool {
	head := ""
	if in.Head != nil {
		head = in.Head(e.Path)
	}
	switch {
	case e.Path == agentfiles.AgentPath && strings.Contains(head, "smith:generated"):
		e.Rule = "smith agent install"
		e.Summary = "smith owns this file while its marker matches. smith upgrade rewrites it."
		e.Change = []string{"edit the file: smith then keeps your version and never overwrites it", "smith agent install shows the difference"}
	case strings.Contains(head, agentdoc.BeginMarker):
		e.Rule = "smith agent file"
		e.Source = "the conventions, the projects and smith.lock of the workspace"
		e.Summary = "smith owns the lines between the smith markers. You own the rest."
		e.Change = []string{"write your own text outside the markers", "to change the generated lines, change the workspace, then run smith agent file", "smith agent file --check fails in CI when the file is out of date"}
	case strings.Contains(head, "Generated by smith"):
		e.Rule = "smith ci generate"
		e.Source = "the projects of the workspace and the [ci] table of smith.toml"
		e.Summary = "smith owns this file. smith add and smith ci generate rewrite it."
		e.Change = []string{"change [ci] in smith.toml or the projects, then run smith ci generate", "smith ci generate --check reports drift"}
	default:
		return false
	}
	e.Known, e.Kind, e.Owner = true, KindGenerated, OwnerSmith
	return true
}

// writtenOnceAtRoot are the files at the root of the workspace that a plugin
// writes once, when smith add creates the first project that needs them.
var writtenOnceAtRoot = map[string]bool{
	"package.json": true, "tsconfig.base.json": true, "tsconfig.json": true,
	"eslint.config.mjs": true, "eslint.config.js": true, "ruff.toml": true,
	"requirements-dev.txt": true, "go.work": true, "Directory.Build.props": true,
	".prettierrc": true, ".editorconfig": true, ".gitignore": true,
}

// explainPlain fills e for a file that no record claims.
func explainPlain(e *Explanation) {
	if e.Project == "" && writtenOnceAtRoot[e.Path] {
		e.Known, e.Kind, e.Owner = true, KindProject, OwnerYou
		e.Rule = "scaffold: smith add writes it once, when the first project that needs it is created"
		e.Source = "the plugin of the first project that needs it"
		e.Summary = "Written once by smith add; yours now. smith does not rewrite it."
		e.Change = []string{"edit it as you like"}
		return
	}
	if e.Project == "" {
		e.Kind = KindUnknown
		e.Summary = "smith does not know this file. No project contains it and no record lists it."
		e.Change = []string{"edit it as you like: smith does not read or write it"}
		return
	}
	e.Known, e.Kind, e.Owner = true, KindProject, OwnerYou
	e.Summary = "You own this file. It is in the project " + e.Project + ", and no record in smith.lock lists it."
	e.Change = []string{"edit it as you like"}
	if len(e.Choices) > 0 {
		e.Source = fmt.Sprintf("the stack %s of the project (choices: %s), if a plugin scaffolded it", e.Stack, strings.Join(e.Choices, ", "))
		e.Rule = "scaffold: a plugin writes the files of a choice once, when it creates the project"
		e.Summary += " If a choice scaffolded it, the plugin wrote it once and it is yours now."
		e.Change = append(e.Change, "a choice does not rewrite the file later: change the choices with smith choices and smith add --with for a new project")
	} else if e.Stack != "" {
		e.Source = "the stack " + e.Stack + " of the project, if a plugin scaffolded it"
		e.Summary += " If the stack scaffolded it, the plugin wrote it once and it is yours now."
	}
}
