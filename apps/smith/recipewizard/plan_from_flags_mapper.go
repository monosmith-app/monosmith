package recipewizard

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
)

// PlanFromFlags builds the plan that `smith recipe new --name ... --fixed
// role=plugin` describes, so a recipe whose plugins are all fixed needs
// neither questions nor an answers file. Each fixed value is a level, "=" and
// one plugin, or for the additive levels (structure, deploy, toolchain) plugins
// joined by commas. An empty version is 0.1.0.
//
// A level the recipe's users choose is described with --chosen, as
// level=question:answer=plugin,answer=plugin, for example
// language=Which language?:python=python,csharp=dotnet. The choice takes the
// level's name, its options are the answers in order (the first is the
// default), and an additive level (structure, deploy, toolchain) joins several
// plugins of one answer with "+" (full=slices+layers). The question cannot
// contain ":".
//
// It returns an error when a fixed or chosen value is malformed, names a level
// twice, or when the plan is not one a recipe can be made from (see Plan.Check).
func PlanFromFlags(name, version, extends string, fixed, chosen []string) (Plan, error) {
	if version == "" {
		version = "0.1.0"
	}
	plan := Plan{Name: name, Version: version, Extends: extends, Levels: map[string]LevelPlan{}}
	for _, f := range fixed {
		level, value, ok := strings.Cut(f, "=")
		level = strings.TrimSpace(level)
		if !ok || level == "" || strings.TrimSpace(value) == "" {
			return Plan{}, fmt.Errorf("--fixed %q must be role=plugin, like language=python or structure=slices,layers", f)
		}
		if _, dup := plan.Levels[level]; dup {
			return Plan{}, fmt.Errorf("--fixed names the level %q twice", level)
		}
		var plugins []string
		for _, p := range strings.Split(value, ",") {
			if p = strings.TrimSpace(p); p != "" {
				plugins = append(plugins, p)
			}
		}
		plan.Levels[level] = LevelPlan{Fixed: plugins}
	}
	for _, c := range chosen {
		level, choice, levelPlan, err := parseChosen(c)
		if err != nil {
			return Plan{}, err
		}
		if _, dup := plan.Levels[level]; dup {
			return Plan{}, fmt.Errorf("--chosen and --fixed name the level %q twice: a level is fixed or chosen", level)
		}
		if plan.Choices == nil {
			plan.Choices = map[string]recipe.Choice{}
		}
		plan.Choices[level] = choice
		plan.Levels[level] = levelPlan
	}
	if err := plan.Check(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

// parseChosen reads one --chosen value: level=question:answer=plugin,answer=plugin.
func parseChosen(value string) (level string, choice recipe.Choice, plan LevelPlan, err error) {
	usage := fmt.Errorf("--chosen %q must be level=question:answer=plugin,answer=plugin, like language=Which language?:python=python,csharp=dotnet", value)
	level, rest, ok := strings.Cut(value, "=")
	level = strings.TrimSpace(level)
	question, answers, hasAnswers := strings.Cut(rest, ":")
	if !ok || level == "" || !hasAnswers || strings.TrimSpace(question) == "" {
		return "", recipe.Choice{}, LevelPlan{}, usage
	}
	plan = LevelPlan{Choice: level, Plugins: map[string]any{}}
	choice = recipe.Choice{Ask: strings.TrimSpace(question)}
	for _, pair := range strings.Split(answers, ",") {
		answer, plugin, ok := strings.Cut(pair, "=")
		answer, plugin = strings.TrimSpace(answer), strings.TrimSpace(plugin)
		if !ok || answer == "" || plugin == "" {
			return "", recipe.Choice{}, LevelPlan{}, usage
		}
		if _, dup := plan.Plugins[answer]; dup {
			return "", recipe.Choice{}, LevelPlan{}, fmt.Errorf("--chosen %s names the answer %q twice", level, answer)
		}
		choice.Options = append(choice.Options, answer)
		if names := strings.Split(plugin, "+"); len(names) > 1 {
			plan.Plugins[answer] = names
		} else {
			plan.Plugins[answer] = plugin
		}
	}
	choice.Default = choice.Options[0]
	return level, choice, plan, nil
}
