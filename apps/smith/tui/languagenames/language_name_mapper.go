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
	"php":    "PHP",
	"swift":  "Swift",
	"bicep":  "Azure Bicep",
	"docker": "Docker",
	"recipe": "A recipe (a ready-made set of projects)",
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
