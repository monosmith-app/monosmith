package doctorchecks

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/libs/words/words"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/jsondoc"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/projectnaming"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/tomltext"
)

// PackageNameFindings warns about each project whose package name does not
// start with the workspace namespace the way a new project's name would
// (@acme/core, acme-core, acme_core, Acme.Core). It is a warning, never a
// failure: a name set on purpose (smith add --package-name) is accepted in
// smith.toml with [[doctor.accept]]. For an npm, Composer, Cargo or Python
// project that no other project depends on, the finding carries the rename of
// the manifest as a fix, and the person confirms it. A project others depend
// on is renamed by hand, because every dependent names it too.
//
// It runs only when [workspace] namespace is set.
//
// Parameters: s is the open session.
// Returns: nothing without a namespace; otherwise one finding per project that does not follow the namespace, or one
// passed finding when every project does.
func PackageNameFindings(s *session.Session) []protocol.Finding {
	if strings.TrimSpace(s.Config.Workspace.Namespace) == "" {
		// The namespace is the workspace name until the person sets one, and a
		// repository smith adopted or converted keeps the names it had: the
		// check starts when the person declares the namespace.
		return nil
	}
	ws := s.Config.Workspace.Info("")
	check := words.Say("naming.doctor.check")
	var out []protocol.Finding
	for _, p := range s.Graph.Projects {
		eco, ok := projectnaming.EcosystemOf(p.Language)
		if !ok || p.PackageName == "" || projectnaming.Follows(ws, eco, p.PackageName) {
			continue
		}
		out = append(out, packageNameFinding(s, p.Name, p.Root, eco, p.PackageName, projectnaming.Name(ws, eco, p.Name), check))
	}
	if len(out) == 0 {
		return []protocol.Finding{{Check: check, OK: true, Area: AreaConfig}}
	}
	return out
}

// packageNameFinding is the warning for one project, with the rename fix when
// the manifest is one smith can edit and nothing depends on the package.
func packageNameFinding(s *session.Session, project, root string, eco projectnaming.Ecosystem, actual, expected, check string) protocol.Finding {
	f := protocol.Finding{
		Check: check, Project: project, Warning: true,
		Problem: words.Say("naming.doctor.problem", "package", actual, "namespace", projectnaming.Namespace(s.Config.Workspace.Info(""))),
		Remedy:  words.Say("naming.doctor.remedy.manual", "expected", expected),
	}
	if dependents := s.Graph.Dependents(project); len(dependents) > 0 {
		f.Remedy = words.Say("naming.doctor.remedy.dependents", "expected", expected, "dependents", strings.Join(dependents, ", "))
		return f
	}
	file, content, ok := renamedManifest(s.Root, root, eco, expected)
	if !ok {
		return f
	}
	f.Remedy = words.Say("naming.doctor.remedy.fix", "expected", expected)
	f.Fix = &protocol.Fix{
		Description: words.Say("naming.doctor.fix", "package", actual, "expected", expected),
		AppliesTo:   []string{project},
		Files:       []protocol.FileEdit{{Path: file, Content: content}},
	}
	return f
}

// renamedManifest is the manifest of the project at root with its package name
// set to name: package.json and composer.json (the "name" key), Cargo.toml
// ([package] name) and pyproject.toml ([project] name). It reports false for
// any other ecosystem, or when the file cannot be read.
func renamedManifest(workspaceRoot, root string, eco projectnaming.Ecosystem, name string) (string, string, bool) {
	var file string
	switch eco {
	case projectnaming.Npm:
		file = "package.json"
	case projectnaming.Composer:
		file = "composer.json"
	case projectnaming.Cargo:
		file = "Cargo.toml"
	case projectnaming.Python:
		file = "pyproject.toml"
	default:
		return "", "", false
	}
	rel := path.Join(root, file)
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(rel)))
	if err != nil {
		return "", "", false
	}
	switch eco {
	case projectnaming.Npm, projectnaming.Composer:
		doc, err := jsondoc.ParseDocument(raw)
		if err != nil {
			return "", "", false
		}
		doc.SetString("name", name)
		return rel, string(doc.Bytes()), true
	case projectnaming.Cargo:
		return rel, tomltext.SetString(string(raw), "package", "name", name), true
	default:
		return rel, tomltext.SetString(string(raw), "project", "name", name), true
	}
}
