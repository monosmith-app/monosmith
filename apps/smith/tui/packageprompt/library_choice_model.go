package packageprompt

import (
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
)

// Library is a project of the workspace that the selected project may depend
// on: the dialog lists it with its kind and language, so nothing is typed.
type Library struct {
	Name, Kind, Language, Version string
}

// RangeChoice is one way to write the version a project asks of a library,
// with what it means in plain words. Spec is the --range value, empty to let
// smith follow the workspace.
type RangeChoice struct {
	Spec, Label, Meaning string
}

// RangeChoices are the real choices of the range for a library of a language,
// the default first. Only a language whose manifest has a version range for a
// workspace dependency has any; the others return nil and the dialog asks
// nothing more.
func RangeChoices(language string, lib Library) []RangeChoice {
	if language != "node" {
		return nil
	}
	choices := []RangeChoice{
		{Spec: "", Label: words.Say("packageprompt.range.follow.label"), Meaning: words.Say("packageprompt.range.follow.meaning")},
		{Spec: "*", Label: words.Say("packageprompt.range.any.label"), Meaning: words.Say("packageprompt.range.any.meaning")},
	}
	if lib.Version != "" {
		choices = append(choices, RangeChoice{Spec: "^" + lib.Version, Label: words.Say("packageprompt.range.current.label", "version", lib.Version), Meaning: words.Say("packageprompt.range.current.meaning")})
	}
	return choices
}

// LibraryMeaning says in plain words what linking the library does.
func LibraryMeaning(project string, lib Library) string {
	return words.Say("packageprompt.library.meaning", "project", project, "library", lib.Name)
}
