package recipewizard

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestCatalogListsLevelsAndWhatFitsEachLanguage(t *testing.T) {
	cat := fakeCandidates().Catalog()
	levels := map[string]LevelOffer{}
	for _, l := range cat.Levels {
		levels[l.Level] = l
	}
	if cat.Levels[0].Level != LevelLanguage {
		t.Errorf("the language comes first, got %s", cat.Levels[0].Level)
	}
	if !levels[LevelStructure].Additive || levels[LevelLanguage].Additive || levels["lint"].Additive {
		t.Errorf("additive flags are wrong: %+v", levels)
	}
	if len(levels[LevelLanguage].Plugins) != 2 {
		t.Errorf("language plugins = %+v", levels[LevelLanguage].Plugins)
	}
	if got := cat.Fits["python"]["lint"]; !reflect.DeepEqual(got, []string{"ruff"}) {
		t.Errorf("lint plugins that fit python = %v", got)
	}
	if got := cat.Fits["node"]["test"]; !reflect.DeepEqual(got, []string{"jest", "vitest"}) {
		t.Errorf("test plugins that fit node = %v", got)
	}
	if _, ok := cat.Fits["python"][LevelLanguage]; ok {
		t.Error("a language is not listed as fitting itself")
	}
}

func TestWriteCatalogJSONRoundTripsAndTextNamesThePlugins(t *testing.T) {
	cat := fakeCandidates().Catalog()
	var js bytes.Buffer
	if err := WriteCatalog(&js, cat, true); err != nil {
		t.Fatal(err)
	}
	var back LevelCatalog
	if err := json.Unmarshal(js.Bytes(), &back); err != nil || !reflect.DeepEqual(back, cat) {
		t.Fatalf("round trip failed: %v", err)
	}
	var text bytes.Buffer
	if err := WriteCatalog(&text, cat, false); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"language (one plugin)", "structure (any number)", "ruff"} {
		if !strings.Contains(text.String(), want) {
			t.Errorf("text lacks %q:\n%s", want, text.String())
		}
	}
}

func TestListLevelsWritesNothingAndJSONNeedsListLevels(t *testing.T) {
	var out bytes.Buffer
	err := Run(Options{ListLevels: true, AsJSON: true, Out: &out, In: strings.NewReader(""), Names: []string{"python", "ruff"}, Lookup: fakeLookup})
	if err != nil || !strings.Contains(out.String(), `"levels"`) {
		t.Fatalf("err=%v out=%s", err, out.String())
	}
	if err := Run(Options{AsJSON: true, Out: &out, In: strings.NewReader("")}); err == nil {
		t.Error("--json alone was accepted")
	}
}
