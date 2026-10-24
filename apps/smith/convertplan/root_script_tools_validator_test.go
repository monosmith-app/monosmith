package convertplan

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestMissingScriptToolsNamesWhatIsNotInstalled(t *testing.T) {
	root := t.TempDir()
	bin := filepath.Join(root, "node_modules", ".bin")
	if err := os.MkdirAll(bin, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bin, "prettier"), nil, 0o755); err != nil {
		t.Fatal(err)
	}
	got := MissingScriptTools(root, "oxfmt --check && prettier . && npm run x; NODE_ENV=a notinstalled-tool; ./local.sh")
	want := []string{"oxfmt", "notinstalled-tool"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMissingScriptToolsJudgesNothingBeforeDependenciesAreInstalled(t *testing.T) {
	if got := MissingScriptTools(t.TempDir(), "oxfmt --check"); got != nil {
		t.Fatalf("got %v, want none", got)
	}
}
