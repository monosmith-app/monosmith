package pluginhost

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCheckNeedsNoTrustButKeepsTheEnvironmentMinimal(t *testing.T) {
	leakyEnvironment(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "env.txt")
	t.Setenv("ENVDUMP_OUT", out)

	cmd := checkCommand("test")
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{self, "__envdump", "--env", "ENVDUMP_OUT", "--env", "AWS_*"})
	// The conformance verdict is not what is under test.
	_ = cmd.Execute()

	env := dumped(t, out)
	if env["AWS_ACCESS_KEY_ID"] != "aws-secret" {
		t.Errorf("--env AWS_* was not honoured: %v", names(env))
	}
	for _, leaked := range []string{"NPM_TOKEN", "SMITH_CACHE_KEY", "UNRELATED_SECRET"} {
		if _, ok := env[leaked]; ok {
			t.Errorf("%s reached a plugin under check", leaked)
		}
	}
	if env["SMITH_PLUGIN"] != "1" {
		t.Error("SMITH_PLUGIN is not set")
	}
}

// examples is docs/examples/plugins, found from this package's directory.
func examples(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("..", "..", "..", "docs", "examples", "plugins"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Skip("the examples are not part of this checkout")
	}
	return dir
}

// checkExample runs `smith plugin check` on a plugin, untrusted and with the
// minimal environment, and requires every check to pass.
func checkExample(t *testing.T, plugin string) {
	t.Helper()
	leakyEnvironment(t)
	cmd := checkCommand("test")
	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{plugin})
	if err := cmd.Execute(); err != nil || strings.Contains(stdout.String(), "✖") {
		t.Fatalf("%s does not pass under a minimal environment: %v\n%s\n%s", plugin, err, stdout.String(), stderr.String())
	}
}

func TestTheGoExampleStillPassesCheckUnderTheMinimalEnvironment(t *testing.T) {
	dir := examples(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is not installed")
	}
	bin := filepath.Join(t.TempDir(), "smith-plugin-widget")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = filepath.Join(dir, "go")
	build.Env = append(os.Environ(), "GOWORK=off")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("building the example: %v\n%s", err, out)
	}
	checkExample(t, bin)
}

func TestTheNodeExamplesStillPassCheckUnderTheMinimalEnvironment(t *testing.T) {
	dir := examples(t)
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node is not installed")
	}
	launch := func(example, script string) string {
		if runtime.GOOS == "windows" {
			return filepath.Join(dir, example, "plugin.cmd")
		}
		return filepath.Join(dir, example, script)
	}
	t.Run("node, no SDK", func(t *testing.T) {
		path := launch("node", "plugin.mjs")
		if info, err := os.Stat(path); err != nil || runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			t.Skip("the launcher is not executable here")
		}
		checkExample(t, path)
	})
	t.Run("typescript, on the SDK", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(dir, "..", "..", "..", "node_modules", "@monosmith", "plugin-sdk")); err != nil {
			t.Skip("the TypeScript SDK is not installed (npm install at the repository root)")
		}
		path := launch("language-widget-ts", "plugin.ts")
		if info, err := os.Stat(path); err != nil || runtime.GOOS != "windows" && info.Mode()&0o111 == 0 {
			t.Skip("the launcher is not executable here")
		}
		checkExample(t, path)
	})
}
