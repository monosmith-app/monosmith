package scaffold

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func folderWith(t *testing.T, files ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range files {
		p := filepath.Join(dir, filepath.FromSlash(f))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestChooseFirstCommit(t *testing.T) {
	cases := []struct {
		name     string
		files    []string
		req      CommitRequest
		input    string
		wantMode string
		wantSec  bool
		wantErr  string
	}{
		{"empty folder keeps the usual rule", nil, CommitRequest{}, "", "", false, ""},
		{"empty folder, no commit", nil, CommitRequest{NoCommit: true}, "", CommitNone, false, ""},
		{"files, no terminal, no flag: refused", []string{"a.txt"}, CommitRequest{}, "", "", false, "--first-commit smith"},
		{"files, --yes: only smith's files", []string{"a.txt"}, CommitRequest{Yes: true}, "", CommitSmith, false, ""},
		{"files, flag all, no secret", []string{"a.txt"}, CommitRequest{Mode: CommitAll}, "", CommitAll, false, ""},
		{"files, flag all with a secret, no terminal: refused", []string{"a.txt", ".env"}, CommitRequest{Mode: CommitAll}, "", "", false, "--include-secrets"},
		{"files, flag all with a secret and --include-secrets", []string{".env"}, CommitRequest{Mode: CommitAll, IncludeSecrets: true}, "", CommitAll, true, ""},
		{"terminal: Enter takes smith's files", []string{"a.txt"}, CommitRequest{Interactive: true}, "\n", CommitSmith, false, ""},
		{"terminal: nothing", []string{"a.txt"}, CommitRequest{Interactive: true}, "2\n", CommitNone, false, ""},
		{"terminal: everything, secrets declined", []string{".env"}, CommitRequest{Interactive: true}, "3\n\n", CommitAll, false, ""},
		{"terminal: everything, secrets confirmed", []string{".env"}, CommitRequest{Interactive: true}, "3\n2\n", CommitAll, true, ""},
		{"no-commit with first-commit all contradict", []string{"a.txt"}, CommitRequest{NoCommit: true, Mode: CommitAll}, "", "", false, "opposite"},
		{"unknown mode", nil, CommitRequest{Mode: "most"}, "", "", false, "none, smith or all"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			c.req.Dir = folderWith(t, c.files...)
			c.req.In, c.req.Out = bufio.NewReader(strings.NewReader(c.input)), &out
			got, err := ChooseFirstCommit(c.req)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("got err %v, want one with %q", err, c.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Mode != c.wantMode || got.IncludeSecrets != c.wantSec {
				t.Fatalf("got %+v, want mode %q secrets %v\n%s", got, c.wantMode, c.wantSec, out.String())
			}
		})
	}
}

func TestChooseFirstCommitLeavesAnExistingRepositoryAlone(t *testing.T) {
	dir := folderWith(t, "a.txt", ".git/HEAD")
	got, err := ChooseFirstCommit(CommitRequest{Dir: dir, Out: &bytes.Buffer{}})
	if err != nil || got.Mode != "" {
		t.Fatalf("%+v %v", got, err)
	}
}

func TestCommitAllPathsLeavesSecretsOutUnlessAllowed(t *testing.T) {
	before := FolderContent{Files: []string{".env", "a.txt", "node_modules/"}}
	if got := commitAllPaths(before, []string{"smith.toml"}, false); strings.Join(got, ",") != "a.txt,smith.toml" {
		t.Fatalf("%v", got)
	}
	if got := commitAllPaths(before, nil, true); strings.Join(got, ",") != ".env,a.txt" {
		t.Fatalf("%v", got)
	}
}
