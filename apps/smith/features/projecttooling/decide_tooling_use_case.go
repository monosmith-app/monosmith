package projecttooling

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/shared/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/shared/problem"
)

// DecideRequest is what Decide needs.
type DecideRequest struct {
	// Projects are the assessed projects; those that are not Node are skipped.
	Projects []Assessment
	// Given are the --tooling answers.
	Given Answers
	// Interactive says a person can answer; Yes is --yes; Preview is a dry run.
	Interactive, Yes, Preview bool
	In                        *bufio.Reader
	Out                       io.Writer
	// Flag is how the command spells the flag in its messages ("--tooling").
	Flag string
}

// Decide settles, for each Node project, whether it keeps its own build and
// lint tools or uses smith's. A --tooling answer settles a project. In a
// terminal each open project is a question that recommends smith's tooling
// when a tool of the project is legacy. With --yes (or in a dry run) an open
// project keeps its own tools, because that changes nothing, and a line says
// how to choose smith's. Without a terminal and without --yes it refuses and
// names the flag: smith never picks for the user.
//
// Parameters: r is the request.
// Returns: the answer of each Node project, by name; an error when nobody can
// answer an open project.
func Decide(r DecideRequest) (map[string]Answer, error) {
	flag := r.Flag
	if flag == "" {
		flag = "--tooling"
	}
	out := map[string]Answer{}
	var open []Assessment
	for _, p := range r.Projects {
		if !p.Node {
			continue
		}
		if a, ok := r.Given.For(p.Name); ok {
			out[p.Name] = a
			continue
		}
		open = append(open, p)
	}
	if len(open) == 0 {
		return out, nil
	}
	asks := r.Interactive && !r.Yes && !r.Preview
	if !asks && !r.Yes && !r.Preview {
		var names []string
		for _, p := range open {
			names = append(names, p.Name)
		}
		return out, problem.New("nobody can say which build and lint tools "+strings.Join(names, ", ")+" should use: there is no terminal and --yes was not given",
			flag+" smith|keep   (smith: smith's build and ESLint config replace the old tools; keep: nothing changes)",
			flag+" <project>=smith|keep answers one project")
	}
	for _, p := range open {
		if !asks {
			out[p.Name] = Keep
			if p.Recommended == Smith {
				fmt.Fprintf(r.Out, "Tooling of %s: keep its own (smith recommends its own tooling: %s; %s %s=smith uses it)\n", p.Name, strings.Join(p.Legacy, ", "), flag, p.Name)
			} else {
				fmt.Fprintf(r.Out, "Tooling of %s: keep its own (%s %s=smith uses smith's build and ESLint config)\n", p.Name, flag, p.Name)
			}
			continue
		}
		answer := choiceprompt.Ask(Question(p), Options(p), string(p.Recommended), r.In, r.Out)
		out[p.Name] = Answer(answer)
		fmt.Fprintf(r.Out, "Tooling of %s: %s (change it with %s %s=smith|keep)\n", p.Name, answer, flag, p.Name)
	}
	return out, nil
}

// Question is the words that ask a project's question.
//
// Parameters: p is the project.
// Returns: the question.
func Question(p Assessment) string {
	if len(p.Legacy) == 0 {
		return p.Name + " builds and lints with its own tools. Keep them, or use smith's build and ESLint config?"
	}
	return p.Name + " uses " + strings.Join(p.Legacy, ", ") + ". Keep its own tools, or use smith's build and ESLint config?"
}

// Options are the answers of a project's question, the recommended one first.
//
// Parameters: p is the project.
// Returns: the options.
func Options(p Assessment) []choiceprompt.Option {
	smith := choiceprompt.Option{Label: "Use smith's", Value: string(Smith), Hint: "smith's build and @monosmith/eslint-config go in first; then each group of old tools is shown and removed"}
	keep := choiceprompt.Option{Label: "Keep its own", Value: string(Keep), Hint: "nothing of its build or lint changes"}
	if p.Recommended == Smith {
		smith.Hint += " (recommended)"
		return []choiceprompt.Option{smith, keep}
	}
	keep.Hint += " (recommended: its tools are current)"
	return []choiceprompt.Option{keep, smith}
}
