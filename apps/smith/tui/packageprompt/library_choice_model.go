package packageprompt

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
		{Spec: "", Label: "Follow the workspace (default)", Meaning: "Use the range the workspace already uses for its own libraries. When it has none yet, smith writes * and says so."},
		{Spec: "*", Label: "Any version (*)", Meaning: "npm uses the library of this workspace whatever its version. You never edit the line when the library changes."},
	}
	if lib.Version != "" {
		choices = append(choices, RangeChoice{Spec: "^" + lib.Version, Label: "The current version (^" + lib.Version + ")", Meaning: "npm uses the workspace copy while its version fits, and the registry copy when it does not. Edit the line on a new major version."})
	}
	return choices
}

// LibraryMeaning says in plain words what linking the library does.
func LibraryMeaning(project string, lib Library) string {
	return project + " will use the code of " + lib.Name + " in this workspace, always its current copy. Nothing is downloaded."
}
