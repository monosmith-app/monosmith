package configedit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExplainSettingSaysWhereTheValueComesFromAndWhy(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "smith.toml"), []byte("[workspace]\nname = \"w\"\ncommit_operations = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	set, err := ExplainSetting(dir, "workspace.commit_operations")
	if err != nil || set.Source != "smith.toml" || set.Value != "true" || set.Reason == "" || set.Undo != "smith config unset workspace.commit_operations" {
		t.Fatalf("%+v %v", set, err)
	}
	if text := set.Text(); !strings.Contains(text, "Go back to the default") || !strings.Contains(text, "Why the default:") {
		t.Errorf("text: %s", text)
	}
	def, err := ExplainSetting(dir, "release.branch")
	if err != nil || def.Source != "default" || def.Default != "main" || def.Reason == "" {
		t.Fatalf("%+v %v", def, err)
	}
	if strings.Contains(def.Text(), "Go back to the default") {
		t.Error("a setting at its default has nothing to undo")
	}
	if _, err := ExplainSetting(dir, "release.nope"); err == nil {
		t.Error("an unknown setting was explained")
	}
}

func TestEverySettingWithAReasonShowsItInTheFieldList(t *testing.T) {
	fields, err := Fields()
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, f := range fields {
		if f.Reason != "" {
			n++
		}
	}
	if n < 10 {
		t.Errorf("only %d settings carry a reason", n)
	}
}
