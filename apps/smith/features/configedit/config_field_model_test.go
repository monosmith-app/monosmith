package configedit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

func TestAStringThatTakesAFewWordsIsEditedWithItsChoices(t *testing.T) {
	// workspace.update_check and workspace.confirm also say "type": "string".
	for _, path := range []string{"ci.attest", "release.api_check", "workspace.update_check", "workspace.confirm"} {
		f, err := Find(path)
		if err != nil {
			t.Fatalf("%s: %v", path, err)
		}
		if f.Type != "string" || len(f.Enum) == 0 {
			t.Fatalf("%s: type %q, choices %v", path, f.Type, f.Enum)
		}
		word := f.Enum[0]
		if got, err := f.Encode([]string{word}); err != nil || got != `"`+word+`"` {
			t.Errorf("%s: %q, %v", path, got, err)
		}
		if _, err := f.Encode([]string{"maybe"}); err == nil || !strings.Contains(err.Error(), "takes "+strings.Join(f.Enum, " or ")) {
			t.Errorf("%s: a word outside the choices must name them, got %v", path, err)
		}
	}
}

// Every setting the editors show whose schema lists its values (an enum, or
// the enum of a list's items) carries them as options, each explained, with
// the default marked, so the TUI and the extensions draw a picker and never a
// text box. The walk reads the schema itself, so a new enum without its
// enumDescriptions fails here.
func TestEverySettingWithAFixedSetOfValuesCarriesItsOptionsExplained(t *testing.T) {
	var root struct {
		Properties map[string]struct {
			Properties map[string]map[string]any `json:"properties"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(workspace.Schema, &root); err != nil {
		t.Fatal(err)
	}
	fields, err := Fields()
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]Field{}
	for _, f := range fields {
		byPath[f.Path] = f
	}
	checked := 0
	for table, node := range root.Properties {
		for key, p := range node.Properties {
			values := words(p["enum"])
			if items, ok := p["items"].(map[string]any); ok && len(values) == 0 {
				values = words(items["enum"])
			}
			f, shown := byPath[table+"."+key]
			if len(values) == 0 || !shown {
				continue
			}
			checked++
			if len(f.Options) != len(values) {
				t.Errorf("%s takes %v but is shown with %d options: add enumDescriptions beside its enum in smith.schema.json", f.Path, values, len(f.Options))
				continue
			}
			defaults := 0
			for i, o := range f.Options {
				if o.Value != values[i] || strings.TrimSpace(o.Description) == "" {
					t.Errorf("%s option %q has no one-line description", f.Path, o.Value)
				}
				if o.Default {
					defaults++
				}
			}
			if f.Type == "string" && defaults > 1 {
				t.Errorf("%s marks %d defaults; a setting of one value has at most one", f.Path, defaults)
			}
			if def, ok := p["default"]; ok && defaults == 0 && len(words(def)) > 0 {
				t.Errorf("%s: its default %v is not one of its values", f.Path, def)
			}
		}
	}
	if checked < 10 {
		t.Fatalf("only %d settings with fixed values were found; the walk lost the schema", checked)
	}
}

// The settings of a project ([projects.<name>]) follow the same rule: a fixed
// set of values is shown as options, each explained.
func TestEveryProjectSettingWithAFixedSetOfValuesCarriesItsOptionsExplained(t *testing.T) {
	fields, err := ProjectFields("web")
	if err != nil {
		t.Fatal(err)
	}
	withEnum := 0
	for _, f := range fields {
		if len(f.Enum) == 0 {
			continue
		}
		withEnum++
		if len(f.Options) != len(f.Enum) {
			t.Errorf("%s takes %v but has %d options", f.Path, f.Enum, len(f.Options))
		}
		for _, o := range f.Options {
			if strings.TrimSpace(o.Description) == "" {
				t.Errorf("%s option %q has no description", f.Path, o.Value)
			}
		}
	}
	if withEnum == 0 {
		t.Fatal("kind takes app, lib or package: the walk found no fixed set")
	}
}

func TestTheDefaultOptionIsMarkedForAWordAndForEachWordOfAList(t *testing.T) {
	marked := func(path string) []string {
		f, err := Find(path)
		if err != nil {
			t.Fatal(err)
		}
		var out []string
		for _, o := range f.Options {
			if o.Default {
				out = append(out, o.Value)
			}
		}
		return out
	}
	if got := marked("workspace.confirm"); len(got) != 1 || got[0] != "always" {
		t.Errorf("workspace.confirm default = %v", got)
	}
	if got := marked("ci.triggers"); strings.Join(got, ",") != "push,pull_request" {
		t.Errorf("ci.triggers defaults = %v", got)
	}
	if got := marked("workspace.update_check"); len(got) != 0 {
		t.Errorf("workspace.update_check has no default until the person chooses, got %v", got)
	}
}

// words reads a JSON value as a list of words: one word, or a list of them.
func words(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []any:
		var out []string
		for _, w := range v {
			if s, ok := w.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func TestAListOfTablesIsNotASettingSmithConfigEdits(t *testing.T) {
	if _, err := Find("release.channels"); err == nil {
		t.Error("release.channels is edited in smith.toml: a list of tables")
	}
	fields, _ := Fields()
	for _, f := range fields {
		if f.Type == "" {
			t.Errorf("%s has no type, so smith config could not edit it", f.Path)
		}
	}
}

func TestFindSaysHowToEditReleaseChannels(t *testing.T) {
	_, err := Find("release.channels")
	if err == nil || !strings.Contains(err.Error(), "[[release.channels]]") {
		t.Fatalf("error = %v", err)
	}
}
