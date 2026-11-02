package internallink

import (
	"fmt"
	"io"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
)

// SplitOptions control how Split reads the names.
type SplitOptions struct {
	// Registry sends every name to the registry (smith install --registry).
	Registry bool
	// Lookup asks the registry about a name that is also a project.
	Lookup Lookup
	// Ask puts the choice to the user; nil when there is no terminal, or the
	// user passed --yes (which never decides an ambiguous name).
	Ask Asker
	// Log receives the one-line notes.
	Log io.Writer
}

// Names are the names of one install, sorted into where they go.
type Names struct {
	// Libraries are the workspace projects to link, in the order given.
	Libraries []*projectgraph.Project
	// Packages are the names for the registry, as given.
	Packages []string
}

// Split decides, for each name given to smith install -w <from>, whether it is
// a library of this workspace or a package of the registry.
//
// A name that begins with workspace: is always the project. With
// SplitOptions.Registry every name is a package. Any other name that matches a
// project (by name, directory, or the identifier it exports) is the project
// only when the registry has no package of that name; when the registry has one,
// or cannot be asked, the user chooses, and with no terminal Split refuses and
// names both fixes. It never picks one of the two silently.
func Split(g *projectgraph.Graph, from *projectgraph.Project, given []string, o SplitOptions) (Names, error) {
	var out Names
	for _, name := range given {
		bare, explicit := StripPrefix(name)
		if explicit && o.Registry {
			return out, fmt.Errorf("%s asks for the project and --registry asks for the registry: use one of them", name)
		}
		if o.Registry {
			out.Packages = append(out.Packages, name)
			continue
		}
		project := Match(g, bare)
		if explicit {
			if project == nil {
				return out, fmt.Errorf("no project %q in this workspace.%s", bare, nearmatch.Suggestion(bare, g.Names()))
			}
			out.Libraries = append(out.Libraries, project)
			continue
		}
		if project == nil {
			out.Packages = append(out.Packages, name)
			continue
		}
		link, err := settle(from, name, project, o)
		if err != nil {
			return out, err
		}
		if link {
			out.Libraries = append(out.Libraries, project)
		} else {
			out.Packages = append(out.Packages, name)
		}
	}
	return out, nil
}

// settle decides a plain name that matches a project: true links the project.
func settle(from *projectgraph.Project, name string, project *projectgraph.Project, o SplitOptions) (bool, error) {
	state := RegistryUnknown
	if o.Lookup != nil {
		state = o.Lookup(from, name)
	}
	if state == RegistryMissing {
		if o.Log != nil {
			fmt.Fprintf(o.Log, "%s is a project of this workspace (%s) and the registry has no package of that name, so smith links the project. To install from the registry anyway: --registry\n", name, project.Root)
		}
		return true, nil
	}
	because := "a package on the registry"
	if state == RegistryUnknown {
		because = "maybe a package on the registry (smith could not ask it)"
	}
	if o.Ask == nil {
		return false, fmt.Errorf("%s is a project of this workspace (%s) and %s. Say which one you mean: smith install -w %s workspace:%s (the project), or smith install -w %s --registry %s (the registry package)", name, project.Root, because, from.Name, name, from.Name, name)
	}
	question := fmt.Sprintf("`%s` is a project of this workspace and %s. Link the project or install the package?", name, because)
	pick, err := o.Ask.Choose(question, []Option{
		{Label: "Link the workspace project " + project.Name, Meaning: from.Name + " will use the code in " + project.Root + ", always its current copy. Nothing is downloaded."},
		{Label: "Install `" + name + "` from the registry", Meaning: from.Name + " will use a published copy from outside this workspace, and the project " + project.Name + " stays unused."},
	})
	if err != nil {
		return false, err
	}
	return pick == 0, nil
}

// joinNames lists names for a message.
func joinNames(names []string) string { return strings.Join(names, ", ") }
