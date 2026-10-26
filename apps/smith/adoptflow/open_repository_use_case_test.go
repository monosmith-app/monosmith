package adoptflow

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/convertchoices"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// TestMain lets the test binary stand in for smith: the plugin host starts
// built-in plugins as `<self> plugin serve <name>`, and without this the test
// binary would re-run the whole suite for each plugin and multiply without end.
func TestMain(m *testing.M) {
	if len(os.Args) == 4 && os.Args[1] == "plugin" && os.Args[2] == "serve" {
		ctor, ok := pluginhost.Builtin(os.Args[3])
		if !ok {
			os.Exit(2)
		}
		if err := serve.Serve(ctor("test"), os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func git(t *testing.T, root string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", root, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// turborepo is a committed Turborepo workspace with ESLint, Jest and rollup.
func turborepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, "package.json", `{"name":"acme","private":true,"workspaces":["packages/*"],"devDependencies":{"turbo":"^2.0.0","eslint":"^9.0.0","jest":"^29.0.0","rollup":"^4.0.0"}}`)
	write(t, root, "turbo.json", `{"tasks":{"build":{"outputs":["dist/**"]},"test":{}}}`)
	write(t, root, "eslint.config.mjs", "export default [];\n")
	write(t, root, "jest.config.js", "module.exports = {};\n")
	write(t, root, "packages/a/package.json", `{"name":"a","version":"1.0.0","scripts":{"build":"tsc","test":"jest"}}`)
	write(t, root, "packages/a/src/index.ts", "export const a = 1;\n")
	git(t, root, "init", "-q")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "init")
	return root
}

func TestAFolderThatIsNotARepositoryIsABlockerNotAnError(t *testing.T) {
	s, err := Open(t.TempDir(), "test")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f := s.Found()
	if f.Kind != KindNotARepo || len(f.Blockers) != 1 || !strings.Contains(f.Blockers[0], "not a git repository") || f.Full {
		t.Fatalf("found %+v", f)
	}
}

func TestAnExistingSmithWorkspaceHasNothingToAdopt(t *testing.T) {
	root := t.TempDir()
	write(t, root, "smith.toml", "[workspace]\nname = \"x\"\n")
	s, err := Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	if f := s.Found(); f.Kind != KindSmith || len(f.Blockers) != 1 || !strings.Contains(f.Blockers[0], "already a smith workspace") {
		t.Fatalf("found %+v", f)
	}
}

func TestATurborepoWorkspaceIsReadForTheWizard(t *testing.T) {
	root := turborepo(t)
	s, err := Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f := s.Found()
	if f.Kind != KindTurborepo || !f.Full || len(f.Blockers) != 0 {
		t.Fatalf("found %+v", f)
	}
	if len(f.Projects) != 1 || f.Projects[0].Name != "a" || f.Projects[0].Root != "packages/a" {
		t.Fatalf("projects %+v", f.Projects)
	}
	for _, tool := range []string{"Turborepo", "ESLint", "Jest", "rollup"} {
		found := false
		for _, got := range f.Tooling {
			found = found || got == tool
		}
		if !found {
			t.Errorf("tooling %v lacks %s", f.Tooling, tool)
		}
	}
}

func TestADirtyTreeBlocksTheConversionAndSaysHowToClearIt(t *testing.T) {
	root := turborepo(t)
	write(t, root, "packages/a/src/index.ts", "export const a = 2;\n")
	s, err := Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	f := s.Found()
	if len(f.Blockers) != 1 || !strings.Contains(f.Blockers[0], "uncommitted changes") || !strings.Contains(f.Blockers[0], "git commit") {
		t.Fatalf("found %+v", f)
	}
}

func TestThePlanOfTheFullLevelIsMadeOnAScratchCopyAndLeavesTheRepositoryAlone(t *testing.T) {
	root := turborepo(t)
	s, err := Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	def, questions, err := s.Prepare()
	if err != nil {
		t.Fatal(err)
	}
	if def.Level != convertchoices.LevelFull || def.Test != convertchoices.Jest || len(questions) != 4 {
		t.Fatalf("defaults %+v with %d questions", def, len(questions))
	}
	preview, err := s.PlanFull(def, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(preview.Text, "Full conversion") || !strings.Contains(preview.Text, "step") {
		t.Fatalf("the review lacks the plan and the closing checks:\n%s", preview.Text)
	}
	if _, err := os.Stat(filepath.Join(root, "smith.toml")); err == nil {
		t.Fatal("planning wrote smith.toml into the repository")
	}
	status := exec.Command("git", "-C", root, "status", "--porcelain")
	if out, _ := status.Output(); len(out) != 0 {
		t.Fatalf("planning changed the repository:\n%s", out)
	}
}

func TestAPlainRepositoryHasNoFullLevelYetAndTheReasonNamesTheIssue(t *testing.T) {
	root := t.TempDir()
	write(t, root, "go.mod", "module example.com/solo\n\ngo 1.22\n")
	write(t, root, "main.go", "package main\n\nfunc main() {}\n")
	git(t, root, "init", "-q")
	git(t, root, "add", "-A")
	git(t, root, "commit", "-q", "-m", "init")
	s, err := Open(root, "test")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	f := s.Found()
	if f.Kind != KindPlain || f.Full || !strings.Contains(f.NoFull, IssueFullForPlain) {
		t.Fatalf("found %+v", f)
	}
	if _, _, err := s.Prepare(); err != ErrNotConvertible {
		t.Fatalf("Prepare on a plain repository: %v", err)
	}
	fix := s.AdoptFix()
	if len(fix.Steps) != 1 || fix.Steps[0].Argv[1] != "adopt" || fix.Steps[0].Argv[2] != "--init" {
		t.Fatalf("adopt fix %+v", fix)
	}
}
