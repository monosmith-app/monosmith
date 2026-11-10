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

// runIn runs a command in dir and fails the test with its output when it fails.
func runIn(t *testing.T, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, out)
	}
	return string(out)
}

// copyTree copies the files of src into dst.
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		if d.IsDir() {
			return os.MkdirAll(filepath.Join(dst, rel), 0o755)
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dst, rel), raw, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

// TestAdoptSplitsTheWorkspaceFromASingleNodeApp runs the built
// smith on a typical single TypeScript app (testdata/node-single-app): the
// workspace files stay at the root, the project folder holds the app, and
// smith doctor finds no failure and no warning. It needs git, npm and the
// network (npm install), so -short skips it.
func TestAdoptSplitsTheWorkspaceFromASingleNodeApp(t *testing.T) {
	if testing.Short() {
		t.Skip("runs npm install")
	}
	// The shared lint config pulls hundreds of packages: on Windows the install takes minutes.
	// Linux covers it; SMITH_E2E_NPM=1 runs it anywhere.
	if runtime.GOOS != "linux" && os.Getenv("SMITH_E2E_NPM") != "1" {
		t.Skip("runs a full npm install: linux only, or set SMITH_E2E_NPM=1")
	}
	for _, tool := range []string{"git", "npm", "go"} {
		if _, err := exec.LookPath(tool); err != nil {
			t.Skipf("%s is not installed", tool)
		}
	}
	base := t.TempDir()
	bin := filepath.Join(base, "smith")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	wd, _ := os.Getwd()
	runIn(t, filepath.Dir(wd), nil, "go", "build", "-o", bin, ".")

	repo := filepath.Join(base, "orders")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatal(err)
	}
	copyTree(t, filepath.Join("testdata", "node-single-app"), repo)
	if err := os.Rename(filepath.Join(repo, "gitignore.txt"), filepath.Join(repo, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	env := []string{"CI=true", "GIT_AUTHOR_NAME=e2e", "GIT_AUTHOR_EMAIL=e2e@example.com", "GIT_COMMITTER_NAME=e2e", "GIT_COMMITTER_EMAIL=e2e@example.com"}
	runIn(t, repo, env, "git", "init", "-q")
	runIn(t, repo, env, "git", "add", "-A")
	runIn(t, repo, env, "git", "commit", "-q", "-m", "init")

	out := runIn(t, repo, env, bin, "adopt", "--init", "--move", "apps/orders-api", "--yes", "--install", "--agent", "add", "--eslint-setup", "shared", "--commit")
	for _, want := range []string{"moved 6 entries into apps/orders-api", "moved 1 document(s) into doc", "moved 6 devDependencies"} {
		if !strings.Contains(out, want) {
			t.Errorf("adopt output lacks %q:\n%s", want, out)
		}
	}

	exists := func(rel string) bool {
		_, err := os.Stat(filepath.Join(repo, filepath.FromSlash(rel)))
		return err == nil
	}
	for _, rel := range []string{
		"orders-api.code-workspace", "README.md", "LICENSE", "docs/architecture.md", ".prettierrc", "doc/MIGRATION_PLAN.md",
		"tsconfig.base.json", "eslint.config.mjs", ".eslintrc.json.pre-smith", "package.json", "package-lock.json", "node_modules",
		"apps/orders-api/package.json", "apps/orders-api/tsconfig.json", "apps/orders-api/jest.config.js",
		"apps/orders-api/src/index.ts", "apps/orders-api/test/total.test.ts", "apps/orders-api/Dockerfile",
	} {
		if !exists(rel) {
			t.Errorf("%s is missing after adopt", rel)
		}
	}
	for _, rel := range []string{"apps/orders-api/orders-api.code-workspace", "apps/orders-api/.prettierrc", "apps/orders-api/.eslintrc.json"} {
		if exists(rel) {
			t.Errorf("%s must not be in the project folder", rel)
		}
	}
	rootPkg, _ := os.ReadFile(filepath.Join(repo, "package.json"))
	projectPkg, _ := os.ReadFile(filepath.Join(repo, "apps", "orders-api", "package.json"))
	for _, dev := range []string{`"typescript"`, `"jest"`, `"ts-jest"`, `"prettier"`, `"@monosmith/eslint-config"`} {
		if !strings.Contains(string(rootPkg), dev) {
			t.Errorf("the root package.json lacks %s:\n%s", dev, rootPkg)
		}
	}
	if strings.Contains(string(projectPkg), "devDependencies") || !strings.Contains(string(projectPkg), `"zod"`) {
		t.Errorf("the project package.json must keep dependencies and lose devDependencies:\n%s", projectPkg)
	}

	doctor := runIn(t, repo, env, bin, "doctor")
	if !regexp.MustCompile(`\d+ passed, 0 failed, \d+ warnings?`).MatchString(doctor) {
		t.Errorf("smith doctor must pass with no failure:\n%s", doctor)
	}
	for _, line := range strings.Split(doctor, "\n") {
		// What the machine has installed (Node.js, Docker, gh, ...) is not the split's doing.
		if strings.HasPrefix(line, "! ") && !strings.HasPrefix(line, "! Node.js is") && !strings.Contains(line, "is not on PATH") {
			t.Errorf("smith doctor warns: %s\n%s", line, doctor)
		}
	}
	if status := runIn(t, repo, env, "git", "status", "--short"); strings.TrimSpace(status) != "" {
		t.Logf("uncommitted after adopt --commit:\n%s", status)
	}
}
