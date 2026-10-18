package recipewizard

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// maxTries is how many times one question is asked before giving up.
const maxTries = 3

// Asker builds a Plan by asking a person, one question at a time.
type Asker struct {
	In  io.Reader
	Out io.Writer
	// Candidates are the plugins that may be offered.
	Candidates *Candidates
	// Parent loads the recipe a person says the new one extends, so a wrong
	// path is caught while it is typed.
	Parent func(extends string) (recipe.Loaded, error)
	// Name and Version are the answers an empty line takes.
	Name, Version string
}

// asking is one run of the questions.
type asking struct {
	Asker
	in   *bufio.Reader
	plan Plan
	// parentChoices are the choices of the parent recipe, which a level may
	// use without declaring them again.
	parentChoices map[string]recipe.Choice
	scenarios     []Scenario
	lastChoice    string
}

// Ask goes through the recipe's name, version and parent, then each level a
// plugin can fill. For a level the person says whether its plugin is fixed,
// a choice the recipe's users make, or left to the parent or unset; only
// plugins that fit what is chosen so far, in every branch of the choices,
// are offered.
func (a Asker) Ask() (Plan, error) {
	k := &asking{Asker: a, in: bufio.NewReader(a.In), parentChoices: map[string]recipe.Choice{}, scenarios: []Scenario{{Answers: map[string]string{}}}}
	k.plan = Plan{Choices: map[string]recipe.Choice{}, Levels: map[string]LevelPlan{}}
	var err error
	if k.plan.Name, err = k.line("Name of the recipe", a.Name, func(v string) error { return (&Plan{Name: v, Version: "0.0.0"}).Check() }); err != nil {
		return Plan{}, err
	}
	if k.plan.Version, err = k.line("Version", a.Version, func(v string) error { return (&Plan{Name: k.plan.Name, Version: v}).Check() }); err != nil {
		return Plan{}, err
	}
	if k.plan.Extends, err = k.line("Extends another recipe (a path from the new file, or an https URL; enter for none)", "", k.checkParent); err != nil {
		return Plan{}, err
	}
	for _, level := range a.Candidates.Levels() {
		if err := k.level(level); err != nil {
			return Plan{}, err
		}
	}
	if len(k.plan.Levels) == 0 {
		return Plan{}, fmt.Errorf("nothing was chosen: a recipe needs at least a language, here or in the recipe it extends")
	}
	return k.plan, nil
}

func (k *asking) checkParent(v string) error {
	if v == "" {
		return nil
	}
	loaded, err := k.Parent(v)
	if err != nil {
		return err
	}
	last := loaded.Chain[len(loaded.Chain)-1].Recipe
	fmt.Fprintf(k.Out, "  extends %s %s\n", last.Name, last.Version)
	for _, doc := range loaded.Chain {
		for name, c := range doc.Recipe.Choices {
			k.parentChoices[name] = c
		}
	}
	if len(k.parentChoices) > 0 {
		fmt.Fprintf(k.Out, "  it asks: %s; a level can use these choices\n", strings.Join(keys(k.parentChoices), ", "))
	}
	return nil
}

// line asks one question. An empty line takes def; check may refuse a value,
// and the question is asked again, up to three times.
func (k *asking) line(question, def string, check func(string) error) (string, error) {
	for try := 0; try < maxTries; try++ {
		if def != "" {
			fmt.Fprintf(k.Out, "%s [%s]: ", question, def)
		} else {
			fmt.Fprintf(k.Out, "%s: ", question)
		}
		text, err := k.in.ReadString('\n')
		value := strings.TrimSpace(text)
		if err != nil && value == "" {
			return "", fmt.Errorf("the input ended while asking %q; use --from-answers to run without a terminal", question)
		}
		if value == "" {
			value = def
		}
		if check != nil {
			if problem := check(value); problem != nil {
				fmt.Fprintf(k.Out, "  %v\n", problem)
				continue
			}
		}
		return value, nil
	}
	return "", fmt.Errorf("no valid answer to %q after %d tries", question, maxTries)
}

// level asks how one level is filled.
func (k *asking) level(level string) error {
	// A plugin that fits one branch of the choices can fill the level as a
	// choice; only one that fits every branch can be fixed.
	every := k.Candidates.For(level, k.scenarios)
	if len(every) == 0 && !k.fitsSomeBranch(level) {
		return nil
	}
	fmt.Fprintf(k.Out, "\n%s: %d plugin(s) fit every branch of the choices so far\n", level, len(every))
	for i := 0; i < maxTries; i++ {
		mode, err := k.line("  fixed (f), a choice for the recipe's users (c), or leave unset (enter)", "", func(v string) error {
			switch strings.ToLower(v) {
			case "", "f", "fixed", "c", "choice", "s", "skip":
				return nil
			}
			return fmt.Errorf("answer f, c or nothing")
		})
		if err != nil {
			return err
		}
		switch strings.ToLower(mode) {
		case "f", "fixed":
			if len(every) == 0 {
				fmt.Fprintln(k.Out, "  no plugin fits every branch; make it a choice (c), or leave it unset")
				continue
			}
			for i, info := range every {
				fmt.Fprintf(k.Out, "  %d  %-14s %s\n", i+1, info.Plugin.Name, Describe(info))
			}
			names, err := k.pick(level, every)
			if err != nil || len(names) == 0 {
				return err
			}
			k.plan.Levels[level] = LevelPlan{Fixed: names}
			k.addMembers(k.scenarios, level, names)
			return nil
		case "c", "choice":
			return k.choice(level)
		}
		return nil
	}
	return nil
}

