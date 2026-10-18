package recipewizard

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/pelletier/go-toml/v2"
)

var (
	// A recipe's name is its project's name and the start of its release tags
	// (team-recipe@1.2.0), so it is lowercase words joined by dashes.
	namePattern    = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	versionPattern = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

// ReadPlan reads an answers file for `smith recipe new --from-answers`: a
// Plan written as TOML, refused with a message when it has an unknown key, a
// name or version a release cannot use, or a level that is not filled in
// exactly one way.
func ReadPlan(data []byte, source string) (Plan, error) {
	var p Plan
	if err := toml.NewDecoder(bytes.NewReader(data)).DisallowUnknownFields().Decode(&p); err != nil {
		var strict *toml.StrictMissingError
		if errors.As(err, &strict) && len(strict.Errors) > 0 {
			var lines []string
			for _, e := range strict.Errors {
				row, col := e.Position()
				lines = append(lines, fmt.Sprintf("%s:%d:%d: unknown key %q; an answers file has name, version, extends, choices and levels", source, row, col, strings.Join(e.Key(), ".")))
			}
			return Plan{}, errors.New(strings.Join(lines, "\n"))
		}
		return Plan{}, fmt.Errorf("%s: %v", source, err)
	}
	if err := p.Check(); err != nil {
		return Plan{}, fmt.Errorf("%s: %w", source, err)
	}
	return p, nil
}

// Check says what is wrong with a plan before it is rendered.
func (p *Plan) Check() error {
	if !namePattern.MatchString(p.Name) {
		return fmt.Errorf("name %q must be lowercase words joined by dashes (team, team-node): it names the recipe's releases", p.Name)
	}
	if !versionPattern.MatchString(p.Version) {
		return fmt.Errorf("version %q must be major.minor.patch, like 0.1.0", p.Version)
	}
	for _, level := range OrderedLevels(p.Levels) {
		l := p.Levels[level]
		switch {
		case len(l.Fixed) > 0 && l.Choice != "":
			return fmt.Errorf("levels.%s: fixed and choice are alternatives; give one", level)
		case len(l.Fixed) == 0 && l.Choice == "":
			return fmt.Errorf("levels.%s: give fixed plugins, or a choice with the plugin for each answer", level)
		case len(l.Fixed) > 1 && !Additive(level):
			return fmt.Errorf("levels.%s: takes one plugin, not %s", level, strings.Join(l.Fixed, ", "))
		case l.Choice != "" && len(l.Plugins) == 0:
			return fmt.Errorf("levels.%s: choice %q needs plugins, by answer", level, l.Choice)
		case l.Choice == "" && len(l.Plugins) > 0:
			return fmt.Errorf("levels.%s: plugins go with a choice", level)
		}
		for answer, v := range l.Plugins {
			names, ok := pluginList(v)
			if !ok {
				return fmt.Errorf("levels.%s.plugins.%s: give a plugin name, or a list of names", level, answer)
			}
			if len(names) > 1 && !Additive(level) {
				return fmt.Errorf("levels.%s.plugins.%s: %s takes one plugin, not %s", level, answer, level, strings.Join(names, ", "))
			}
		}
		if declared, ok := p.Choices[l.Choice]; ok && l.Choice != "" {
			for answer := range l.Plugins {
				if !contains(declared.Options, answer) {
					return fmt.Errorf("levels.%s.plugins.%s: %q is not an option of choice %q (%s)", level, answer, answer, l.Choice, strings.Join(declared.Options, ", "))
				}
			}
		}
	}
	for name, c := range p.Choices {
		if len(c.Options) < 2 {
			return fmt.Errorf("choices.%s: a choice needs at least two options", name)
		}
		if c.Default != "" && !contains(c.Options, c.Default) {
			return fmt.Errorf("choices.%s: the default %q is not one of the options", name, c.Default)
		}
	}
	return nil
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// PluginsOf returns the plugins a level uses for an answer; the answer is
// ignored for fixed plugins.
func (l LevelPlan) PluginsOf(answer string) []string {
	if l.Choice == "" {
		return l.Fixed
	}
	names, _ := pluginList(l.Plugins[answer])
	return names
}

// pluginList reads a plugin name or a list of names, as TOML decodes them
// (a string, or a list of any) or as the wizard builds them (a list of
// strings).
func pluginList(v any) ([]string, bool) {
	if names, ok := v.([]string); ok {
		return names, len(names) > 0
	}
	return recipe.UseList(v)
}
