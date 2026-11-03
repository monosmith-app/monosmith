package toolinstall

import (
	"path/filepath"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Source is what one loaded plugin declares.
type Source struct {
	Plugin    string
	Languages []string
	Tools     []protocol.ToolNeed
}

// Use is a project that uses a language and the choices of its stack.
type Use struct {
	Project  string
	Language string
	// Root is the absolute folder of the project.
	Root    string
	Choices []string
}

// Needed lists the tools the workspace needs, once each, in the order the
// plugins declare them. A tool of a language plugin is needed when a project of
// the language exists, or the language is named in extraLanguages (smith add
// --lang python). A tool tied to a choice (azure-functions) is needed when a
// project's stack has the choice, or the choice is in extraChoices. A tool tied
// to files (gradle) is needed only by a project whose folder holds one of them,
// so a language named by flag does not bring it. The core
// tools come first and are always needed.
//
// Parameters: workspace is the absolute workspace root; core are the tools smith
// itself needs; sources are the loaded plugins; uses are the projects.
// Returns: the needs, each with the folders to look for a pin file.
func Needed(workspace string, core []protocol.ToolNeed, sources []Source, uses []Use, extraLanguages, extraChoices []string) []Need {
	var out []Need
	index := map[string]int{}
	add := func(t protocol.ToolNeed, u *Use) {
		i, seen := index[t.Name]
		if !seen {
			index[t.Name] = len(out)
			out = append(out, Need{ToolNeed: t, Dirs: []string{workspace}})
			i = len(out) - 1
		}
		if u != nil {
			if !slices.Contains(out[i].Projects, u.Project) {
				out[i].Projects = append(out[i].Projects, u.Project)
			}
			if u.Root != "" && !slices.Contains(out[i].Dirs, u.Root) {
				out[i].Dirs = append(out[i].Dirs, u.Root)
			}
		}
	}
	for _, t := range core {
		add(t, nil)
	}
	for _, src := range sources {
		for _, t := range src.Tools {
			matched := false
			for i := range uses {
				u := &uses[i]
				if !sourceHasLanguage(src, u.Language) {
					continue
				}
				if (t.Choice == "" || slices.Contains(u.Choices, t.Choice)) && projectHasFile(u.Root, t.Files) {
					add(t, u)
					matched = true
				}
			}
			if matched {
				continue
			}
			langExtra := slices.ContainsFunc(extraLanguages, func(l string) bool { return sourceHasLanguage(src, l) })
			switch {
			case t.Choice == "" && langExtra && len(t.Files) == 0, t.Choice != "" && slices.Contains(extraChoices, t.Choice):
				add(t, nil)
			}
		}
	}
	return out
}

// sourceHasLanguage reports whether a plugin answers to a language name.
func sourceHasLanguage(src Source, lang string) bool {
	lang = strings.ToLower(lang)
	return lang == strings.ToLower(src.Plugin) || slices.ContainsFunc(src.Languages, func(l string) bool { return strings.EqualFold(l, lang) })
}

// Lookup finds tools by name among everything the core and the plugins declare,
// whether or not a project needs them. The first declaration of a name wins.
//
// Returns: the tools in the order asked, and the names that nothing declares.
func Lookup(workspace string, core []protocol.ToolNeed, sources []Source, names []string) (found []Need, unknown []string) {
	for _, name := range names {
		var hit *protocol.ToolNeed
		for _, t := range core {
			if t.Name == name {
				hit = &t
				break
			}
		}
		for _, src := range sources {
			for _, t := range src.Tools {
				if hit == nil && t.Name == name {
					hit = &t
				}
			}
		}
		if hit == nil {
			unknown = append(unknown, name)
			continue
		}
		found = append(found, Need{ToolNeed: *hit, Dirs: []string{filepath.Clean(workspace)}})
	}
	return found, unknown
}
