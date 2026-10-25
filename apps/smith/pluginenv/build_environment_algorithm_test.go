package pluginenv

import (
	"slices"
	"strings"
	"testing"
)

func names(env []string) []string {
	out := make([]string, 0, len(env))
	for _, e := range env {
		n, _, _ := strings.Cut(e, "=")
		out = append(out, n)
	}
	slices.Sort(out)
	return out
}

var parent = []string{
	"PATH=/bin", "HOME=/home/a", "LANG=en", "LC_ALL=C", "TERM=xterm", "HTTPS_PROXY=http://p",
	"SMITH_PLUGIN=1", "SMITH_CACHE_KEY=secret", "SMITH_CACHE_TOKEN=t", "SMITH_RELEASE_BRANCH=main",
	"NPM_TOKEN=npm", "AWS_ACCESS_KEY_ID=a", "AWS_SECRET_ACCESS_KEY=b", "GITHUB_TOKEN=g", "MY_VAR=1",
}

func TestBuildGivesOnlyTheBaseline(t *testing.T) {
	got := names(Build(parent, nil, nil, "linux"))
	want := []string{"HOME", "HTTPS_PROXY", "LANG", "LC_ALL", "PATH", "SMITH_PLUGIN", "TERM"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestBuildAddsDeclaredNamesAndPatterns(t *testing.T) {
	got := names(Build(parent, []string{"NPM_TOKEN", "AWS_*"}, nil, "linux"))
	for _, n := range []string{"NPM_TOKEN", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY"} {
		if !slices.Contains(got, n) {
			t.Errorf("%s was not passed: %v", n, got)
		}
	}
	for _, n := range []string{"GITHUB_TOKEN", "MY_VAR", "SMITH_CACHE_KEY"} {
		if slices.Contains(got, n) {
			t.Errorf("%s leaked: %v", n, got)
		}
	}
}

func TestAManifestCannotAskForSmithSecrets(t *testing.T) {
	got := names(Build(parent, []string{"SMITH_*", "SMITH_CACHE_KEY"}, nil, "linux"))
	for _, n := range []string{"SMITH_CACHE_KEY", "SMITH_CACHE_TOKEN", "SMITH_RELEASE_BRANCH"} {
		if slices.Contains(got, n) {
			t.Errorf("%s was granted by a manifest: %v", n, got)
		}
	}
}

func TestSmithTomlCanGrantSmithSecrets(t *testing.T) {
	got := names(Build(parent, nil, []string{"SMITH_CACHE_KEY"}, "linux"))
	if !slices.Contains(got, "SMITH_CACHE_KEY") || slices.Contains(got, "SMITH_CACHE_TOKEN") {
		t.Fatalf("got %v", got)
	}
}

func TestWindowsNamesIgnoreCaseAndKeepTheSystemVariables(t *testing.T) {
	env := []string{"Path=C:\\x", "SystemRoot=C:\\Windows", "ComSpec=cmd", "aws_region=eu", "Other=1"}
	got := names(Build(env, []string{"AWS_*"}, nil, "windows"))
	want := []string{"ComSpec", "Path", "SystemRoot", "aws_region"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	if slices.Contains(names(Build(env, nil, nil, "linux")), "SystemRoot") {
		t.Fatal("Windows variables are in the Linux baseline")
	}
}

func TestVisibleListsNamesOnly(t *testing.T) {
	got := Visible([]string{"AWS_*", "SMITH_CACHE_KEY", "NPM_TOKEN"}, []string{"NPM_TOKEN", "SMITH_CACHE_KEY"})
	want := []string{"AWS_*", "NPM_TOKEN", "SMITH_CACHE_KEY"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
