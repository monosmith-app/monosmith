package adoptmove

import "testing"

func TestFixGitignore(t *testing.T) {
	in := "node_modules\n/dist\n!/dist/keep\n/coverage/\n/other\ndist/\n"
	got, changed := FixGitignore(in, "apps/web", []string{"dist", "coverage"})
	want := "node_modules\n/apps/web/dist\n!/apps/web/dist/keep\n/apps/web/coverage/\n/other\ndist/\n"
	if !changed || got != want {
		t.Fatalf("got %q (changed=%v), want %q", got, changed, want)
	}
	if _, changed := FixGitignore("build\n", "apps/web", []string{"dist"}); changed {
		t.Error("a line without a leading slash must not change")
	}
}

func TestFixGitignoreKeepsWindowsLineEndings(t *testing.T) {
	got, _ := FixGitignore("/dist\r\nlog\r\n", "libs/x", []string{"dist"})
	if got != "/libs/x/dist\r\nlog\r\n" {
		t.Fatalf("got %q", got)
	}
}
