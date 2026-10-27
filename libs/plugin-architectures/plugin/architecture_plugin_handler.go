package plugin

import (
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/check"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/imports"
	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Architecture is one architecture's convention plugin.
type Architecture struct {
	Version string
	rules   rules.Ruleset
	root    string
}

// Names lists the architectures this package serves, which are also the
// convention names ([conventions.hexagonal]).
func Names() []string { return rules.Names() }

// New returns the plugin of the architecture called name; version is
// reported in the handshake. It is nil for a name that is not served.
func New(name, version string) *Architecture {
	rs, ok := rules.For(name)
	if !ok {
		return nil
	}
	return &Architecture{Version: version, rules: rs}
}

// Describe implements serve.Plugin.
func (a *Architecture) Describe(params protocol.InitializeParams) (protocol.InitializeResult, error) {
	a.root = params.WorkspaceRoot
	return protocol.InitializeResult{
		Plugin:    protocol.Peer{Name: a.rules.Name, Version: a.Version},
		Kind:      protocol.KindConvention,
		Languages: languagesOf(a.rules),
	}, nil
}

// Languages are the languages the architecture called name declares in its
// handshake: the ones its ruleset lists, or every language the check can read.
// It is nil for a name that is not served.
func Languages(name string) []string {
	rs, ok := rules.For(name)
	if !ok {
		return nil
	}
	return languagesOf(rs)
}

// reads is the way to read a project's imports when the architecture applies
// to its language.
func (a *Architecture) reads(language string) (imports.Language, bool) {
	lang, ok := imports.For(language)
	if !ok {
		return imports.Language{}, false
	}
	if len(a.rules.Languages) > 0 && !slices.Contains(a.rules.Languages, lang.Name) {
		return imports.Language{}, false
	}
	return lang, true
}

func languagesOf(rs rules.Ruleset) []string {
	if len(rs.Languages) > 0 {
		return slices.Clone(rs.Languages)
	}
	return imports.Names()
}

// Contribute implements serve.Contributor: an architecture-check task for every
// project of a language it reads, run before lint.
func (a *Architecture) Contribute(params protocol.ContributeParams) (protocol.ContributeResult, error) {
	res := protocol.ContributeResult{Projects: map[string]protocol.Contribution{}}
	for _, p := range params.Projects {
		lang, ok := a.reads(p.Language)
		if !ok {
			continue
		}
		c := protocol.Contribution{Tasks: map[string]protocol.Task{TaskName: architectureCheck(a.rules, lang)}}
		if _, ok := p.Tasks["lint"]; ok {
			c.DependsOn = map[string][]string{"lint": {TaskName}}
		}
		res.Projects[p.Name] = c
	}
	return res, nil
}

// Doctor implements serve.DoctorContributor: a project with no folder named for
// a part of the architecture is checked by nothing, which is worth saying.
func (a *Architecture) Doctor(params protocol.DoctorParams) (protocol.DoctorResult, error) {
	var bare []string
	checked := 0
	for _, p := range params.Projects {
		if _, ok := a.reads(p.Language); !ok {
			continue
		}
		folders, err := check.LayerFolders(filepath.Join(a.root, filepath.FromSlash(p.Root)), p.Language, a.rules)
		if err != nil {
			continue
		}
		checked++
		if len(folders) == 0 {
			bare = append(bare, p.Name)
		}
	}
	if checked == 0 {
		return protocol.DoctorResult{Findings: []protocol.Finding{}}, nil
	}
	f := protocol.Finding{Check: "Code follows the " + a.rules.Name + " architecture (" + a.rules.Summary + ")", OK: len(bare) == 0}
	if !f.OK {
		f.Warning = true
		f.Detail = "no folder is named for a part of the architecture (" + strings.Join(a.rules.FolderNames(), ", ") + ") in " + named(bare) + ", so its rules check nothing there"
		f.Remedy = "move the code into folders with those names, or run smith architecture none to stop checking the layout"
	}
	return protocol.DoctorResult{Findings: []protocol.Finding{f}}, nil
}

// named lists a few project names and counts the rest.
func named(names []string) string {
	const shown = 4
	slices.Sort(names)
	if len(names) <= shown {
		return strings.Join(names, ", ")
	}
	return strings.Join(names[:shown], ", ") + " and " + strconv.Itoa(len(names)-shown) + " more"
}