// fitsSomeBranch reports whether a plugin fits at least one scenario.
func (k *asking) fitsSomeBranch(level string) bool {
	for _, s := range k.scenarios {
		if len(k.Candidates.For(level, []Scenario{s})) > 0 {
			return true
		}
	}
	return false
}

// choice fills a level with a plugin per answer of a choice.
func (k *asking) choice(level string) error {
	suggested := k.lastChoice
	if suggested == "" {
		suggested = level
	}
	name, err := k.line("  name of the choice", suggested, func(v string) error {
		if !namePattern.MatchString(v) {
			return fmt.Errorf("a choice is named with lowercase words joined by dashes")
		}
		return nil
	})
	if err != nil {
		return err
	}
	options, err := k.options(name)
	if err != nil {
		return err
	}
	k.lastChoice = name
	var next []Scenario
	plugins := map[string]any{}
	for _, answer := range options {
		sub := consistent(k.scenarios, name, answer)
		candidates := k.Candidates.For(level, sub)
		if len(candidates) == 0 {
			fmt.Fprintf(k.Out, "  no plugin fits %s=%s; %s stays unset for it\n", name, answer, level)
			next = append(next, sub...)
			continue
		}
		fmt.Fprintf(k.Out, "  when %s=%s:\n", name, answer)
		for i, info := range candidates {
			fmt.Fprintf(k.Out, "    %d  %-14s %s\n", i+1, info.Plugin.Name, Describe(info))
		}
		names, err := k.pick(level, candidates)
		if err != nil {
			return err
		}
		if len(names) == 0 {
			next = append(next, sub...)
			continue
		}
		if len(names) == 1 {
			plugins[answer] = names[0]
		} else {
			plugins[answer] = names
		}
		k.addMembers(sub, level, names)
		next = append(next, sub...)
	}
	k.scenarios = next
	if len(plugins) > 0 {
		k.plan.Levels[level] = LevelPlan{Choice: name, Plugins: plugins}
	}
	return nil
}

// options settles the answers of a choice: those it already has, here or in
// the parent, or new ones the person types.
func (k *asking) options(name string) ([]string, error) {
	if c, ok := k.plan.Choices[name]; ok {
		fmt.Fprintf(k.Out, "  the choice %s already has the answers %s\n", name, strings.Join(c.Options, ", "))
		return c.Options, nil
	}
	if c, ok := k.parentChoices[name]; ok && len(c.Options) > 0 {
		fmt.Fprintf(k.Out, "  the choice %s comes from the parent, with the answers %s\n", name, strings.Join(c.Options, ", "))
		return c.Options, nil
	}
	question, err := k.line("  question to ask", titled(name), nil)
	if err != nil {
		return nil, err
	}
	var options []string
	if _, err := k.line("  answers, separated by commas", "", func(v string) error {
		options = nil
		for _, o := range strings.Split(v, ",") {
			if o = strings.TrimSpace(o); o != "" && !slices.Contains(options, o) {
				options = append(options, o)
			}
		}
		if len(options) < 2 {
			return fmt.Errorf("give at least two answers")
		}
		return nil
	}); err != nil {
		return nil, err
	}
	def, err := k.line("  default answer (enter for none)", "", func(v string) error {
		if v != "" && !slices.Contains(options, v) {
			return fmt.Errorf("the default must be one of %s", strings.Join(options, ", "))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	k.plan.Choices[name] = recipe.Choice{Ask: question, Options: options, Default: def}
	return options, nil
}

// pick asks which of the numbered plugins to use: one, or for the additive
// levels any number. Nothing leaves the level unset.
func (k *asking) pick(level string, candidates []protocol.InitializeResult) ([]string, error) {
	var names []string
	several := ""
	if Additive(level) {
		several = ", several separated by commas"
	}
	_, err := k.line("    use (a number or a name"+several+"; enter to leave unset)", "", func(v string) error {
		names = nil
		for _, token := range strings.FieldsFunc(v, func(r rune) bool { return r == ',' || r == ' ' }) {
			name := ""
			if n, err := strconv.Atoi(token); err == nil && n >= 1 && n <= len(candidates) {
				name = candidates[n-1].Plugin.Name
			} else if slices.ContainsFunc(candidates, func(c protocol.InitializeResult) bool { return c.Plugin.Name == token }) {
				name = token
			} else {
				return fmt.Errorf("%q is not in the list", token)
			}
			if !slices.Contains(names, name) {
				names = append(names, name)
			}
		}
		if len(names) > 1 && !Additive(level) {
			return fmt.Errorf("%s takes one plugin", level)
		}
		return nil
	})
	return names, err
}

// addMembers records plugins chosen for a level in each scenario.
func (k *asking) addMembers(scenarios []Scenario, level string, names []string) {
	role, _ := roleOf(level)
	for i := range scenarios {
		for _, name := range names {
			if info, ok := k.Candidates.Info(name); ok {
				scenarios[i].Members = append(scenarios[i].Members, pluginhost.Member{Info: info, Role: role})
			}
		}
	}
}

// consistent copies the scenarios an answer is compatible with, with the
// answer recorded.
func consistent(scenarios []Scenario, name, answer string) []Scenario {
	var out []Scenario
	for _, s := range scenarios {
		if current, set := s.Answers[name]; set && current != answer {
			continue
		}
		c := Scenario{Answers: map[string]string{}, Members: slices.Clone(s.Members)}
		for k, v := range s.Answers {
			c.Answers[k] = v
		}
		c.Answers[name] = answer
		out = append(out, c)
	}
	return out
}

// titled turns a choice's name into its question: "kind" asks "Kind".
func titled(name string) string {
	if name == "" {
		return ""
	}
	return strings.ToUpper(name[:1]) + strings.ReplaceAll(name[1:], "-", " ")
}
