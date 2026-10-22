package sbom

import (
	"regexp"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/components"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var lockSpec = regexp.MustCompile(`^    ([^\s(]+) \(([^)]+)\)$`)

// ComponentsFromLock lists the gems a Gemfile.lock resolved: the specs of its
// GEM, GIT and GITHUB sections. PATH gems are the repository's own.
func ComponentsFromLock(lock string) []protocol.Component {
	var out []protocol.Component
	section := ""
	for _, line := range strings.Split(strings.ReplaceAll(lock, "\r\n", "\n"), "\n") {
		if line != "" && !strings.HasPrefix(line, " ") {
			section = strings.TrimSpace(line)
			continue
		}
		if section != "GEM" && section != "GIT" && section != "GITHUB" {
			continue
		}
		if m := lockSpec.FindStringSubmatch(line); m != nil {
			out = append(out, components.New("gem", "", m[1], m[2]))
		}
	}
	return components.Dedupe(out)
}

var declared = regexp.MustCompile(`^\s*(?:gem|\w+\.add(?:_runtime)?_dependency)\s*\(?\s*["']([^"']+)["']\s*(?:,\s*["']([^"']+)["'])?`)

// ComponentsFromManifests lists the gems a Gemfile or gemspec declares, by
// the first version requirement each gives.
func ComponentsFromManifests(texts ...string) []protocol.Component {
	var out []protocol.Component
	for _, text := range texts {
		for _, line := range strings.Split(text, "\n") {
			if m := declared.FindStringSubmatch(line); m != nil {
				out = append(out, components.New("gem", "", m[1], m[2]))
			}
		}
	}
	return components.Dedupe(out)
}
