package recipewizard

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/recipe"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// fakePlugins describe a small set of plugins like the handshakes of the
// real ones: two languages, a linter and a test runner for each, a structure
// plugin and a deploy plugin.
func fakePlugins() map[string]protocol.InitializeResult {
	language := func(name string, languages ...string) protocol.InitializeResult {
		return protocol.InitializeResult{Plugin: protocol.Peer{Name: name, Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleLanguage, protocol.RoleToolchain, protocol.RoleScaffold}, Languages: languages, Capabilities: protocol.Capabilities{Generate: true}}
	}
	tool := func(name, slot, lang string) protocol.InitializeResult {
		return protocol.InitializeResult{Plugin: protocol.Peer{Name: name, Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleToolchain}, Provides: []string{slot}, Requires: []protocol.PluginRef{protocol.RequiresLanguage(lang)}}
	}
	return map[string]protocol.InitializeResult{
		"node":   language("node", "node", "javascript"),
		"python": language("python", "python"),
		"eslint": tool("eslint", "lint", "node"),
		"jest":   tool("jest", "test", "node"),
		"vitest": tool("vitest", "test", "node"),
		"ruff":   tool("ruff", "lint", "python"),
		"pytest": tool("pytest", "test", "python"),
		"docker": {Plugin: protocol.Peer{Name: "docker", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleDeploy}},
		"slices": {Plugin: protocol.Peer{Name: "slices", Version: "1.0.0"}, Roles: []protocol.Role{protocol.RoleStructure}},
	}
}

func fakeLookup(name string) (protocol.InitializeResult, error) {
	if info, ok := fakePlugins()[name]; ok {
		return info, nil
	}
	return protocol.InitializeResult{}, fmt.Errorf("%w: %q", recipe.ErrUnknownPlugin, name)
}

func fakeCandidates() *Candidates {
	var names []string
	for name := range fakePlugins() {
		names = append(names, name)
	}
	c, _ := NewCandidates(names, fakeLookup)
	return c
}

func scenarioWith(t *testing.T, names ...string) Scenario {
	t.Helper()
	var s Scenario
	for _, name := range names {
		info, ok := fakePlugins()[name]
		if !ok {
			t.Fatalf("no fake plugin %s", name)
		}
		role := protocol.RoleToolchain
		if slicesHas(info.Roles, protocol.RoleLanguage) {
			role = protocol.RoleLanguage
		}
		s.Members = append(s.Members, pluginhost.Member{Info: info, Role: role})
	}
	return s
}

func slicesHas(roles []protocol.Role, r protocol.Role) bool {
	for _, x := range roles {
		if x == r {
			return true
		}
	}
	return false
}

func names(infos []protocol.InitializeResult) string {
	var out []string
	for _, i := range infos {
		out = append(out, i.Plugin.Name)
	}
	return strings.Join(out, ",")
}

func TestOnlyPluginsThatFitTheLanguageAreOffered(t *testing.T) {
	c := fakeCandidates()
	if got := names(c.For("lint", []Scenario{scenarioWith(t, "python")})); got != "ruff" {
		t.Errorf("python lint candidates = %q; a toolchain that requires node must not be offered next to python", got)
	}
	if got := names(c.For("lint", []Scenario{scenarioWith(t, "node")})); got != "eslint" {
		t.Errorf("node lint candidates = %q", got)
	}
	if got := names(c.For("test", []Scenario{scenarioWith(t, "node")})); got != "jest,vitest" {
		t.Errorf("node test candidates = %q", got)
	}
	// To be fixed, a plugin must fit every branch of the choices.
	if got := names(c.For("lint", []Scenario{scenarioWith(t, "python"), scenarioWith(t, "node")})); got != "" {
		t.Errorf("a linter fitting only one language is offered as fixed: %q", got)
	}
	if got := names(c.For("language", nil)); got != "node,python" {
		t.Errorf("language candidates = %q", got)
	}
	if got := c.Levels(); strings.Join(got, ",") != "language,scaffold,lint,test,structure,deploy" {
		t.Errorf("levels = %v", got)
	}
}

