package pipeline

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestAllowedInstallScriptsAddsWhatTheDefaultStackBrings(t *testing.T) {
	ws := t.TempDir()
	manifest := `{"devDependencies": {"typescript": "~6", "jest": "^30", "@monosmith/eslint-config": "^1"}}`
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(manifest), 0o644)
	got := AllowedInstallScripts(ws, []string{"esbuild", "unrs-resolver"})
	want := []string{"esbuild", "unrs-resolver", "@parcel/watcher"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%v, want %v", got, want)
	}
}

func TestAllowedInstallScriptsAddsNothingWithoutTheStack(t *testing.T) {
	ws := t.TempDir()
	if got := AllowedInstallScripts(ws, nil); len(got) != 0 {
		t.Fatalf("no package.json: %v", got)
	}
	_ = os.WriteFile(filepath.Join(ws, "package.json"), []byte(`{"devDependencies": {"vitest": "^3"}}`), 0o644)
	if got := AllowedInstallScripts(ws, []string{"sharp"}); !reflect.DeepEqual(got, []string{"sharp"}) {
		t.Fatalf("%v", got)
	}
}
