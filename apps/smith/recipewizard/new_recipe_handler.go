package recipewizard

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Options are what `smith recipe new` runs with.
type Options struct {
	// Target is where the recipe is written; empty is <name>.recipe.toml in
	// the current folder.
	Target string
	// FromAnswers is a Plan file; set, nothing is asked.
	FromAnswers string
	// Name, RecipeVersion, Extends and Fixed describe a recipe whose plugins
	// are all fixed (see PlanFromFlags); with Name or Fixed set, nothing is asked.
	Name, RecipeVersion, Extends string
	Fixed                        []string
	// Force replaces an existing file without asking.
	Force bool
	// Interactive says a person is there to answer.
	Interactive bool
	In          io.Reader
	Out         io.Writer
	// Version is smith's, which the built-in plugins report as their own.
	Version string
	// Names and Lookup replace the plugins smith knows (the built-in ones,
	// those smith.toml declares and those installed), for tests.
	Names  []string
	Lookup recipe.PluginLookup
}

// NewCommand is `smith recipe new`, added to `smith recipe`.
func NewCommand(version string) *cobra.Command {
	var opts Options
	cmd := &cobra.Command{
		Use:   "new [path]",
		Short: "Write a recipe by answering questions",
		Long: `Asks for a recipe's name, version and the recipe it extends, then goes through
the levels (language, scaffold, each task slot such as lint and test,
structure, deploy). For each it lists the plugins that fit what is chosen so
far, and you say whether one is fixed, a choice the recipe's users make
(python or csharp, say, with a plugin per answer), or left unset.

The recipe is checked for every combination of answers before it is written,
so a recipe this command writes resolves. It is written as the TOML you would
write by hand, to <name>.recipe.toml or the path you give. An existing file is
replaced only after you confirm, or with --force.

Without a terminal, --from-answers reads the same answers from a file, so a
script or a test gets the same recipe the questions would. A recipe whose
plugins are all fixed needs no file either: --name and --fixed level=plugin
(as often as needed) write it, which is what the interface's form runs. A
choice the recipe's users make still needs the questions or an answers file.`,
		Example: `  smith recipe new
  smith recipe new recipes/team/team.recipe.toml
  smith recipe new --from-answers answers.toml
  smith recipe new --name team --fixed language=python --fixed lint=ruff`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.Target = args[0]
			}
			opts.In, opts.Out, opts.Version = cmd.InOrStdin(), cmd.OutOrStdout(), version
			opts.Interactive = term.IsTerminal(int(os.Stdin.Fd())) && term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("CI") == ""
			return Run(opts)
		},
	}
	cmd.Flags().StringVar(&opts.FromAnswers, "from-answers", "", "read the answers from a TOML file instead of asking (see docs/reference/recipes.md)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "the recipe's name; with --fixed, nothing is asked")
	cmd.Flags().StringVar(&opts.RecipeVersion, "recipe-version", "", "the recipe's version (0.1.0 when not given)")
	cmd.Flags().StringVar(&opts.Extends, "extends", "", "the recipe this one extends: a path from the new file, or an https URL")
	cmd.Flags().StringArrayVar(&opts.Fixed, "fixed", nil, "fix a level's plugin, as level=plugin (language=python, structure=slices,layers); repeat for each level")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "replace the file if it exists")
	return cmd
}

// Run asks (or reads) the plan, validates the recipe it makes and writes it.
func Run(o Options) error {
	root, _ := workspace.FindRoot(".")
	var external map[string]workspace.ExternalPlugin
	if root != "" {
		if cfg, err := workspace.Load(root); err == nil {
			external = cfg.Plugins.External
		}
	}
	lookup := o.Lookup
	if lookup == nil {
		lookup = recipe.Catalog{Root: root, Version: o.Version, External: external}.Lookup
	}
	loader := &recipe.Loader{Root: root}
	in := bufio.NewReader(o.In)
	dir := "."
	if o.Target != "" {
		dir = filepath.Dir(o.Target)
	}

	var plan Plan
	switch {
	case o.FromAnswers != "":
		data, err := os.ReadFile(o.FromAnswers)
		if err != nil {
			return err
		}
		if plan, err = ReadPlan(data, o.FromAnswers); err != nil {
			return err
		}
	case o.Name != "" || len(o.Fixed) > 0:
		var err error
		if plan, err = PlanFromFlags(o.Name, o.RecipeVersion, o.Extends, o.Fixed); err != nil {
			return err
		}
	case o.Interactive:
		names := o.Names
		if names == nil {
			names = PluginNames(root, external)
		}
		candidates, skipped := NewCandidates(names, lookup)
		for _, s := range skipped {
			fmt.Fprintf(o.Out, "not offered, %s\n", s)
		}
		name := ""
		if o.Target != "" {
			name = strings.TrimSuffix(filepath.Base(o.Target), ".recipe.toml")
		}
		var err error
		plan, err = Asker{
			In: in, Out: o.Out, Candidates: candidates, Name: name, Version: "0.1.0",
			Parent: func(extends string) (recipe.Loaded, error) { return loader.Load(refIn(dir, extends)) },
		}.Ask()
		if err != nil {
			return err
		}
	default:
		return errors.New("smith recipe new asks questions, and there is no terminal to ask them on: write the answers in a file and use --from-answers (docs/reference/recipes.md)")
	}

	target := o.Target
	if target == "" {
		target = plan.Name + ".recipe.toml"
	}
	if !strings.HasSuffix(target, ".recipe.toml") {
		fmt.Fprintf(o.Out, "note: smith finds recipe projects by the name *.recipe.toml, and %s does not end that way\n", target)
	}
	text, err := Creator{Loader: loader, Lookup: lookup}.Render(plan, target)
	if err != nil {
		return fmt.Errorf("nothing was written: %w", err)
	}
	if err := Write(target, text, o.Force); errors.Is(err, ErrExists) && o.Interactive {
		fmt.Fprintf(o.Out, "%s exists. Replace it? [y/N]: ", target)
		answer, _ := in.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			return fmt.Errorf("%w: left as it is", err)
		}
		err = Write(target, text, true)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	fmt.Fprintf(o.Out, "\nWrote %s (recipe %s %s).\nCheck it any time with: smith recipe check %s --all\n", target, plan.Name, plan.Version, target)
	return nil
}
