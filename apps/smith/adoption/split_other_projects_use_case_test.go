package adoption

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// machineWarning is a doctor warning about what the runner has installed
// (a toolchain older or newer than CI's, a missing CLI), which no split causes.
var machineWarning = regexp.MustCompile(`^! (.+ is \d[\d.]*, CI uses|.+ is not on PATH)`)

// adoptFixture copies testdata/<fixture> into a new git repository named name,
// runs the built smith there with `adopt --init --move <dir> --yes` and the
// extra arguments, and returns the repository and the adopt output.
func adoptFixture(t *testing.T, fixture, name, dir string, extra ...string) (repo, bin, out string, env []string) {
	t.Helper()
	for _, tool := range []string{"git", "go"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	base := t.TempDir()
	bin = filepath.Join(base, "smith")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	wd, _ := os.Getwd()
	runIn(t, filepath.Dir(wd), nil, "go", "build", "-o", bin, ".")
	repo = filepath.Join(base, name)
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	copyTree(t, filepath.Join("testdata", fixture), repo)
	if err := os.Rename(filepath.Join(repo, "gitignore.txt"), filepath.Join(repo, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	env = []string{"CI=true", "GIT_AUTHOR_NAME=e2e", "GIT_AUTHOR_EMAIL=e2e@example.com", "GIT_COMMITTER_NAME=e2e", "GIT_COMMITTER_EMAIL=e2e@example.com"}
	runIn(t, repo, env, "git", "init", "-q")
	runIn(t, repo, env, "git", "add", "-A")
	runIn(t, repo, env, "git", "commit", "-q", "-m", "init")
	args := append([]string{"adopt", "--init", "--move", dir, "--yes", "--agent", "add"}, extra...)
	out = runIn(t, repo, env, bin, args...)
	return repo, bin, out, env
}

// requireFiles fails the test for each path that is missing (want true) or
// present (want false) in the repository.
func requireFiles(t *testing.T, repo string, want bool, paths ...string) {
	t.Helper()
	for _, rel := range paths {
		_, err := os.Stat(filepath.Join(repo, filepath.FromSlash(rel)))
		if (err == nil) != want {
			t.Errorf("%s: present = %v, want %v", rel, err == nil, want)
		}
	}
}

// requireDoctorPasses fails the test unless smith doctor reports no failure
// and no warning beyond those of the machine.
func requireDoctorPasses(t *testing.T, repo, bin string, env []string) {
	t.Helper()
	doctor := runIn(t, repo, env, bin, "doctor")
	if !regexp.MustCompile(`\d+ passed, 0 failed, \d+ warnings?`).MatchString(doctor) {
		t.Errorf("smith doctor must pass with no failure:\n%s", doctor)
	}
	for _, line := range strings.Split(doctor, "\n") {
		if strings.HasPrefix(line, "! ") && !machineWarning.MatchString(line) {
			t.Errorf("smith doctor warns: %s\n%s", line, doctor)
		}
	}
}

func TestAdoptSplitsASingleGoModule(t *testing.T) {
	if testing.Short() {
		t.Skip("builds smith and adopts a repository")
	}
	repo, bin, out, env := adoptFixture(t, "go-single-app", "orders", "apps/orders-svc")
	if !strings.Contains(out, "wrote go.work: ./apps/orders-svc") {
		t.Errorf("adopt output lacks the go.work line:\n%s", out)
	}
	requireFiles(t, repo, true, "go.work", ".golangci.yml", "README.md", "orders-svc.code-workspace", "doc/NOTES.md",
		"apps/orders-svc/go.mod", "apps/orders-svc/main.go", "apps/orders-svc/orders/total.go")
	requireFiles(t, repo, false, "go.mod", "apps/orders-svc/.golangci.yml", "apps/orders-svc/orders-svc.code-workspace", "apps/orders-svc/NOTES.md")
	requireDoctorPasses(t, repo, bin, env)
}

func TestAdoptSplitsASingleRustCrate(t *testing.T) {
	if testing.Short() {
		t.Skip("builds smith and adopts a repository")
	}
	if _, err := exec.LookPath("cargo"); err != nil {
		t.Skip("cargo is not installed")
	}
	repo, bin, out, env := adoptFixture(t, "rust-single-app", "orders", "apps/orders-cli")
	if !strings.Contains(out, "split: the root Cargo.toml is a workspace with apps/orders-cli") {
		t.Errorf("adopt output lacks the workspace note:\n%s", out)
	}
	requireFiles(t, repo, true, "Cargo.toml", ".cargo/config.toml", "rustfmt.toml", "README.md",
		"apps/orders-cli/Cargo.toml", "apps/orders-cli/src/main.rs")
	requireFiles(t, repo, false, "src", "apps/orders-cli/rustfmt.toml")
	root, _ := os.ReadFile(filepath.Join(repo, "Cargo.toml"))
	crate, _ := os.ReadFile(filepath.Join(repo, "apps", "orders-cli", "Cargo.toml"))
	if !strings.Contains(string(root), `members = ["apps/orders-cli"]`) || !strings.Contains(string(crate), "[dev-dependencies]") {
		t.Errorf("root:\n%s\ncrate:\n%s", root, crate)
	}
	requireDoctorPasses(t, repo, bin, env)
}

func TestAdoptSplitsASinglePythonProject(t *testing.T) {
	if testing.Short() {
		t.Skip("runs pip install")
	}
	if runtime.GOOS != "linux" && os.Getenv("SMITH_E2E_NPM") != "1" {
		t.Skip("creates a venv and installs the tools: linux only, or set SMITH_E2E_NPM=1")
	}
	if _, err := exec.LookPath("python3"); err != nil {
		if _, err := exec.LookPath("python"); err != nil {
			t.Skip("python is not installed")
		}
	}
	repo, bin, out, env := adoptFixture(t, "python-single-app", "orders", "apps/orders-lib", "--install")
	if !strings.Contains(out, "split: the [tool.ruff] tables are in the root ruff.toml") {
		t.Errorf("adopt output lacks the ruff note:\n%s", out)
	}
	requireFiles(t, repo, true, "ruff.toml", "requirements-dev.txt", ".venv", "README.md",
		"apps/orders-lib/pyproject.toml", "apps/orders-lib/src/orders/total.py", "apps/orders-lib/tests/test_total.py")
	requireFiles(t, repo, false, "pyproject.toml", "src")
	project, _ := os.ReadFile(filepath.Join(repo, "apps", "orders-lib", "pyproject.toml"))
	for _, gone := range []string{"pytest>=", "ruff>=", "mypy>=", "tool.ruff"} {
		if strings.Contains(string(project), gone) {
			t.Errorf("the project still holds %q:\n%s", gone, project)
		}
	}
	for _, kept := range []string{`"requests>=2.31"`, "[tool.mypy]", "[tool.pytest.ini_options]"} {
		if !strings.Contains(string(project), kept) {
			t.Errorf("the project lost %q:\n%s", kept, project)
		}
	}
	requireDoctorPasses(t, repo, bin, env)
}
