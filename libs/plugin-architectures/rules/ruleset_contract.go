package rules

import "strings"

// Layer is one part of an architecture, recognised by the folder names it
// goes by in code.
type Layer struct {
	// Name is how reports call it: "domain".
	Name string
	// Folders are the folder names that mean this layer, in any spelling:
	// they are compared without case, dashes, underscores or dots, so
	// "use-cases", "use_cases" and "UseCases" are one name.
	Folders []string
	// Rank orders the layers from the inside out. A layer may depend on
	// layers of its own or a lower rank, never on a higher one.
	Rank int
	// Meaning says in plain words what belongs in it.
	Meaning string
}

// Ruleset is one architecture's layout rules.
type Ruleset struct {
	// Name is the architecture's name, also its convention name.
	Name string
	// Layers are its parts, from the inside out.
	Layers []Layer
	// Contexts means the layers repeat inside each context (a bounded
	// context folder holds its own domain, application and infrastructure):
	// code may reach another context only through its root, not into its
	// layers.
	Contexts bool
	// Modules means the code is split into modules under a modules/ folder:
	// another module may be reached only through the files directly in the
	// module's folder or its api/ folder.
	Modules bool
	// IsolatedLayer names a layer whose sub-folders must not depend on each
	// other (the adapters of a hexagon); empty when none.
	IsolatedLayer string
	// Summary is one line of the rules, for the doctor and the reports.
	Summary string
}

// Normalise spells a folder name the way the rules compare it: lowercase,
// without dashes, underscores or dots.
func Normalise(name string) string {
	r := strings.NewReplacer("-", "", "_", "", ".", "")
	return strings.ToLower(r.Replace(name))
}

// LayerOf is the layer a folder name means, if it means one.
func (rs Ruleset) LayerOf(folder string) (Layer, bool) {
	n := Normalise(folder)
	for _, l := range rs.Layers {
		for _, f := range l.Folders {
			if Normalise(f) == n {
				return l, true
			}
		}
	}
	return Layer{}, false
}

// LayerByName finds a layer by its name.
func (rs Ruleset) LayerByName(name string) (Layer, bool) {
	for _, l := range rs.Layers {
		if l.Name == name {
			return l, true
		}
	}
	return Layer{}, false
}

// FolderNames lists the first folder name of each layer, for messages.
func (rs Ruleset) FolderNames() []string {
	var out []string
	for _, l := range rs.Layers {
		out = append(out, l.Folders[0])
	}
	return out
}
