package languagenames

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// languages is the catalog key of the name of each language, by language id.
var languages = map[string]string{
	"node":   "languagenames.name.node",
	"python": "languagenames.name.python",
	"go":     "languagenames.name.go",
	"rust":   "languagenames.name.rust",
	"jvm":    "languagenames.name.jvm",
	"dotnet": "languagenames.name.dotnet",
	"cpp":    "languagenames.name.cpp",
	"dart":   "languagenames.name.dart",
	"ruby":   "languagenames.name.ruby",
	"objc":   "languagenames.name.objc",
	"php":    "languagenames.name.php",
	"vba":    "languagenames.name.vba",
	"swift":  "languagenames.name.swift",
	"lua":    "languagenames.name.lua",
	"bicep":  "languagenames.name.bicep",
	"bundle": "languagenames.name.bundle",
	"docker": "languagenames.name.docker",
	"recipe": "languagenames.name.recipe",

	"applescript": "languagenames.name.applescript",
}

// popular are the languages most people reach for, in the order the lists
// show them; the rest follow alphabetically.
var popular = []string{"node", "python", "go", "jvm", "dotnet", "rust"}

// Of is the language as people say it; an unknown id comes back as it is.
func Of(id string) string {
	if key, ok := languages[strings.ToLower(id)]; ok {
		return words.Say(key)
	}
	return id
}

// Order puts the popular languages first and the rest in the order given.
func Order(ids []string) []string {
	seen := map[string]bool{}
	var out []string
	has := map[string]bool{}
	for _, id := range ids {
		has[id] = true
	}
	for _, id := range popular {
		if has[id] {
			out = append(out, id)
			seen[id] = true
		}
	}
	for _, id := range ids {
		if !seen[id] {
			out = append(out, id)
		}
	}
	return out
}

// KindOf is a project kind in words: an app is deployed, a library is shared
// inside the workspace, a package is published.
func KindOf(kind string) string {
	switch kind {
	case "app":
		return words.Say("languagenames.kind.app")
	case "lib":
		return words.Say("languagenames.kind.lib")
	case "package":
		return words.Say("languagenames.kind.package")
	}
	return kind
}

// descriptions is the catalog key of the description of each language, by id.
var descriptions = map[string]string{
	"node":   "languagenames.about.node",
	"python": "languagenames.about.python",
	"go":     "languagenames.about.go",
	"rust":   "languagenames.about.rust",
	"jvm":    "languagenames.about.jvm",
	"dotnet": "languagenames.about.dotnet",
	"cpp":    "languagenames.about.cpp",
	"dart":   "languagenames.about.dart",
	"ruby":   "languagenames.about.ruby",
	"objc":   "languagenames.about.objc",
	"php":    "languagenames.about.php",
	"vba":    "languagenames.about.vba",
	"swift":  "languagenames.about.swift",
	"lua":    "languagenames.about.lua",
	"bicep":  "languagenames.about.bicep",
	"bundle": "languagenames.about.bundle",
	"docker": "languagenames.about.docker",
	"recipe": "languagenames.about.recipe",

	"applescript": "languagenames.about.applescript",
}

// Describe is a one-line description of what a language gives a new project,
// or "" for a language it does not know.
func Describe(id string) string {
	if key, ok := descriptions[strings.ToLower(id)]; ok {
		return words.Say(key)
	}
	return ""
}

// MostUsed is the language that most of the given projects use, for the
// default of a new project; a tie goes to the one seen first, and "" means
// there is none.
func MostUsed(used []string) string {
	count := map[string]int{}
	var seen []string
	for _, id := range used {
		if id == "" {
			continue
		}
		if count[id] == 0 {
			seen = append(seen, id)
		}
		count[id]++
	}
	best := ""
	for _, id := range seen {
		if best == "" || count[id] > count[best] {
			best = id
		}
	}
	return best
}
