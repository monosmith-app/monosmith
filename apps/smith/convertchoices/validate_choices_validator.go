package convertchoices

import (
	"fmt"
	"strings"
)

// Validate checks the values the flags gave. A choice flag needs the full
// level: a conservative conversion decides nothing about the old tooling.
func Validate(c Choices, g Given) error {
	if c.Level != LevelConservative && c.Level != LevelFull {
		return fmt.Errorf("--level %q is not one of conservative, full", c.Level)
	}
	if c.Level == LevelConservative && g.Any() {
		return fmt.Errorf("%s only apply with --level full, which decides what happens to the old tooling", flagsGiven(g))
	}
	for _, check := range []struct {
		flag, value string
		given       bool
		allowed     []string
	}{
		{"--lint", c.Lint, g.Lint, []string{Smith, Keep}},
		{"--test", c.Test, g.Test, []string{Jest, Vitest, Keep}},
		{"--build", c.Build, g.Build, []string{Smith, Keep}},
		{"--ci", c.CI, g.CI, []string{Smith, Keep}},
	} {
		if !check.given {
			continue
		}
		ok := false
		for _, a := range check.allowed {
			ok = ok || a == check.value
		}
		if !ok {
			return fmt.Errorf("%s %q is not one of %s", check.flag, check.value, strings.Join(check.allowed, ", "))
		}
	}
	return nil
}

func flagsGiven(g Given) string {
	var names []string
	for _, f := range []struct {
		name string
		on   bool
	}{{"--lint", g.Lint}, {"--test", g.Test}, {"--build", g.Build}, {"--ci", g.CI}} {
		if f.on {
			names = append(names, f.name)
		}
	}
	return strings.Join(names, ", ")
}
