package languagenames

import "strings"

var languages = map[string]string{
	"node":   "TypeScript / JavaScript",
	"python": "Python",
	"go":     "Go",
	"rust":   "Rust",
	"jvm":    "Java / Kotlin",
	"dotnet": "C# / .NET",
	"cpp":    "C / C++",
	"dart":   "Dart / Flutter",
	"ruby":   "Ruby",
	"objc":   "Objective-C",
	"php":    "PHP",
	"vba":    "VBA",
	"swift":  "Swift",
	"lua":    "Lua",
	"bicep":  "Azure Bicep",
	"bundle": "Bundle",
	"docker": "Docker",
	"recipe": "A recipe (a ready-made set of projects)",

	"applescript": "AppleScript",
}

// popular are the languages most people reach for, in the order the lists
// show them; the rest follow alphabetically.
var popular = []string{"node", "python", "go", "jvm", "dotnet", "rust"}

// Of is the language as people say it; an unknown id comes back as it is.
func Of(id string) string {
	if n, ok := languages[strings.ToLower(id)]; ok {
		return n
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
		return "app"
	case "lib":
		return "library"
	case "package":
		return "package"
	}
	return kind
}

var descriptions = map[string]string{
	"node":   "runs on Node, with tests and linting set up",
	"python": "scripts, services and libraries",
	"go":     "services and command-line tools",
	"rust":   "fast, safe programs and libraries",
	"jvm":    "built with Gradle or Maven",
	"dotnet": ".NET apps and libraries",
	"cpp":    "native programs and libraries",
	"dart":   "Flutter apps and Dart packages",
	"ruby":   "gems and web apps",
	"objc":   "CocoaPods libraries",
	"php":    "web apps and Composer packages",
	"vba":    "Office macros kept in git as text",
	"swift":  "packages and command-line tools built with SwiftPM",
	"lua":    "scripts and LuaRocks libraries",
	"bicep":  "Azure infrastructure as code",
	"bundle": "apps assembled from other projects: app.json and flow.json",
	"docker": "container images",
	"recipe": "A ready-made set of projects that work together",

	"applescript": "scripts that automate macOS apps",
}

// Describe is a one-line description of what a language gives a new project,
// or "" for a language it does not know.
func Describe(id string) string { return descriptions[strings.ToLower(id)] }

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
