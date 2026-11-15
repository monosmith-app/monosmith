package guide

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func stepNamed(t *testing.T, key string) Step {
	t.Helper()
	for _, s := range NewWorkspaceSteps() {
		if s.Key == key {
			return s
		}
	}
	t.Fatalf("no step %s", key)
	return Step{}
}

// In a folder that holds files and no repository the form asks what the first
// commit holds; with a secret-looking name it asks the version that names the risk.
func TestTheFormAsksTheFirstCommitOnlyWhereTheFolderHoldsFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	answers := map[string][]string{"name": {"acme"}}
	plain, secrets := stepNamed(t, answerFirstCommit), stepNamed(t, answerFirstCommitSecrets)
	if !plain.Skip(answers) || !secrets.Skip(answers) {
		t.Fatal("a folder that does not exist asks nothing")
	}
	if err := os.MkdirAll("acme", 0o755); err != nil {
		t.Fatal(err)
	}
	if !plain.Skip(answers) || !secrets.Skip(answers) {
		t.Fatal("an empty folder asks nothing")
	}
	if err := os.WriteFile(filepath.Join("acme", "notes.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if plain.Skip(answers) || !secrets.Skip(answers) {
		t.Fatal("a folder with files asks the plain question")
	}
	if err := os.WriteFile(filepath.Join("acme", ".env"), []byte("KEY=1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !plain.Skip(answers) || secrets.Skip(answers) {
		t.Fatal("a folder with a secret-looking name asks the question that names the risk")
	}
	if err := os.MkdirAll(filepath.Join("acme", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if !plain.Skip(answers) || !secrets.Skip(answers) {
		t.Fatal("a repository that exists asks nothing")
	}
}

func TestTheFormAsksTheModulePrefixOnlyForGo(t *testing.T) {
	step := stepNamed(t, answerModulePrefix)
	if !step.Skip(map[string][]string{"name": {"acme"}, "language": {"node"}}) {
		t.Error("a Node workspace has no module prefix")
	}
	go_ := map[string][]string{"name": {"acme"}, "language": {"go"}}
	if step.Skip(go_) {
		t.Error("a Go workspace asks")
	}
	if p := step.PlaceholderFor(go_); !strings.HasSuffix(p, "/acme") {
		t.Errorf("the suggestion ends with the workspace name: %q", p)
	}
}

func TestTheDecisionsBecomeFlags(t *testing.T) {
	for _, tc := range []struct {
		name    string
		answers map[string][]string
		want    []string
	}{
		{"nothing asked, nothing passed", map[string][]string{"name": {"acme"}, "language": {"node"}}, nil},
		{"smith's files", map[string][]string{answerFirstCommit: {"smith"}}, []string{"--first-commit", "smith"}},
		{"no commit", map[string][]string{answerFirstCommit: {"none"}}, []string{"--first-commit", "none"}},
		{"everything", map[string][]string{answerFirstCommit: {"all"}}, []string{"--first-commit", "all"}},
		{"everything with the secrets is the explicit yes", map[string][]string{answerFirstCommitSecrets: {"all+secrets"}}, []string{"--first-commit", "all", "--include-secrets"}},
		{"a typed Go prefix", map[string][]string{"language": {"go"}, answerModulePrefix: {"github.com/acme/app"}}, []string{"--module-prefix", "github.com/acme/app"}},
	} {
		if got := decisionFlags(tc.answers); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
	// Enter on the Go question takes the suggestion, and the flag says which.
	got := decisionFlags(map[string][]string{"name": {"acme"}, "language": {"go"}})
	if len(got) != 2 || got[0] != "--module-prefix" || !strings.HasSuffix(got[1], "/acme") {
		t.Errorf("got %v", got)
	}
}
