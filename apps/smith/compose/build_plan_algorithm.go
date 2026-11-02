package compose

import (
	"strconv"
	"strings"
)

// PlanLine is one path the add writes or changes.
type PlanLine struct {
	// Action is create or change.
	Action string `json:"action"`
	Path   string `json:"path"`
	Note   string `json:"note,omitempty"`
}

// Plan is where a complete selection ends: what to run, what it writes, what
// it needs and how to keep it as a recipe.
type Plan struct {
	// Command is the exact line to run, with Argv as its words.
	Command string   `json:"command"`
	Argv    []string `json:"argv"`
	// Files are the paths the add writes or changes, as `smith add --dry-run`
	// lists them.
	Files []PlanLine `json:"files"`
	// Tools are the programs the selection needs on the computer.
	Tools []Tool `json:"tools"`
	// Notes say in plain words what is not obvious: the defaults that were
	// added and why.
	Notes []string `json:"notes"`
	// RecipeCommand writes the plugins of the selection as a recipe; RecipeNote
	// says what a recipe keeps and what it does not.
	RecipeCommand string   `json:"recipeCommand,omitempty"`
	RecipeArgv    []string `json:"recipeArgv,omitempty"`
	RecipeNote    string   `json:"recipeNote,omitempty"`
}

// PlanInput is what the plan is built from.
type PlanInput struct {
	// Kind is app, lib or package; Name is the project name, or empty to show
	// <name> in the command.
	Kind, Name string
	// Check is the complete check of the picks.
	Check Check
	// Files are the lines of the add plan, from scaffold.PlanAdd.
	Files []PlanLine
	// Tools finds the tools of a pick; nil means none.
	Tools ToolSource
	// Plugins are the plugin names the selection builds with: the scaffold, then
	// the toolchain. Language is the language plugin's name.
	Language, Scaffold string
	Toolchain          []string
}

// BuildPlan makes the plan of a complete check. It writes nothing.
//
// Parameters: in is the complete check with the add plan's lines.
// Returns: the plan; for an incomplete check it holds only the notes.
func BuildPlan(in PlanInput) Plan {
	p := Plan{Files: in.Files, Tools: []Tool{}, Notes: []string{}}
	if in.Kind == "" {
		in.Kind = "app"
	}
	name := in.Name
	if name == "" {
		name = "<name>"
	}
	if !in.Check.Complete {
		p.Notes = append(p.Notes, "The selection is not complete yet: "+firstNonEmpty(in.Check.Conflict, in.Check.Missing, "answer the open question")+".")
		return p
	}
	p.Argv = []string{"smith", "add", in.Kind, name, "--lang", in.Check.Language, "--with", strings.Join(in.Check.Selection, ",")}
	p.Command = strings.Join(p.Argv, " ")
	if len(in.Check.Added) > 0 {
		p.Notes = append(p.Notes, strings.Join(in.Check.Added, " and ")+" "+verb(len(in.Check.Added))+" added by default because a choice needs that kind of tool. Remove "+pronoun(len(in.Check.Added))+" from --with only if you pick another tool of that kind.")
	}
	seen := map[string]bool{}
	if in.Tools != nil {
		for _, c := range in.Check.Selection {
			for _, t := range in.Tools(c, in.Check.Language) {
				if key := t.Name + "/" + t.Binary; !seen[key] {
					seen[key] = true
					p.Tools = append(p.Tools, t)
				}
			}
		}
	}
	p.RecipeArgv, p.RecipeNote = recipeOf(in)
	p.RecipeCommand = strings.Join(quoteWords(p.RecipeArgv), " ")
	return p
}

// recipeOf is the `smith recipe new` line that fixes the plugins of the
// selection, and what a recipe keeps.
func recipeOf(in PlanInput) ([]string, string) {
	if in.Language == "" {
		return nil, ""
	}
	argv := []string{"smith", "recipe", "new", "--name", strings.Join(in.Check.Selection, "-"), "--fixed", "language=" + in.Language}
	if in.Scaffold != "" {
		argv = append(argv, "--fixed", "scaffold="+in.Scaffold)
	}
	if len(in.Toolchain) > 0 {
		argv = append(argv, "--fixed", "toolchain="+strings.Join(in.Toolchain, ","))
	}
	note := "A recipe fixes the plugins, not the choices: another project made from it can pick a different framework of the same plugin. smith add --with already saves this exact selection as a stack in smith.toml."
	return argv, note
}

// quoteWords quotes the words that have a space or a quote, so the line can be
// pasted into a shell.
func quoteWords(argv []string) []string {
	out := make([]string, len(argv))
	for i, w := range argv {
		out[i] = w
		if strings.ContainsAny(w, " 	\"'") {
			out[i] = strconv.Quote(w)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func verb(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}

func pronoun(n int) string {
	if n == 1 {
		return "it"
	}
	return "them"
}
