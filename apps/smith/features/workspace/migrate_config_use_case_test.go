package workspace

import (
	"strings"
	"testing"
)

func TestMigrateTo(t *testing.T) {
	migrations := []Migration{
		{From: 2, Change: "rename [ci] pipelines to providers", Apply: func(s string) (string, error) {
			return strings.ReplaceAll(s, "pipelines =", "providers ="), nil
		}},
		{From: 1, Change: "add [ci]", Apply: func(s string) (string, error) {
			return s + "\n[ci]\npipelines = [\"github\"]\n", nil
		}},
	}
	src := "# my workspace\nschema = 1 # keep\n\n[workspace]\nname = \"acme\"\n"
	out, changes, err := migrateTo(src, 3, migrations)
	if err != nil {
		t.Fatal(err)
	}
	want := "# my workspace\nschema = 3\n\n[workspace]\nname = \"acme\"\n\n[ci]\nproviders = [\"github\"]\n"
	if out != want || strings.Join(changes, "|") != "schema 1 → 2: add [ci]|schema 2 → 3: rename [ci] pipelines to providers" {
		t.Fatalf("%q\n%v", out, changes)
	}
	again, more, _ := migrateTo(out, 3, migrations)
	if again != out || len(more) != 0 {
		t.Fatal("a migrated file is left alone")
	}
	if _, _, err := migrateTo("schema = 4\n", 3, migrations); err == nil {
		t.Fatal("a newer schema needs a newer smith")
	}
	if _, _, err := migrateTo("[workspace]\nname = \"x\"\n", 2, nil); err == nil || !strings.Contains(err.Error(), "no migration from smith.toml schema 1") {
		t.Fatalf("%v", err)
	}
	if out, changes, _ := MigrateText("[workspace]\nname = \"x\"\n"); len(changes) != 0 || out != "[workspace]\nname = \"x\"\n" {
		t.Fatal("today's schema needs nothing")
	}
}
