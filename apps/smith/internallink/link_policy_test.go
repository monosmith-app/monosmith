package internallink

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/projectgraph"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func project(name, root, language string, kind protocol.ProjectKind, exports []string, deps ...string) *projectgraph.Project {
	return &projectgraph.Project{
		Project: protocol.Project{Name: name, Root: root, Language: language, Kind: kind, Exports: exports},
		Plugin:  language,
		Deps:    deps,
	}
}

func labGraph(t *testing.T) *projectgraph.Graph {
	t.Helper()
	g, err := projectgraph.New([]*projectgraph.Project{
		project("shop-web", "apps/shop-web", "node", protocol.ProjectApp, []string{"shop-web"}, "ui"),
		project("ui", "libs/ui", "node", protocol.ProjectLib, []string{"@acme/ui"}, "core"),
		project("core", "libs/core", "node", protocol.ProjectLib, []string{"@acme/core"}),
		project("core-py", "libs/core-py", "python", protocol.ProjectLib, []string{"shop_core_py"}),
		project("api", "apps/api", "python", protocol.ProjectApp, []string{"api"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestMatchFindsAProjectByNameDirectoryOrExportedName(t *testing.T) {
	g := labGraph(t)
	cases := []struct{ name, want string }{
		{"core", "core"},
		{"libs/core", "core"},
		{"./libs/core/", "core"},
		{"@acme/core", "core"},
		{"shop_core_py", "core-py"},
		{"shop-core-py", "core-py"},
		{"Shop.Core.Py", "core-py"},
		{"lodash", ""},
		{"@acme/missing", ""},
		{"", ""},
	}
	for _, c := range cases {
		got := ""
		if p := Match(g, c.name); p != nil {
			got = p.Name
		}
		if got != c.want {
			t.Errorf("Match(%q) = %q, want %q", c.name, got, c.want)
		}
	}
}

func TestStripPrefix(t *testing.T) {
	if name, ok := StripPrefix("workspace:core"); name != "core" || !ok {
		t.Errorf("got %q %v", name, ok)
	}
	if name, ok := StripPrefix("core"); name != "core" || ok {
		t.Errorf("got %q %v", name, ok)
	}
}

func TestRefuseNamesTheProblemAndTheFix(t *testing.T) {
	g := labGraph(t)
	cases := []struct {
		name      string
		from, lib string
		want      []string
	}{
		{"itself", "ui", "ui", []string{"cannot depend on itself", "libs/ui"}},
		{"an app", "ui", "shop-web", []string{"shop-web is an app", "apps/shop-web", "smith kind"}},
		{"another language", "shop-web", "core-py", []string{"shop-web is node", "core-py is python"}},
		{"a cycle", "core", "ui", []string{"makes a cycle", "core → ui → core"}},
		{"a longer cycle", "core", "shop-web", []string{"is an app"}},
	}
	for _, c := range cases {
		err := Refuse(g, g.Get(c.from), g.Get(c.lib))
		if err == nil {
			t.Errorf("%s: no refusal", c.name)
			continue
		}
		for _, w := range c.want {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("%s: %q does not contain %q", c.name, err, w)
			}
		}
	}
}

func TestRefuseAllowsAValidLink(t *testing.T) {
	g := labGraph(t)
	if err := Refuse(g, g.Get("shop-web"), g.Get("core")); err != nil {
		t.Fatal(err)
	}
	if err := Refuse(g, g.Get("ui"), g.Get("core")); err != nil {
		t.Fatal(err)
	}
}

func TestRefuseNamesTheWholeCycle(t *testing.T) {
	g, err := projectgraph.New([]*projectgraph.Project{
		project("a", "libs/a", "go", protocol.ProjectLib, []string{"a"}, "b"),
		project("b", "libs/b", "go", protocol.ProjectLib, []string{"b"}, "c"),
		project("c", "libs/c", "go", protocol.ProjectLib, []string{"c"}),
	})
	if err != nil {
		t.Fatal(err)
	}
	got := Refuse(g, g.Get("c"), g.Get("a"))
	if got == nil || !strings.Contains(got.Error(), "c → a → b → c") {
		t.Fatalf("got %v", got)
	}
}

type fakeAsker struct {
	pick     int
	asked    int
	question string
	options  []Option
}

func (f *fakeAsker) Choose(q string, o []Option) (int, error) {
	f.asked++
	f.question, f.options = q, o
	if f.pick < 0 {
		return 0, errors.New("no answer")
	}
	return f.pick, nil
}

func lookupReturns(state RegistryState) Lookup {
	return func(*projectgraph.Project, string) RegistryState { return state }
}

func TestSplitSendsPlainNamesToTheRegistryAndProjectNamesToTheWorkspace(t *testing.T) {
	g := labGraph(t)
	var log strings.Builder
	got, err := Split(g, g.Get("shop-web"), []string{"zod", "@acme/core", "lodash"}, SplitOptions{Lookup: lookupReturns(RegistryMissing), Log: &log})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Libraries) != 1 || got.Libraries[0].Name != "core" || strings.Join(got.Packages, ",") != "zod,lodash" {
		t.Fatalf("got %+v", got)
	}
	if !strings.Contains(log.String(), "no package of that name") || !strings.Contains(log.String(), "--registry") {
		t.Errorf("the choice is not told: %q", log.String())
	}
}

func TestSplitWorkspacePrefixAlwaysLinksAndRegistryFlagAlwaysInstalls(t *testing.T) {
	g := labGraph(t)
	asker := &fakeAsker{}
	got, err := Split(g, g.Get("shop-web"), []string{"workspace:core"}, SplitOptions{Lookup: lookupReturns(RegistryHas), Ask: asker})
	if err != nil || len(got.Libraries) != 1 || asker.asked != 0 {
		t.Fatalf("got %+v %v asked %d", got, err, asker.asked)
	}
	got, err = Split(g, g.Get("shop-web"), []string{"core", "@acme/ui"}, SplitOptions{Registry: true, Ask: asker})
	if err != nil || len(got.Libraries) != 0 || len(got.Packages) != 2 || asker.asked != 0 {
		t.Fatalf("got %+v %v asked %d", got, err, asker.asked)
	}
	if _, err = Split(g, g.Get("shop-web"), []string{"workspace:core"}, SplitOptions{Registry: true}); err == nil {
		t.Fatal("workspace: with --registry must be refused")
	}
}

func TestSplitWorkspacePrefixOnAnUnknownNameSuggests(t *testing.T) {
	g := labGraph(t)
	_, err := Split(g, g.Get("shop-web"), []string{"workspace:cor"}, SplitOptions{})
	if err == nil || !strings.Contains(err.Error(), "Did you mean core") {
		t.Fatalf("got %v", err)
	}
}

func TestSplitAsksWhenTheRegistryHasThePackageOrCannotBeAsked(t *testing.T) {
	g := labGraph(t)
	for _, state := range []RegistryState{RegistryHas, RegistryUnknown} {
		for _, pick := range []int{0, 1} {
			asker := &fakeAsker{pick: pick}
			got, err := Split(g, g.Get("shop-web"), []string{"core"}, SplitOptions{Lookup: lookupReturns(state), Ask: asker})
			if err != nil || asker.asked != 1 || len(asker.options) != 2 {
				t.Fatalf("state %v: %+v %v asked %d", state, got, err, asker.asked)
			}
			if (pick == 0) != (len(got.Libraries) == 1) || (pick == 1) != (len(got.Packages) == 1) {
				t.Errorf("state %v pick %d: %+v", state, pick, got)
			}
			if asker.options[0].Meaning == "" || asker.options[1].Meaning == "" {
				t.Error("each option must say what it does")
			}
		}
	}
}

func TestSplitWithoutATerminalRefusesAndNamesBothFixes(t *testing.T) {
	g := labGraph(t)
	for _, state := range []RegistryState{RegistryHas, RegistryUnknown} {
		_, err := Split(g, g.Get("shop-web"), []string{"core"}, SplitOptions{Lookup: lookupReturns(state)})
		if err == nil {
			t.Fatalf("state %v: no refusal", state)
		}
		for _, w := range []string{"workspace:core", "--registry core", "libs/core"} {
			if !strings.Contains(err.Error(), w) {
				t.Errorf("state %v: %q does not contain %q", state, err, w)
			}
		}
	}
}

func TestSplitPassesAnAskerFailureOn(t *testing.T) {
	g := labGraph(t)
	_, err := Split(g, g.Get("shop-web"), []string{"core"}, SplitOptions{Lookup: lookupReturns(RegistryHas), Ask: &fakeAsker{pick: -1}})
	if err == nil {
		t.Fatal("an unanswered question must fail")
	}
}

func TestTerminalAskerTakesANumberOrTheDefaultAndAsksAgainOnJunk(t *testing.T) {
	opts := []Option{{Label: "one", Meaning: "first"}, {Label: "two"}}
	cases := []struct {
		in   string
		want int
	}{{"\n", 0}, {"2\n", 1}, {"x\n9\n2\n", 1}}
	for _, c := range cases {
		var out strings.Builder
		got, err := NewTerminalAsker(strings.NewReader(c.in), &out).Choose("Q?", opts)
		if err != nil || got != c.want {
			t.Errorf("%q: got %d %v", c.in, got, err)
		}
		if !strings.Contains(out.String(), "one (default)") || !strings.Contains(out.String(), "first") {
			t.Errorf("the default and the meaning must be shown: %q", out.String())
		}
	}
	if _, err := NewTerminalAsker(strings.NewReader(""), &strings.Builder{}).Choose("Q?", opts); err == nil {
		t.Error("end of input must fail")
	}
}
