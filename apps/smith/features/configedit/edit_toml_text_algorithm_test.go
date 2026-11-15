package configedit

import (
	"strings"
	"testing"
)

const sample = `#:schema https://example.com/smith.schema.json
# A workspace.
schema = 1

[workspace]
name = "acme"   # the name

[ci]
# Pipelines to generate.
providers = [
  "github",  # for now
]
devcontainer = true

# Release settings follow.
# [release]
# branch = "main"

[cache] # where results go
dir = ".smith/cache"
`

func TestSetKeyChangesOnlyTheValue(t *testing.T) {
	got := SetKey(sample, "workspace", "name", `"globex"`)
	want := `name = "globex"   # the name`
	if !strings.Contains(got, want) || len(got) != len(sample)+2 {
		t.Fatalf("got:\n%s", got)
	}
	got = SetKey(sample, "ci", "providers", `["github", "azure"]`)
	if !strings.Contains(got, "# Pipelines to generate.\nproviders = [\"github\", \"azure\"]\ndevcontainer = true") {
		t.Fatalf("a multi-line array is one value:\n%s", got)
	}
}

func TestSetKeyAddsKeysAndTables(t *testing.T) {
	got := SetKey(sample, "cache", "max_size_mb", "512")
	if !strings.Contains(got, "[cache] # where results go\ndir = \".smith/cache\"\nmax_size_mb = 512\n") {
		t.Fatalf("added key:\n%s", got)
	}
	got = SetKey(sample, "workspace", "scope", `"@acme"`)
	if !strings.Contains(got, "name = \"acme\"   # the name\nscope = \"@acme\"\n\n[ci]") {
		t.Fatalf("added key before the blank line:\n%s", got)
	}
	got = SetKey(sample, "ci", "native_runners", `["ubuntu-latest"]`)
	if !strings.Contains(got, "devcontainer = true\nnative_runners = [\"ubuntu-latest\"]\n\n# Release settings follow.") {
		t.Fatalf("the next table's comments stay with it:\n%s", got)
	}
	got = SetKey(sample, "release", "branch", `"trunk"`)
	if !strings.Contains(got, "dir = \".smith/cache\"\n\n[release]\nbranch = \"trunk\"\n") || !strings.Contains(got, `# branch = "main"`) {
		t.Fatalf("added table:\n%s", got)
	}
}

func TestUnsetAndGetKey(t *testing.T) {
	got, removed := UnsetKey(sample, "ci", "providers")
	if !removed || strings.Contains(got, "providers") || strings.Contains(got, "for now") || !strings.Contains(got, "# Pipelines to generate.\ndevcontainer = true") {
		t.Fatalf("unset:\n%s", got)
	}
	if got, _ := UnsetKey(SetKey(sample, "release", "branch", `"trunk"`), "release", "branch"); got != sample {
		t.Fatalf("a table left empty goes:\n%s", got)
	}
	if got, _ := UnsetKey(sample, "cache", "dir"); !strings.Contains(got, "[cache] # where results go") {
		t.Fatalf("a table with a comment stays:\n%s", got)
	}
	if _, removed := UnsetKey(sample, "release", "branch"); removed {
		t.Fatal("a commented-out key is not set")
	}
	if v, ok := GetKey(sample, "workspace", "name"); !ok || v != `"acme"` {
		t.Fatalf("get %q %v", v, ok)
	}
	if v, ok := GetKey(sample, "ci", "providers"); !ok || v != "[\n  \"github\",  # for now\n]" {
		t.Fatalf("get array %q", v)
	}
}
