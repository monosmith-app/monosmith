package architecture

import (
	"fmt"
	"strings"
)

// None is the choice of no architecture: smith checks no code layout.
const None = "none"

// Architecture is one choice a workspace can make.
type Architecture struct {
	// Name is what the person picks: "vertical-slices".
	Name string
	// Meaning is one line saying what the choice is, in plain words.
	Meaning string
	// Table is the smith.toml table that turns it on ("conventions.vertical-slices");
	// empty when the choice writes nothing.
	Table string
}

// Writes says in plain words what choosing it changes in smith.toml.
func (a Architecture) Writes() string {
	if a.Table == "" {
		return "removes the lines that turn the other choices on"
	}
	return "adds [" + a.Table + "] to smith.toml, and removes the other choices' lines"
}

// catalog is every architecture smith knows, in the order they are offered.
// Adding one (DDD, MVC, hexagonal...) is adding a row here and the plugin that
// checks it; nothing else lists the architectures.
var catalog = []Architecture{
	{
		Name:    "vertical-slices",
		Meaning: "code grouped by what it does: one flat folder per outcome, files named for their role",
		Table:   "conventions.vertical-slices",
	},
	{
		Name:    None,
		Meaning: "smith checks no code layout",
	},
}

// Catalog lists the known architectures, in the order they are offered.
func Catalog() []Architecture { return append([]Architecture(nil), catalog...) }

// Names lists their names.
func Names() []string {
	names := make([]string, len(catalog))
	for i, a := range catalog {
		names[i] = a.Name
	}
	return names
}

// Find is the architecture called name.
func Find(name string) (Architecture, error) {
	for _, a := range catalog {
		if a.Name == name {
			return a, nil
		}
	}
	return Architecture{}, fmt.Errorf("no architecture called %q; the choices are %s", name, strings.Join(Names(), ", "))
}
