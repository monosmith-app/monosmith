package composeview

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/compose"
)

func TestTheRefusalNamesTheFlagsThatWorkWithoutATerminal(t *testing.T) {
	msg := Refusal().Error()
	for _, want := range []string{"needs a terminal", "--json", "--with", "--search"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("the refusal lacks %q: %s", want, msg)
		}
	}
}

func TestTheCommandHasTheFlagsTheScreensCall(t *testing.T) {
	cmd := Command("test")
	for _, flag := range []string{"json", "search", "lang", "with", "kind", "name"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Fatalf("smith compose has no --%s", flag)
		}
	}
}

func TestFunctionHostsNeedTheirLocalToolAndOtherChoicesNeedNone(t *testing.T) {
	if tools := functionTools("aws-lambda", "node"); len(tools) != 1 || tools[0].Binary != "sam" {
		t.Fatalf("%+v", tools)
	}
	if tools := functionTools("react", "node"); tools != nil {
		t.Fatalf("%+v", tools)
	}
}

func TestChoiceLinesSayWhatEachChoiceWritesInThePluginsWords(t *testing.T) {
	cat := compose.Catalogue{Items: []compose.Item{{ID: "node/react", Name: "react", Title: "React", Description: "a React app"}}}
	lines := choiceLines(cat, compose.Check{Language: "node", Selection: []string{"react", "unknown"}})
	if len(lines) != 1 || lines[0].Path != "React" || lines[0].Note != "a React app" {
		t.Fatalf("%+v", lines)
	}
}

func TestTheCatalogueIsBuiltFromWhatPluginsPublishAndNotFromAList(t *testing.T) {
	cat := compose.Build(compose.Source{Languages: []compose.Language{{Name: "node"}}})
	if len(cat.Items) != 1 || cat.Items[0].Kind != compose.KindLanguage {
		t.Fatalf("with no choices only the language is listed: %+v", cat.Items)
	}
}
