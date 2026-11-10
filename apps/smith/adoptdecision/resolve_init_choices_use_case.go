package adoptdecision

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// Providers are the CI providers smith writes pipelines for.
var Providers = []string{"github", "azure", "gitlab", "circleci", "bitbucket"}

// Request is what Resolve needs.
type Request struct {
	Flags Flags
	Facts Facts
	// Yes is --yes: take the safe default of every question and print it.
	Yes bool
	// Preview is a dry run: take the defaults without asking and without refusing.
	Preview bool
	// Interactive says a person can answer.
	Interactive bool
	In          *bufio.Reader
	Out         io.Writer
}

// asks says a person is asked: a terminal, no --yes, and not a dry run.
func (r Request) asks() bool { return r.Interactive && !r.Yes && !r.Preview }

// Resolve settles the choices of adopt --init. A flag settles its choice. In a
// terminal each open choice is a question that says what every answer does,
// with the safe one as the default. With --yes (or in a dry run) each open
// choice takes the safe default and a line says which and how to change it.
// Without a terminal and without --yes it refuses and names the flags,
// because nobody can answer. Safe means nothing is written that the person
// did not ask for: no .gitattributes, no agent files, the pipelines the
// repository already uses.
//
// Parameters: r is the request.
// Returns: the choices; an error for a flag value smith does not know or an
// open choice nobody can settle.
func Resolve(r Request) (Choices, error) {
	var c Choices
	if err := validate(r.Flags); err != nil {
		return c, err
	}
	c.Gitattributes, c.Agent, c.Existing = r.Flags.Gitattributes, r.Flags.Agent, r.Flags.Existing
	if r.Flags.CIGiven {
		c.CI = withoutNone(r.Flags.CI)
	}
	needsGitattributes := c.Gitattributes == "" && !r.Facts.GitattributesExists
	needsAgent := c.Agent == ""
	needsCI := !r.Flags.CIGiven
	if !r.Interactive && !r.Yes && !r.Preview {
		var flags []string
		if needsGitattributes {
			flags = append(flags, "--gitattributes add|skip   (add writes `* text=auto eol=lf` and renormalises line endings; skip keeps yours)")
		}
		if needsAgent {
			flags = append(flags, "--agent add|skip|never   (add writes the agent file and merges the MCP server into .mcp.json)")
		}
		if needsCI {
			flags = append(flags, "--ci <provider>|none   (providers: "+strings.Join(Providers, ", ")+")")
		}
		if len(flags) > 0 {
			return c, problem.New("nobody can answer the questions of smith adopt --init: there is no terminal and --yes was not given",
				append(flags, "--yes takes the safe default of each and prints it")...)
		}
	}
	say := func(format string, a ...any) { fmt.Fprintf(r.Out, format+"\n", a...) }
	if c.Gitattributes == "" {
		switch {
		case r.Facts.GitattributesExists:
			c.Gitattributes = GitattributesSkip
		case r.asks():
			c.Gitattributes = scaffold.AskChoice("Line endings: should smith write a .gitattributes with `* text=auto eol=lf`? Git then converts every text file in the repository to LF.", []scaffold.NewOption{
				{Label: "No, keep my line endings", Value: GitattributesSkip, Hint: "nothing is renormalised (default)"},
				{Label: "Yes, write it", Value: GitattributesAdd, Hint: "text files become LF; on a CRLF repository this changes many files"},
			}, GitattributesSkip, r.In, r.Out)
			say("Line endings: %s (change it with --gitattributes add|skip)", gitattributesText(c.Gitattributes))
		default:
			c.Gitattributes = GitattributesSkip
			say("Line endings: %s (--gitattributes add writes `* text=auto eol=lf`)", gitattributesText(c.Gitattributes))
		}
	}
	if c.Agent == "" {
		merge := ""
		if r.Facts.MCPExists {
			merge = "; your .mcp.json is merged, your servers stay"
		}
		switch {
		case r.asks():
			c.Agent = scaffold.AskChoice("The MonoSmith agent lets Claude Code and VS Code run smith in this repository through an MCP server. Add its files?", []scaffold.NewOption{
				{Label: "Not now", Value: AgentSkip, Hint: "nothing is written; smith agent install adds it later (default)"},
				{Label: "Add them", Value: AgentAdd, Hint: ".claude/agents/smith.md, .mcp.json, .vscode/mcp.json" + merge},
				{Label: "Never", Value: AgentNever, Hint: "records [agent] enabled = false in smith.toml"},
			}, AgentSkip, r.In, r.Out)
			say("Agent files: %s (change it with --agent add|skip|never)", c.Agent)
		default:
			c.Agent = AgentSkip
			say("Agent files: skip (--agent add writes them%s; smith agent install does it later)", merge)
		}
	}
	if needsCI {
		c.CI = askCI(r, say)
	}
	c.Existing = settleExisting(r, c, say)
	return c, nil
}

