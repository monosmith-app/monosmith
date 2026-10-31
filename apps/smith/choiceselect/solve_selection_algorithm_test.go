package choiceselect

import (
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func offer(plugin, axis, name string, mutate func(*Offer)) Offer {
	o := Offer{Choice: protocol.Choice{Axis: axis, Name: name, Language: "node"}, Plugin: plugin}
	if mutate != nil {
		mutate(&o)
	}
	return o
}

// webOffers is a small catalog shaped like the web choices.
func webOffers() []Offer {
	both := []protocol.Role{protocol.RoleScaffold, protocol.RoleToolchain}
	return []Offer{
		offer("vite", "framework", "react", func(o *Offer) { o.Roles, o.Needs = both, []string{"lint", "test"} }),
		offer("vite", "framework", "vue", func(o *Offer) { o.Roles, o.Needs = both, []string{"lint", "test"} }),
		offer("eslint", "lint", "eslint", func(o *Offer) { o.Roles, o.Default = []protocol.Role{protocol.RoleToolchain}, true }),
		offer("vitest", "test", "vitest", func(o *Offer) { o.Roles, o.Default = []protocol.Role{protocol.RoleToolchain}, true }),
		offer("jest", "test", "jest", func(o *Offer) { o.Roles = []protocol.Role{protocol.RoleToolchain} }),
		offer("vite", "style", "bootstrap", nil),
		offer("vite", "style", "tailwind", nil),
		offer("vite", "style", "chakra", func(o *Offer) { o.Requires = []string{"react"} }),
		offer("vite", "library", "p5", func(o *Offer) { o.Many, o.Requires = true, []string{"react"} }),
		offer("vite", "library", "d3", func(o *Offer) { o.Many = true }),
		offer("pip", "framework", "flask", func(o *Offer) { o.Language = "python"; o.Roles = both }),
		offer("kit", "framework", "kit", func(o *Offer) { o.Roles = both }),
		offer("vite", "style", "sass", func(o *Offer) { o.Conflicts = []string{"bootstrap"} }),
	}
}

func TestSolveBuildsAStackFromTheNamesAndTheDefaults(t *testing.T) {
	sel, err := Solve(webOffers(), []string{"react", "vitest", "bootstrap"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sel.Names(), []string{"react", "vitest", "bootstrap", "eslint"}; !slices.Equal(got, want) {
		t.Fatalf("choices %v, want %v", got, want)
	}
	if !slices.Equal(sel.Added, []string{"eslint"}) {
		t.Errorf("added %v: only eslint was left to the default", sel.Added)
	}
	st := sel.Stack("react-vitest-bootstrap", "")
	if st.Language != "node" || st.Scaffold != "vite" || !slices.Equal(st.Toolchain, []string{"vite", "vitest", "eslint"}) {
		t.Errorf("stack %+v", st)
	}
}

func TestSolveFillsNeededAxesWithTheirDefault(t *testing.T) {
	sel, err := Solve(webOffers(), []string{"vue"})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := sel.Names(), []string{"vue", "eslint", "vitest"}; !slices.Equal(got, want) {
		t.Fatalf("choices %v, want %v", got, want)
	}
}

func TestSolveOrdersAChoiceAfterThoseItRequires(t *testing.T) {
	sel, err := Solve(webOffers(), []string{"chakra", "react"})
	if err != nil {
		t.Fatal(err)
	}
	if sel.Names()[0] != "react" || slices.Index(sel.Names(), "chakra") < slices.Index(sel.Names(), "react") {
		t.Errorf("order %v", sel.Names())
	}
}

func TestSolveAllowsManyChoicesOnAnAxisThatAllowsIt(t *testing.T) {
	if _, err := Solve(webOffers(), []string{"react", "p5", "d3"}); err != nil {
		t.Fatal(err)
	}
}

func TestSolveRefusesWithOneMessageThatNamesTheChoicesAndTheFix(t *testing.T) {
	cases := map[string]struct {
		names []string
		want  []string
	}{
		"axis":        {[]string{"react", "vue"}, []string{"react and vue", "framework axis", "keep one"}},
		"style":       {[]string{"react", "bootstrap", "tailwind"}, []string{"bootstrap and tailwind", "style axis", "keep one"}},
		"requires":    {[]string{"vue", "chakra"}, []string{"chakra requires react", "add react to --with", "drop chakra"}},
		"conflicts":   {[]string{"react", "sass", "p5", "bootstrap"}, []string{"sass and bootstrap"}},
		"language":    {[]string{"react", "flask"}, []string{"react is for node", "flask for python"}},
		"no scaffold": {[]string{"bootstrap"}, []string{"no choice in bootstrap writes the project", "react", "vue"}},
		"unknown":     {[]string{"reakt"}, []string{`"reakt"`, "did you mean react"}},
		"none":        {nil, []string{"--with react"}},
		"owner":       {[]string{"kit", "bootstrap"}, []string{"bootstrap is provided by the vite plugin", "react"}},
	}
	for name, tc := range cases {
		_, err := Solve(webOffers(), tc.names)
		if err == nil {
			t.Errorf("%s: accepted %v", name, tc.names)
			continue
		}
		for _, want := range tc.want {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("%s: %q lacks %q", name, err, want)
			}
		}
		if strings.Count(err.Error(), "\n") > 0 {
			t.Errorf("%s: more than one line: %q", name, err)
		}
	}
}

func TestSolveRefusesANeededAxisWithNoDefault(t *testing.T) {
	offers := slices.DeleteFunc(webOffers(), func(o Offer) bool { return o.Name == "vitest" })
	offers = append(offers, offer("vitest", "test", "vitest", func(o *Offer) { o.Roles = []protocol.Role{protocol.RoleToolchain} }))
	_, err := Solve(offers, []string{"react"})
	if err == nil || !strings.Contains(err.Error(), "react needs a choice on the test axis") || !strings.Contains(err.Error(), "jest, vitest") {
		t.Fatalf("got %v", err)
	}
}

func TestIndexRefusesTwoPluginsOfferingOneName(t *testing.T) {
	_, err := Index([]Offer{offer("a", "test", "x", nil), offer("b", "test", "x", nil)})
	if err == nil || !strings.Contains(err.Error(), "a and b") {
		t.Fatalf("got %v", err)
	}
}

func TestProblemsNameWhyAChoiceIsUnusable(t *testing.T) {
	offers := []Offer{
		offer("vite", "style", "chakra", func(o *Offer) { o.Requires = []string{"react"} }),
		offer("vite", "framework", "svelte", func(o *Offer) { o.Needs = []string{"test"} }),
		offer("pip", "framework", "flask", func(o *Offer) { o.Language = "python" }),
		offer("vite", "style", "bootstrap", nil),
	}
	got := Problems(offers, func(l string) bool { return l == "node" })
	for name, want := range map[string]string{"chakra": "requires react", "svelte": "test axis", "flask": "python plugin is not enabled"} {
		if !strings.Contains(got[name], want) {
			t.Errorf("%s: %q lacks %q", name, got[name], want)
		}
	}
	if _, bad := got["bootstrap"]; bad {
		t.Errorf("bootstrap is usable: %q", got["bootstrap"])
	}
}
