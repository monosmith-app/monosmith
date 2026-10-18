package recipedrift

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Report is one project checked against its recipe.
type Report struct {
	Project string
	// Recipe is the recipe's name as the project recorded it.
	Recipe string
	// From is the version the project recorded, To the recipe's now.
	From, To string
	// Ref is the recipe that was resolved, as smith.toml names it.
	Ref string
	// Changes say, in the recipe's terms, what the recipe changes for the
	// project: its plugins, their settings and the answers recorded.
	Changes []string
	// Added are the plugins the recipe now adds to the project's stack. The
	// files a new project gets from a plugin are starter files, which an
	// upgrade does not write into an existing project.
	Added []string
	// Problem says why the project could not be checked: the recipe cannot be
	// loaded, a choice has no answer, the plugins no longer work together.
	Problem string

	// What the project is built with now, set when the recipe resolved.
	Stack      protocol.Stack
	Settings   map[string]map[string]any
	Answers    map[string]string
	Resolution recipe.Resolution
}

// Drifted reports whether the project differs from its recipe: a newer (or
// different) version, or a plugin, setting or answer that is not what the
// recipe now says. A project that could not be checked has not drifted.
func (r Report) Drifted() bool {
	return r.Problem == "" && (r.From != r.To || len(r.Changes) > 0)
}

// Summary is the report in a sentence, for doctor and upgrade.
func (r Report) Summary() string {
	switch {
	case r.Problem != "":
		return fmt.Sprintf("%s cannot be checked against its recipe %s: %s", r.Project, r.Recipe, r.Problem)
	case !r.Drifted():
		return fmt.Sprintf("%s follows its recipe %s%s", r.Project, r.Recipe, versionSuffix(r.To))
	case r.From == r.To:
		return fmt.Sprintf("%s differs from its recipe %s%s: %s", r.Project, r.Recipe, versionSuffix(r.To), strings.Join(r.Changes, "; "))
	}
	head := fmt.Sprintf("%s is on %s%s; the recipe is %s", r.Project, r.Recipe, versionSuffix(r.From), orNone(r.To))
	if len(r.Changes) == 0 {
		return head + ", which changes nothing the project uses"
	}
	return head + ", which " + strings.Join(r.Changes, "; ")
}

func versionSuffix(v string) string {
	if v == "" {
		return ""
	}
	return " " + v
}

func orNone(v string) string {
	if v == "" {
		return "unversioned"
	}
	return v
}
