package configedit

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// projectWorkspace is a workspace with a Node app at apps/web (its manifest is
// package.json) and a Go program declared in smith.toml at tools/gen.
func projectWorkspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	_ = os.MkdirAll(filepath.Join(root, "apps", "web"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "apps", "web", "package.json"), []byte(`{"name":"web","version":"1.2.0"}`), 0o644)
	_ = os.WriteFile(filepath.Join(root, workspace.FileName), []byte("schema = 1\n\n[workspace]\nname = \"acme\"\n\n[projects.gen]\nroot = \"tools/gen\"\nlanguage = \"go\"\n"), 0o644)
	return root
}

var web = ProjectFacts{Name: "web", Root: "apps/web", Language: "node", Kind: "app", Version: "1.2.0"}

func settingKeys(v ProjectView) []string {
	var out []string
	for _, s := range v.Settings {
		out = append(out, s.Key())
	}
	return out
}

func TestAProjectWithAManifestShowsItsValuesReadOnlyWithTheFileAndOnlyTheSettingsItTakes(t *testing.T) {
	root := projectWorkspace(t)
	v, err := ReadProjectSettings(root, web)
	if err != nil {
		t.Fatal(err)
	}
	if v.Declared || v.Manifest != "apps/web/package.json" {
		t.Fatalf("declared %v, manifest %q", v.Declared, v.Manifest)
	}
	var facts []string
	for _, f := range v.Facts {
		facts = append(facts, f.Key+"="+f.Value)
		if f.Source != "apps/web/package.json" || f.Explanation == "" {
			t.Errorf("%s must say what it is and point at the manifest: %+v", f.Key, f)
		}
	}
	if strings.Join(facts, " ") != "name=web version=1.2.0 language=node kind=app root=apps/web" {
		t.Fatalf("facts %v", facts)
	}
	keys := settingKeys(v)
	for _, owned := range []string{"root", "language", "kind"} {
		if strings.Contains(" "+strings.Join(keys, " ")+" ", " "+owned+" ") {
			t.Errorf("%s comes from the manifest and its plugin, so it is not a setting here: %v", owned, keys)
		}
	}
	for _, want := range []string{"tag_prefix", "bare_tags", "stack", "tags", "split_repo", "registry"} {
		if !strings.Contains(strings.Join(keys, " "), want) {
			t.Errorf("%s is missing from %v", want, keys)
		}
	}
}

func TestADeclaredProjectEditsItsRootLanguageAndKindAndKindIsAChoice(t *testing.T) {
	root := projectWorkspace(t)
	v, err := ReadProjectSettings(root, ProjectFacts{Name: "gen", Root: "tools/gen", Language: "go", Kind: "app"})
	if err != nil {
		t.Fatal(err)
	}
	if !v.Declared || v.Manifest != workspace.FileName {
		t.Fatalf("declared %v, manifest %q", v.Declared, v.Manifest)
	}
	byKey := map[string]ProjectSetting{}
	for _, s := range v.Settings {
		byKey[s.Key()] = s
	}
	if s := byKey["root"]; !s.Set || s.Value != `"tools/gen"` {
		t.Fatalf("root = %+v", s)
	}
	kind := byKey["kind"]
	if kind.Set || len(kind.Options) != 3 || kind.Options[0].Description == "" {
		t.Fatalf("kind is app, lib or package, each explained: %+v", kind)
	}
}

func TestWritingAProjectSettingWritesItsTableAndClearingTheLastRemovesTheTable(t *testing.T) {
	root := projectWorkspace(t)
	file := filepath.Join(root, workspace.FileName)
	if err := Set(root, ProjectPath("web", "tag_prefix"), []string{"web-v"}); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(file)
	if !strings.Contains(string(raw), "[projects.web]\ntag_prefix = \"web-v\"\n") {
		t.Fatalf("smith.toml:\n%s", raw)
	}
	if value, set, _ := Get(root, ProjectPath("web", "tag_prefix")); !set || value != `"web-v"` {
		t.Fatalf("get = %q %v", value, set)
	}
	if err := Unset(root, ProjectPath("web", "tag_prefix")); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(file)
	if strings.Contains(string(raw), "[projects.web]") || !strings.Contains(string(raw), "[projects.gen]") {
		t.Fatalf("the empty table must go and the other stay:\n%s", raw)
	}
}

func TestAProjectSettingTheValidatorsRefuseKeepsTheFileAndSaysWhy(t *testing.T) {
	root := projectWorkspace(t)
	file := filepath.Join(root, workspace.FileName)
	before, _ := os.ReadFile(file)
	for _, tc := range []struct{ project, key, value, why string }{
		{"web", "kind", "lib", "need root"},                 // only a declared project takes kind
		{"gen", "kind", "service", "app or lib or package"}, // the enum
		{"web", "colour", "blue", "is not a setting of [projects.web]"},
	} {
		if err := Set(root, ProjectPath(tc.project, tc.key), []string{tc.value}); err == nil || !strings.Contains(err.Error(), tc.why) {
			t.Errorf("%s.%s = %s: %v, want %q", tc.project, tc.key, tc.value, err, tc.why)
		}
	}
	if err := Set(root, ProjectPath("web", "bare_tags"), []string{"true"}); err != nil {
		t.Fatal(err)
	}
	if err := Set(root, ProjectPath("gen", "bare_tags"), []string{"true"}); err == nil || !strings.Contains(err.Error(), "only one project") {
		t.Fatalf("two projects with bare tags: %v", err)
	}
	if err := Unset(root, ProjectPath("web", "bare_tags")); err != nil {
		t.Fatal(err)
	}
	if after, _ := os.ReadFile(file); string(after) != string(before) {
		t.Fatalf("smith.toml changed:\n%s", after)
	}
}

func TestAProjectNameThatIsNotABareKeyIsQuotedInItsPath(t *testing.T) {
	path := ProjectPath("ms.sharepoint", "tag_prefix")
	f, err := Find(path)
	if err != nil || f.Table() != `projects."ms.sharepoint"` || f.Key() != "tag_prefix" {
		t.Fatalf("%s: table %q key %q, %v", path, f.Table(), f.Key(), err)
	}
}

func TestConfigListWithAProjectPrintsItsViewAndSetChecksTheProjectExists(t *testing.T) {
	root := projectWorkspace(t)
	t.Chdir(root)
	loader := func(_, name string, _ io.Writer) (ProjectFacts, error) {
		if name != "web" {
			return ProjectFacts{}, errors.New(name + " is not a project of this workspace")
		}
		return web, nil
	}
	run := func(args ...string) (string, error) {
		cmd := Command(loader)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(args)
		err := cmd.Execute()
		return out.String(), err
	}
	text, err := run("list", "-p", "web", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var v ProjectView
	if err := json.Unmarshal([]byte(text), &v); err != nil || v.Manifest != "apps/web/package.json" || len(v.Facts) == 0 || len(v.Settings) == 0 {
		t.Fatalf("%v: %s", err, text)
	}
	if out, err := run("set", "-p", "web", "tags", "edge", "public"); err != nil || !strings.Contains(out, `projects.web.tags = ["edge", "public"]`) {
		t.Fatalf("%q, %v", out, err)
	}
	if _, err := run("set", "-p", "nope", "tag_prefix", "x"); err == nil || !strings.Contains(err.Error(), "not a project") {
		t.Fatalf("an unknown project must be refused: %v", err)
	}
	if out, _ := run("list", "-p", "web"); !strings.Contains(out, "web: values from apps/web/package.json") || !strings.Contains(out, "Settings of [projects.web] in smith.toml") {
		t.Fatalf("text:\n%s", out)
	}
}
