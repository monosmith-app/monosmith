package adoptmove

import "testing"

func TestFixGitignore(t *testing.T) {
	in := "node_modules\n/dist\n!/dist/keep\n/coverage/\n/other\ndist/\n"
	got, changed := FixGitignore(in, "apps/web", []string{"dist", "coverage"}, []string{"other"})
	want := "node_modules\n/apps/web/dist\n!/apps/web/dist/keep\n/apps/web/coverage/\n/other\ndist/\n"
	if !changed || got != want {
		t.Fatalf("got %q (changed=%v), want %q", got, changed, want)
	}
	if _, changed := FixGitignore("build\n", "apps/web", []string{"dist"}, nil); changed {
		t.Error("a line without a leading slash must not change")
	}
}

func TestFixGitignoreKeepsWindowsLineEndings(t *testing.T) {
	got, _ := FixGitignore("/dist\r\nlog\r\n", "libs/x", []string{"dist"}, nil)
	if got != "/libs/x/dist\r\nlog\r\n" {
		t.Fatalf("got %q", got)
	}
}

// Build output the repository does not track (developit/mitt ignores /dist,
// /node_modules and /index.d.ts) is made in the project's folder after the
// move, so its line must also cover that folder (#1155 lab).
func TestFixGitignoreCoversUntrackedOutputInTheProject(t *testing.T) {
	in := "/dist\n/node_modules\n/index.d.ts\n/.github/local\n/*.log\n"
	got, changed := FixGitignore(in, "packages/mitt", []string{"src", "test"}, []string{".github"})
	want := "/dist\n/packages/mitt/dist\n/node_modules\n/packages/mitt/node_modules\n/index.d.ts\n/packages/mitt/index.d.ts\n/.github/local\n/*.log\n"
	if !changed || got != want {
		t.Fatalf("got %q (changed=%v), want %q", got, changed, want)
	}
}
