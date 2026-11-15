package kindchange

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Related is another project the one changing is connected to.
type Related struct {
	Name     string
	Kind     protocol.ProjectKind
	Language string
}

// Subject is everything the review of a kind change needs to know.
type Subject struct {
	Name     string
	Language string
	From, To protocol.ProjectKind
	// Root and NewRoot are the folder now and after the change.
	Root, NewRoot string
	// TargetExists says something is already at NewRoot.
	TargetExists bool
	// Dependents import the project; Dependencies are imported by it.
	Dependents   []Related
	Dependencies []Related
	// LatestTag is the newest release tag of the project, empty when it was
	// never released.
	LatestTag string
	// HasReleaseSettings says smith.toml has [projects.<name>.release].
	HasReleaseSettings bool
	// HasDockerfile says the project's folder holds a Dockerfile.
	HasDockerfile bool
}

// Refusal is a reason not to go ahead. A forceable one is a judgement the
// person may overrule with --force; the others cannot be done at all.
type Refusal struct {
	Reason    string
	Forceable bool
}

// Review checks a kind change before anything is asked of the plugin: what
// makes it impossible or would break something, and what the person should
// know even if it goes ahead.
func Review(s Subject) (refusals []Refusal, warnings []string) {
	if s.From == s.To {
		return []Refusal{{Reason: fmt.Sprintf("%s is already %s", s.Name, article(s.To))}}, nil
	}
	if s.TargetExists {
		refusals = append(refusals, Refusal{Reason: s.NewRoot + " already exists, so " + s.Name + " cannot move there"})
	}
	if s.To == protocol.ProjectApp && len(s.Dependents) > 0 {
		refusals = append(refusals, Refusal{
			Reason:    fmt.Sprintf("an app is run, not imported, and %s %s it", names(s.Dependents), plural(s.Dependents, "imports", "import")),
			Forceable: true,
		})
	}
	if s.From == protocol.ProjectPackage && s.To != protocol.ProjectPackage {
		if published := ofKind(s.Dependents, protocol.ProjectPackage); len(published) > 0 {
			refusals = append(refusals, Refusal{
				Reason:    fmt.Sprintf("%s that import%s %s: once %s is not published, installing %s would fail", publishedPhrase(published), plural(published, "s", ""), s.Name, s.Name, plural(published, "it", "them")),
				Forceable: true,
			})
		}
		if s.LatestTag != "" {
			refusals = append(refusals, Refusal{
				Reason:    fmt.Sprintf("%s was released (the newest tag is %s): its published versions stay where they are but would never be updated", s.Name, s.LatestTag),
				Forceable: true,
			})
		}
	}

	if s.To == protocol.ProjectPackage {
		if internal := notOfKind(s.Dependencies, protocol.ProjectPackage); len(internal) > 0 {
			warnings = append(warnings, fmt.Sprintf("%s imports %s, which %s not published: it ships inside %s", s.Name, names(internal), isAre(internal), s.Name))
		}
		warnings = append(warnings, fmt.Sprintf("the next smith release will publish %s: check its registry credentials and the version it starts from", s.Name))
	}
	if s.From == protocol.ProjectApp && s.HasDockerfile {
		warnings = append(warnings, fmt.Sprintf("%s has a Dockerfile; as %s nothing deploys it, so it stays in the folder unused", s.Name, article(s.To)))
	}
	if s.HasReleaseSettings && s.To != protocol.ProjectApp {
		warnings = append(warnings, fmt.Sprintf("smith.toml has [projects.%s.release], which only applies to an app: remove it", s.Name))
	}
	if other := otherLanguage(s.Dependents, s.Language); len(other) > 0 {
		warnings = append(warnings, fmt.Sprintf("%s depend%s on %s through its schemas, not its code, so the %s plugin cannot rewrite what they say about it: check them", names(other), plural(other, "s", ""), s.Name, s.Language))
	}
	return refusals, warnings
}

func article(k protocol.ProjectKind) string {
	switch k {
	case protocol.ProjectApp:
		return "an app"
	case protocol.ProjectLib:
		return "a lib"
	}
	return "a package"
}

func names(rs []Related) string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Name
	}
	return strings.Join(out, ", ")
}

// publishedPhrase is "web is a published package" or "web, api are published
// packages".
func publishedPhrase(rs []Related) string {
	if len(rs) == 1 {
		return names(rs) + " is a published package"
	}
	return names(rs) + " are published packages"
}

func isAre(rs []Related) string {
	if len(rs) == 1 {
		return "is"
	}
	return "are"
}

func plural(rs []Related, one, many string) string {
	if len(rs) == 1 {
		return one
	}
	return many
}

func ofKind(rs []Related, kind protocol.ProjectKind) []Related {
	var out []Related
	for _, r := range rs {
		if r.Kind == kind {
			out = append(out, r)
		}
	}
	return out
}

func notOfKind(rs []Related, kind protocol.ProjectKind) []Related {
	var out []Related
	for _, r := range rs {
		if r.Kind != kind {
			out = append(out, r)
		}
	}
	return out
}

func otherLanguage(rs []Related, language string) []Related {
	var out []Related
	for _, r := range rs {
		if r.Language != language {
			out = append(out, r)
		}
	}
	return out
}
