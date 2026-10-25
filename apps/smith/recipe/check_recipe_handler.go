package recipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
)

// ErrInvalid is returned by `smith recipe check` after it has printed why a
// recipe is not valid, so the caller exits non-zero without printing it again.
var ErrInvalid = errors.New("the recipe is not valid")

// checkReport is `smith recipe check --json`: the resolution, and the errors
// that stopped it from being valid.
type checkReport struct {
	Resolution
	// Pinned says URL recipes were pinned in the workspace's lockfile.
	Pinned bool `json:"pinned"`
	// Checked is how many combinations of answers --all resolved and validated.
	Checked int      `json:"checked,omitempty"`
	Errors  []string `json:"errors"`
}

// Command is `smith recipe`.
func Command(version string) *cobra.Command {
	cmd := &cobra.Command{Use: "recipe", Short: "Check recipes: a team's choice of plugins and settings"}
	var choices []string
	var asJSON, all bool
	check := &cobra.Command{
		Use:   "check <file-or-url>",
		Short: "Resolve and validate a recipe, showing what it would use and where each value came from",
		Long: `Loads a recipe and the recipes it extends (a path, or an https URL pinned by
content hash in ` + LockFile + `), answers its choices, and checks the result with
the rules a stack is checked with: one plugin per exclusive level and slot,
requirements met, no conflicts, every plugin installed at a version in range.

Choices are answered with --choice name=answer, else by their default. It
prints the choices asked, the plugin for each level and the merged settings,
each with the recipe, file and line it came from, or why the recipe is not
valid. The exit code is non-zero when it is not. --json prints the same for
tools.`,
		Example: `  smith recipe check team.recipe.toml
  smith recipe check team.recipe.toml --choice language=csharp --choice kind=service
  smith recipe check https://example.com/recipes/team.recipe.toml --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			answers, err := parseAnswers(choices)
			if err != nil {
				return err
			}
			return runCheck(cmd.OutOrStdout(), cmd.ErrOrStderr(), args[0], answers, asJSON, all, version)
		},
	}
	check.Flags().BoolVar(&all, "all", false, "check every combination of answers (up to 64), not only the one given or defaulted")
	check.Flags().StringArrayVar(&choices, "choice", nil, "answer a choice: name=answer (repeatable)")
	check.Flags().BoolVar(&asJSON, "json", false, "print JSON")
	cmd.AddCommand(check)
	return cmd
}

func parseAnswers(pairs []string) (map[string]string, error) {
	answers := map[string]string{}
	for _, p := range pairs {
		name, answer, ok := strings.Cut(p, "=")
		if !ok || name == "" || answer == "" {
			return nil, fmt.Errorf("--choice %q: write it as name=answer", p)
		}
		answers[name] = answer
	}
	return answers, nil
}

func runCheck(out, errOut io.Writer, ref string, answers map[string]string, asJSON, all bool, version string) error {
	root, _ := workspace.FindRoot(".")
	var external map[string]workspace.ExternalPlugin
	if root != "" {
		if cfg, err := workspace.Load(root); err == nil {
			external = cfg.Plugins.External
		}
	}
	report := checkReport{Pinned: root != ""}
	loaded, err := (&Loader{Root: root}).Load(ref)
	lookup := Catalog{Root: root, Version: version, External: external}.Lookup
	if err == nil && all {
		// Every combination of answers; the report below still shows the one
		// given or defaulted, when it resolves.
		report.Checked, err = ValidateAll(loaded, lookup)
	}
	resolved := false
	if err == nil {
		report.Resolution, err = Resolve(loaded, answers)
		resolved = err == nil
		var missing *MissingAnswerError
		if all && errors.As(err, &missing) {
			// Every branch checked out; there is only no single answer to show.
			err = nil
		}
	}
	if err == nil && resolved {
		err = Validate(report.Resolution, lookup)
	}
	if err != nil {
		report.Errors = strings.Split(err.Error(), "\n")
	}
	if asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if encErr := enc.Encode(report); encErr != nil {
			return encErr
		}
	} else {
		printReport(out, report)
		for _, line := range report.Errors {
			fmt.Fprintln(errOut, "error: "+line)
		}
	}
	if err != nil {
		return ErrInvalid
	}
	return nil
}

func printReport(out io.Writer, r checkReport) {
	if r.Name != "" {
		fmt.Fprintf(out, "Recipe %s %s\n", r.Name, r.Version)
		var chain []string
		for _, l := range r.Chain {
			chain = append(chain, fmt.Sprintf("%s (%s)", strings.TrimSpace(l.Name+" "+l.Version), l.Source))
		}
		fmt.Fprintf(out, "  chain: %s\n", strings.Join(chain, " > "))
	}
	if len(r.Asked) > 0 {
		fmt.Fprintln(out, "\nChoices")
		for _, q := range r.Asked {
			how := "default"
			switch {
			case q.Answer == "":
				how = "needs an answer"
			case q.Given:
				how = "answered"
			}
			opts := ""
			if len(q.Options) > 0 {
				opts = "; options " + strings.Join(q.Options, ", ")
			}
			ask := ""
			if q.Ask != "" {
				ask = " " + strings.TrimSuffix(q.Ask, "?") + "?"
			}
			fmt.Fprintf(out, "  %-12s %-14s (%s%s)%s\n", q.Name, orDash(q.Answer), how, opts, ask)
		}
	}
	if len(r.Uses) > 0 {
		fmt.Fprintln(out, "\nPlugins")
		for _, u := range r.Uses {
			fmt.Fprintf(out, "  %-12s %-24s %s\n", u.Key, strings.Join(u.Plugins, ", "), u.Origin)
		}
	}
	if len(r.Origins) > 0 {
		fmt.Fprintln(out, "\nSettings")
		paths := make([]string, 0, len(r.Origins))
		for p := range r.Origins {
			if strings.HasPrefix(p, "settings.") {
				paths = append(paths, p)
			}
		}
		sort.Strings(paths)
		for _, p := range paths {
			fmt.Fprintf(out, "  %-40s %-16v %s\n", strings.TrimPrefix(p, "settings."), settingAt(r.Settings, strings.TrimPrefix(p, "settings.")), r.Origins[p])
		}
	}
	if len(r.Errors) == 0 {
		if r.Checked > 0 {
			fmt.Fprintf(out, "\nChecked %d combination(s) of answers: every one is valid.\n", r.Checked)
		}
		fmt.Fprintln(out, "\nThe recipe is valid.")
	}
}

// settingAt reads a setting by its dotted path (plugin.key.key).
func settingAt(settings map[string]any, path string) any {
	var cur any = settings
	for _, part := range strings.Split(path, ".") {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[part]
	}
	return cur
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