func TestTheWizardBuildsAPlanFromTheAnswers(t *testing.T) {
	script := strings.Join([]string{
		"team", "", "", // name, version (0.1.0), no parent
		"c", "", "", "python, node", "python", // language: a choice named language
		"2", "1", // python, then node
		"",                // scaffold: unset
		"c", "", "1", "1", // lint: the same choice; ruff for python, eslint for node
		"c", "", "1", "1", // test: pytest for python, jest for node
		"f", "1", // structure: slices, fixed
		"", // deploy: unset
	}, "\n") + "\n"
	var out bytes.Buffer
	plan, err := Asker{In: strings.NewReader(script), Out: &out, Candidates: fakeCandidates(), Version: "0.1.0"}.Ask()
	if err != nil {
		t.Fatalf("%v\n%s", err, out.String())
	}
	if plan.Name != "team" || plan.Version != "0.1.0" || plan.Extends != "" {
		t.Errorf("plan = %+v", plan)
	}
	if c := plan.Choices["language"]; strings.Join(c.Options, ",") != "python,node" || c.Default != "python" || c.Ask != "Language" {
		t.Errorf("choice = %+v", c)
	}
	for level, want := range map[string]map[string]string{
		"language": {"python": "python", "node": "node"},
		"lint":     {"python": "ruff", "node": "eslint"},
		"test":     {"python": "pytest", "node": "jest"},
	} {
		l := plan.Levels[level]
		if l.Choice != "language" {
			t.Errorf("%s: choice = %q", level, l.Choice)
		}
		for answer, plugin := range want {
			if got := l.PluginsOf(answer); len(got) != 1 || got[0] != plugin {
				t.Errorf("%s when %s = %v, want %s\n%s", level, answer, got, plugin, out.String())
			}
		}
	}
	if got := plan.Levels["structure"].Fixed; len(got) != 1 || got[0] != "slices" {
		t.Errorf("structure = %v", got)
	}
	if _, ok := plan.Levels["scaffold"]; ok {
		t.Errorf("scaffold was left unset, but the plan has %+v", plan.Levels["scaffold"])
	}
	if _, ok := plan.Levels["deploy"]; ok {
		t.Error("deploy was left unset")
	}
	// The question for the python branch lists ruff only.
	listing := out.String()
	python := listing[strings.Index(listing, "when language=python:\n    1  ruff"):]
	if !strings.Contains(python, "1  ruff") || strings.Contains(python[:strings.Index(python, "when language=node")], "eslint") {
		t.Errorf("the python lint question must offer ruff and not eslint:\n%s", listing)
	}
}

func TestTheWizardRefusesWhatItCannotUse(t *testing.T) {
	// A bad name, then a good one; a choice with one answer, then two.
	script := strings.Join([]string{"Team", "team", "", "", "c", "", "", "python", "python, node", "", "1", "1"}, "\n") + "\n"
	var out bytes.Buffer
	_, err := Asker{In: strings.NewReader(script), Out: &out, Candidates: fakeCandidates(), Version: "0.1.0"}.Ask()
	if err != nil {
		t.Logf("%v", err)
	}
	for _, want := range []string{"must be lowercase words joined by dashes", "give at least two answers"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("the wizard did not say %q:\n%s", want, out.String())
		}
	}
}

func TestTheWizardStopsWhenTheInputEnds(t *testing.T) {
	_, err := Asker{In: strings.NewReader("team\n"), Out: &bytes.Buffer{}, Candidates: fakeCandidates(), Version: "0.1.0"}.Ask()
	if err == nil || !strings.Contains(err.Error(), "--from-answers") {
		t.Errorf("err = %v; it should point to --from-answers", err)
	}
}
