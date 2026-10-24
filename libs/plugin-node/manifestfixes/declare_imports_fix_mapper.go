package manifestfixes

import (
	"fmt"
	"path"
	"sort"

	"github.com/monosmith-app/monosmith/libs/plugin-node/declaredimports"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// DeclareImportsDescription is the same for every project, so the host folds
// the fix of all of them into one.
const DeclareImportsDescription = "declare the packages each project imports in its package.json"

// need is one package a project imports without declaring it.
type need struct {
	name string
	// dev is true when only tests and tooling import it: devDependencies.
	dev bool
	// moveFrom is the spec it has in devDependencies now, when a shipped file
	// imports what only devDependencies declare.
	moveFrom string
	// specs are the ranges to choose from, the most used in the workspace first.
	specs []string
}

// DeclareImportsFix adds the packages a project imports to its package.json,
// each with the range the workspace already uses (the most used one; the
// installed version when no project declares it), and installs. When a
// package is declared with different ranges in the workspace the fix asks
// which to take (the first such package; any other takes the most used).
// A package no project declares and that is not installed has no known
// version and is left to the user. It returns nil when nothing can be done.
func DeclareImportsFix(workspaceRoot string, project protocol.Project, undeclared []declaredimports.UndeclaredImport) *protocol.Fix {
	manifest := path.Join(project.Root, "package.json")
	raw, err := readFile(workspaceRoot, manifest)
	if err != nil {
		return nil
	}
	current, err := jsondoc.ParseDocument(raw)
	if err != nil {
		return nil
	}
	known := declaredimports.WorkspaceSpecs(workspaceRoot)
	needs := map[string]*need{}
	for _, u := range undeclared {
		n := needs[u.Package]
		if n == nil {
			n = &need{name: u.Package, dev: true}
			needs[u.Package] = n
		}
		if !declaredimports.DevFile(u.File) {
			n.dev = false
		}
		if u.DevOnly {
			if dev := current.Object("devDependencies", false); dev != nil {
				n.moveFrom, _ = dev.String(u.Package)
			}
		}
	}
	var resolved []*need
	for _, name := range sortedNames(needs) {
		n := needs[name]
		switch {
		case n.moveFrom != "":
			n.specs = []string{n.moveFrom}
		case len(known[name]) > 0:
			for _, use := range known[name] {
				n.specs = append(n.specs, use.Spec)
			}
		default:
			if v := declaredimports.InstalledVersion(workspaceRoot, name); v != "" {
				n.specs = []string{"^" + v}
			}
		}
		if len(n.specs) > 0 {
			resolved = append(resolved, n)
		}
	}
	if len(resolved) == 0 {
		return nil
	}

	fix := &protocol.Fix{
		Description: DeclareImportsDescription,
		AppliesTo:   []string{project.Name},
		Steps:       []protocol.FixStep{{Description: "install what the package.json files now declare", Argv: []string{"npm", "install", "--no-audit", "--no-fund"}}},
	}
	pick := map[string]string{}
	var ask *need
	for _, n := range resolved {
		pick[n.name] = n.specs[0]
		if ask == nil && len(n.specs) > 1 {
			ask = n
		}
	}
	render := func() (protocol.FileEdit, bool) {
		doc, err := jsondoc.ParseDocument(raw)
		if err != nil {
			return protocol.FileEdit{}, false
		}
		for _, n := range resolved {
			declare(doc, n, pick[n.name])
		}
		return protocol.FileEdit{Path: manifest, Content: render(doc, raw)}, true
	}
	if ask == nil {
		file, ok := render()
		if !ok {
			return nil
		}
		fix.Files = []protocol.FileEdit{file}
		return fix
	}
	choice := protocol.FixChoice{Question: fmt.Sprintf("%s: which version of %s should package.json declare? The workspace uses several.", project.Name, ask.name)}
	for _, spec := range ask.specs {
		pick[ask.name] = spec
		file, ok := render()
		if !ok {
			return nil
		}
		label := ask.name + "@" + spec
		for _, use := range known[ask.name] {
			if use.Spec == spec {
				label += fmt.Sprintf(" (declared in %d package.json files)", use.Projects)
			}
		}
		choice.Options = append(choice.Options, protocol.FixChoiceOption{Label: label, Fix: protocol.Fix{Files: []protocol.FileEdit{file}}})
	}
	// A fix that asks is one per project; the host does not fold it.
	fix.Choices = []protocol.FixChoice{choice}
	return fix
}

// declare adds the package to dependencies (devDependencies for tooling only),
// keeping the section sorted when it was, and drops a devDependencies entry
// it moves.
func declare(doc *jsondoc.Document, n *need, spec string) {
	if n.moveFrom != "" {
		if dev := doc.Object("devDependencies", false); dev != nil {
			dev.Delete(n.name)
			if dev.Len() == 0 {
				doc.Delete("devDependencies")
			}
		}
	}
	section := "dependencies"
	if n.dev {
		section = "devDependencies"
	}
	deps := doc.Object(section, true)
	keys := deps.Keys()
	sorted := sort.StringsAreSorted(keys)
	deps.SetString(n.name, spec)
	if !sorted {
		return
	}
	// Rebuild in order: the document only appends new keys.
	all := map[string]string{}
	for _, k := range deps.Keys() {
		all[k], _ = deps.String(k)
	}
	names := deps.Keys()
	sort.Strings(names)
	for _, k := range names {
		deps.Delete(k)
	}
	for _, k := range names {
		deps.SetString(k, all[k])
	}
}

func sortedNames(m map[string]*need) []string {
	names := make([]string, 0, len(m))
	for name := range m {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