// askCI settles the providers: the ones the repository uses by default.
func askCI(r Request, say func(string, ...any)) []string {
	detected := r.Facts.Detected
	fallback := []string{"github"}
	if len(detected) > 0 {
		fallback = detected
	}
	if !r.asks() {
		if len(detected) > 0 {
			say("CI: %s, the provider the repository already uses (--ci <provider>|none changes it)", strings.Join(detected, ", "))
		} else {
			say("CI: github, because no provider was detected (--ci <provider>|none changes it)")
		}
		return fallback
	}
	var options []scaffold.NewOption
	def := "default"
	label := "GitHub Actions"
	hint := "no pipeline was found; GitHub is the usual choice"
	if len(detected) > 0 {
		label, hint = strings.Join(detected, ", "), "the repository already uses it"
	}
	options = append(options, scaffold.NewOption{Label: label, Value: def, Hint: hint + " (default)"})
	for _, p := range Providers {
		if !slices.Contains(fallback, p) {
			options = append(options, scaffold.NewOption{Label: p, Value: p, Hint: "smith writes its pipeline for " + p})
		}
	}
	options = append(options, scaffold.NewOption{Label: "None", Value: NoCI, Hint: "smith writes no pipeline"})
	answer := scaffold.AskChoice("Which CI should smith write the pipelines for? It also writes .devcontainer and tools/smith/install.sh.", options, def, r.In, r.Out)
	switch answer {
	case "default":
		say("CI: %s (change it with --ci <provider>|none)", strings.Join(fallback, ", "))
		return fallback
	case NoCI:
		say("CI: none (change it with --ci <provider>)")
		return nil
	}
	say("CI: %s (change it with --ci <provider>|none)", answer)
	return []string{answer}
}

// settleExisting decides what happens to a pipeline of the team's own at a
// path smith would write: yours moves to <file>.pre-smith (the default, and
// nothing is lost), or yours stays and that provider is dropped.
func settleExisting(r Request, c Choices, say func(string, ...any)) string {
	if c.Existing != "" {
		return c.Existing
	}
	var own []string
	for _, p := range c.CI {
		if path := r.Facts.OwnPipelines[p]; path != "" {
			own = append(own, path)
		}
	}
	if len(own) == 0 {
		return ExistingReplace
	}
	if !r.asks() {
		say("Your own pipeline %s: replaced by smith's, yours kept as <file>.pre-smith (--ci-existing keep leaves it alone)", strings.Join(own, ", "))
		return ExistingReplace
	}
	answer := scaffold.AskChoice("You already have a pipeline: "+strings.Join(own, ", ")+". What should happen to it?", []scaffold.NewOption{
		{Label: "Replace it, keep a copy", Value: ExistingReplace, Hint: "smith writes its pipeline; yours moves to <file>.pre-smith (default)"},
		{Label: "Keep mine", Value: ExistingKeep, Hint: "smith writes no pipeline for that provider"},
	}, ExistingReplace, r.In, r.Out)
	say("Your own pipeline: %s (change it with --ci-existing replace|keep)", answer)
	return answer
}

// WithoutKept removes from the providers the ones whose own pipeline stays
// (Existing is ExistingKeep), because smith then writes nothing there.
//
// Parameters: c are the choices; own maps a provider to the path of the team's own pipeline.
// Returns: the providers smith writes.
func WithoutKept(c Choices, own map[string]string) []string {
	if c.Existing != ExistingKeep {
		return c.CI
	}
	var out []string
	for _, p := range c.CI {
		if own[p] == "" {
			out = append(out, p)
		}
	}
	return out
}

func gitattributesText(v string) string {
	if v == GitattributesAdd {
		return "add .gitattributes"
	}
	return "keep yours, no .gitattributes written"
}

func withoutNone(list []string) []string {
	var out []string
	for _, p := range list {
		if p != NoCI {
			out = append(out, p)
		}
	}
	return out
}

func validate(f Flags) error {
	if f.Gitattributes != "" && f.Gitattributes != GitattributesAdd && f.Gitattributes != GitattributesSkip {
		return problem.New(fmt.Sprintf("--gitattributes %s: use add or skip", f.Gitattributes))
	}
	if f.Agent != "" && f.Agent != AgentAdd && f.Agent != AgentSkip && f.Agent != AgentNever {
		return problem.New(fmt.Sprintf("--agent %s: use add, skip or never", f.Agent))
	}
	if f.Existing != "" && f.Existing != ExistingReplace && f.Existing != ExistingKeep {
		return problem.New(fmt.Sprintf("--ci-existing %s: use replace or keep", f.Existing))
	}
	for _, p := range f.CI {
		if p != NoCI && !slices.Contains(Providers, p) {
			return problem.New(fmt.Sprintf("--ci %s: smith writes pipelines for %s, or none", p, strings.Join(Providers, ", ")))
		}
		if p == NoCI && len(f.CI) > 1 {
			return problem.New("--ci none cannot be combined with a provider")
		}
	}
	return nil
}
