package pluginhost

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// trustWorkspace is a workspace whose smith.toml declares the envdump plugin.
func trustWorkspace(t *testing.T) (root, self string) {
	t.Helper()
	leakyEnvironment(t)
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root = t.TempDir()
	toml := "[workspace]\nname = \"w\"\n\n[plugins.external.envdump]\ncommand = [" + quote(self) + ", \"__envdump\"]\n"
	if err := os.WriteFile(filepath.Join(root, workspace.FileName), []byte(toml), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	return root, self
}

func quote(s string) string { return `'` + s + `'` }

func run(t *testing.T, cmd interface {
	SetArgs([]string)
	Execute() error
}, args ...string) error {
	t.Helper()
	cmd.SetArgs(args)
	return cmd.Execute()
}

func TestTrustWithoutATerminalNeedsYes(t *testing.T) {
	root, _ := trustWorkspace(t)
	trust := trustCommands()[0]
	var out bytes.Buffer
	trust.SetOut(&out)
	err := run(t, trust, "envdump")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "without a terminal") || !strings.Contains(p.Remedy[0], "smith plugin trust envdump --yes") {
		t.Fatalf("got %v", err)
	}
	store, _ := plugintrust.OpenStore()
	if _, ok, _ := store.Find(root, "envdump"); ok {
		t.Fatal("trusted without being asked")
	}
}

func TestTrustYesPinAndUntrust(t *testing.T) {
	root, self := trustWorkspace(t)
	trust, untrust := trustCommands()[0], trustCommands()[1]
	var out bytes.Buffer
	trust.SetOut(&out)
	if err := run(t, trust, "envdump", "--yes", "--pin"); err != nil {
		t.Fatal(err)
	}
	store, _ := plugintrust.OpenStore()
	if _, ok, _ := store.Find(root, "envdump"); !ok {
		t.Fatal("not recorded")
	}
	cfg, err := workspace.Load(root)
	if err != nil {
		t.Fatalf("smith.toml no longer loads: %v", err)
	}
	fp, _ := plugintrust.Measure([]string{self, "__envdump"})
	if got := cfg.Plugins.External["envdump"].SHA256; got != plugintrust.PinnedDigest(fp) {
		t.Fatalf("pinned %q, want %q\n%s", got, plugintrust.PinnedDigest(fp), out.String())
	}
	if !strings.Contains(out.String(), "Trusted envdump") || !strings.Contains(out.String(), "Pinned sha256") {
		t.Fatalf("output: %s", out.String())
	}

	var gone bytes.Buffer
	untrust.SetOut(&gone)
	if err := run(t, untrust, "envdump"); err != nil {
		t.Fatal(err)
	}
	if _, ok, _ := store.Find(root, "envdump"); ok {
		t.Fatal("still trusted")
	}
	if !strings.Contains(gone.String(), "No longer trusting envdump") {
		t.Fatalf("output: %s", gone.String())
	}
	// Starting needs trust again; the pin in smith.toml does not vouch for it
	// on a machine.
	err = pluginadmit.Admit(root, mustResolve(t, root, cfg))
	if !errors.Is(err, pluginadmit.ErrUntrusted) {
		t.Fatalf("got %v", err)
	}
}

func mustResolve(t *testing.T, root string, cfg workspace.Config) []pluginadmit.Plugin {
	t.Helper()
	all, err := ResolveAllExternal(root, cfg.Plugins.External)
	if err != nil {
		t.Fatal(err)
	}
	return all
}

func TestTrustRefusesABuiltInPlugin(t *testing.T) {
	trustWorkspace(t)
	trust := trustCommands()[0]
	err := run(t, trust, BuiltinNames()[0], "--yes")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "built-in plugin") {
		t.Fatalf("got %v", err)
	}
}
