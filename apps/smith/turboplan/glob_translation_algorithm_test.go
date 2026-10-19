package turboplan

import (
	"reflect"
	"testing"
)

func TestOutputDirs(t *testing.T) {
	got := OutputDirs([]string{"dist/**", ".next/**", "!.next/cache/**", "dist/types/**", "coverage/lcov.info", "*.tsbuildinfo", "$TURBO_ROOT$/x/**", "build"})
	want := Outputs{
		Dirs:     []string{"dist", ".next", "coverage", "build"},
		Excluded: []string{"!.next/cache/**"},
		Loose:    []string{"*.tsbuildinfo", "$TURBO_ROOT$/x/**"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got  %+v\nwant %+v", got, want)
	}
}

func TestInputGlobs(t *testing.T) {
	got := InputGlobs([]string{"$TURBO_DEFAULT$", "!README.md", "$TURBO_ROOT$/tsconfig.json", "src/**"})
	want := []string{"**/*", "!README.md", "{workspaceRoot}/tsconfig.json", "src/**"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
	if RootGlob("./.env") != "{workspaceRoot}/.env" || RootGlob("!**/.env.local") != "!{workspaceRoot}/**/.env.local" {
		t.Fatal("global dependencies are workspace-relative")
	}
}
