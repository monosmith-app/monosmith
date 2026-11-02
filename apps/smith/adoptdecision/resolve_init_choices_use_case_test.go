package adoptdecision

import (
	"bufio"
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func resolve(t *testing.T, r Request, input string) (Choices, string, error) {
	t.Helper()
	var out bytes.Buffer
	r.In, r.Out = bufio.NewReader(strings.NewReader(input)), &out
	c, err := Resolve(r)
	return c, out.String(), err
}

func TestResolveRefusesWithoutATerminalAndNamesTheFlags(t *testing.T) {
	_, _, err := resolve(t, Request{}, "")
	if err == nil {
		t.Fatal("want a refusal")
	}
	for _, want := range []string{"--gitattributes", "--agent", "--ci", "--yes"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal lacks %s:\n%v", want, err)
		}
	}
	// A repository that has .gitattributes is not asked about it.
	_, _, err = resolve(t, Request{Facts: Facts{GitattributesExists: true}}, "")
	if err == nil || strings.Contains(err.Error(), "--gitattributes") {
		t.Fatalf("%v", err)
	}
}

func TestResolveTakesFlagsInSilence(t *testing.T) {
	r := Request{Flags: Flags{Gitattributes: GitattributesAdd, Agent: AgentNever, CI: []string{"gitlab"}, CIGiven: true}}
	c, out, err := resolve(t, r, "")
	if err != nil || out != "" {
		t.Fatalf("%v %q", err, out)
	}
	want := Choices{Gitattributes: GitattributesAdd, Agent: AgentNever, CI: []string{"gitlab"}, Existing: ExistingReplace}
	if !reflect.DeepEqual(c, want) {
		t.Fatalf("got %+v, want %+v", c, want)
	}
}

func TestResolveYesTakesTheSafeDefaultsAndSaysSo(t *testing.T) {
	r := Request{Yes: true, Facts: Facts{Detected: []string{"gitlab"}, MCPExists: true}}
	c, out, err := resolve(t, r, "")
	if err != nil {
		t.Fatal(err)
	}
	if c.Gitattributes != GitattributesSkip || c.Agent != AgentSkip || !reflect.DeepEqual(c.CI, []string{"gitlab"}) {
		t.Fatalf("%+v", c)
	}
	for _, want := range []string{"--gitattributes add", "--agent add", "merged", "gitlab", "--ci"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestResolveNoneDetectedFallsBackToGitHubAndSaysWhy(t *testing.T) {
	c, out, err := resolve(t, Request{Yes: true}, "")
	if err != nil || !reflect.DeepEqual(c.CI, []string{"github"}) || !strings.Contains(out, "no provider was detected") {
		t.Fatalf("%+v %v\n%s", c, err, out)
	}
}

func TestResolveAsksInATerminal(t *testing.T) {
	// gitattributes: yes (2); agent: never (3); ci: None (last); no own pipeline.
	r := Request{Interactive: true}
	c, out, err := resolve(t, r, "2\n3\n6\n")
	if err != nil {
		t.Fatal(err)
	}
	if c.Gitattributes != GitattributesAdd || c.Agent != AgentNever || len(c.CI) != 0 {
		t.Fatalf("%+v\n%s", c, out)
	}
	// Enter takes the safe default of each.
	c, _, _ = resolve(t, Request{Interactive: true, Facts: Facts{Detected: []string{"azure"}}}, "\n\n\n")
	if c.Gitattributes != GitattributesSkip || c.Agent != AgentSkip || !reflect.DeepEqual(c.CI, []string{"azure"}) {
		t.Fatalf("%+v", c)
	}
}

func TestResolveAsksAboutAPipelineOfYours(t *testing.T) {
	r := Request{Interactive: true, Flags: Flags{Gitattributes: GitattributesSkip, Agent: AgentSkip},
		Facts: Facts{Detected: []string{"github"}, OwnPipelines: map[string]string{"github": ".github/workflows/ci.yml"}}}
	c, out, err := resolve(t, r, "\n2\n")
	if err != nil || c.Existing != ExistingKeep || !strings.Contains(out, ".github/workflows/ci.yml") {
		t.Fatalf("%+v %v\n%s", c, err, out)
	}
	if got := WithoutKept(c, r.Facts.OwnPipelines); len(got) != 0 {
		t.Fatalf("a kept pipeline drops the provider: %v", got)
	}
}

func TestResolveRejectsAnUnknownFlagValue(t *testing.T) {
	cases := []Flags{
		{Gitattributes: "maybe"}, {Agent: "sometimes"}, {Existing: "merge"},
		{CI: []string{"jenkins"}, CIGiven: true}, {CI: []string{"none", "github"}, CIGiven: true},
	}
	for _, f := range cases {
		if _, _, err := resolve(t, Request{Flags: f, Yes: true}, ""); err == nil {
			t.Errorf("%+v: want an error", f)
		}
	}
}
