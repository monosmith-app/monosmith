package recipewizard

import (
	"fmt"
	"strings"
)

// PlanFromFlags builds the plan that `smith recipe new --name ... --fixed
// role=plugin` describes, so a recipe whose plugins are all fixed needs
// neither questions nor an answers file. Each fixed value is a level, "=" and
// one plugin, or for the additive levels (structure, deploy, toolchain) plugins
// joined by commas. An empty version is 0.1.0.
//
// It returns an error when a fixed value is not role=plugin, names a level
// twice, or when the plan is not one a recipe can be made from (see Plan.Check).
func PlanFromFlags(name, version, extends string, fixed []string) (Plan, error) {
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
	if err := plan.Check(); err != nil {
		return Plan{}, err
	}
	return plan, nil
}
