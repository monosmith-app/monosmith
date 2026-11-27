package recipewizard

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspacerecipe"
)

// planFromWorkspace builds the plan of `smith recipe new --from-workspace`:
// the levels the workspace at root already uses (see package
// workspacerecipe), fixed, under the name and version given or asked. It
// prints what the recipe fixes before anything is checked or written, and
// returns the plan with the file it goes to by default,
// recipes/<name>.recipe.toml in the workspace.
//
// On a terminal it asks what the flags leave out: the language when the
// workspace has several, the name and the version (0.1.0 by default).
// Without one, --name is needed. It returns an error outside a workspace,
// next to --from-answers, --fixed or --chosen, and when the workspace has no
// project of the language.
func planFromWorkspace(o Options, root string, in *bufio.Reader) (Plan, string, error) {
	if root == "" {
		return Plan{}, "", errors.New("--from-workspace reads a workspace, and there is no smith.toml here or above")
	}
	if o.FromAnswers != "" || len(o.Fixed) > 0 || len(o.Chosen) > 0 {
		return Plan{}, "", errors.New("--from-workspace fixes the plugins the workspace uses: leave out --from-answers, --fixed and --chosen")
	}
	var facts workspacerecipe.Facts
	if o.Facts != nil {
		facts = *o.Facts
	} else {
		var err error
		if facts, err = workspacerecipe.Load(root, o.Version); err != nil {
			return Plan{}, "", err
		}
	}

	language := o.Language
	if langs := workspacerecipe.Languages(facts); language == "" && o.Interactive && len(langs) > 1 {
		fmt.Fprintf(o.Out, "The projects of this workspace use %s. Which language does the recipe fix? [%s]: ", strings.Join(langs, ", "), langs[0])
		language = readLine(in)
	}
	reading, err := workspacerecipe.Read(facts, language)
	if err != nil {
		return Plan{}, "", err
	}

	name, version := o.Name, o.RecipeVersion
	if name == "" {
		if !o.Interactive {
			return Plan{}, "", errors.New("--from-workspace needs the recipe's name: pass --name (lowercase words joined by dashes)")
		}
		for {
			fmt.Fprint(o.Out, "Recipe name (lowercase words joined by dashes, like team-go): ")
			line, readErr := in.ReadString('\n')
			name = strings.TrimSpace(line)
			problem := planProblem(name, reading.Fixed)
			if problem == "" {
				break
			}
			if readErr != nil {
				return Plan{}, "", errors.New("no valid recipe name was given: nothing was written")
			}
			fmt.Fprintln(o.Out, problem)
		}
		if version == "" {
			fmt.Fprint(o.Out, "Version [0.1.0]: ")
			version = readLine(in)
		}
	}

	fmt.Fprintln(o.Out, describeReading(reading))
	for _, f := range reading.Fixed {
		level, plugins, _ := strings.Cut(f, "=")
		fmt.Fprintf(o.Out, "  %-10s %s\n", level, strings.ReplaceAll(plugins, ",", ", "))
	}
	plan, err := PlanFromFlags(name, version, o.Extends, reading.Fixed, nil)
	if err != nil {
		return Plan{}, "", err
	}
	target := filepath.Join(root, "recipes", plan.Name+".recipe.toml")
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, target); err == nil {
			target = rel
		}
	}
	return plan, target, nil
}

// planProblem says what is wrong with a recipe of that name fixing those
// level=plugin values, or "" when nothing is.
func planProblem(name string, fixed []string) string {
	if _, err := PlanFromFlags(name, "", "", fixed, nil); err != nil {
		return err.Error()
	}
	return ""
}

// describeReading is the line before the levels a workspace recipe fixes.
func describeReading(r workspacerecipe.Reading) string {
	stack := "no stack, the language plugin alone"
	if r.StackName != "" {
		stack = "the stack " + r.StackName
	}
	return fmt.Sprintf("From this workspace (%d %s project(s), %s), the recipe fixes:", r.Projects, r.Language, stack)
}

// readLine is one answer typed on the terminal, without its spaces.
func readLine(in *bufio.Reader) string {
	line, _ := in.ReadString('\n')
	return strings.TrimSpace(line)
}
