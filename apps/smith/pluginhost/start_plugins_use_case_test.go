package pluginhost

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/serve"
)

// envPlugin is an external plugin that writes the environment it was started
// with to the file named by ENVDUMP_OUT, as proof of what it could see.
type envPlugin struct{}

func (envPlugin) Describe(protocol.InitializeParams) (protocol.InitializeResult, error) {
	return protocol.InitializeResult{
		Plugin: protocol.Peer{Name: "envdump", Version: "0.1.0"},
		Kind:   protocol.KindLanguage,
		Roles:  []protocol.Role{protocol.RoleLanguage},
	}, nil
}

// TestMain lets the test binary stand in for smith (the host starts built-in
// plugins as `<self> plugin serve <name>`) and for an external plugin that
// reports its environment (`<self> __envdump`).
func TestMain(m *testing.M) {
	switch {
	case len(os.Args) == 4 && os.Args[1] == "plugin" && os.Args[2] == "serve":
		ctor, ok := Builtin(os.Args[3])
		if !ok {
			os.Exit(2)
		}
		if err := serve.Serve(ctor("test"), os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	case len(os.Args) == 2 && os.Args[1] == "__envdump":
		if out := os.Getenv("ENVDUMP_OUT"); out != "" {
			_ = os.WriteFile(out, []byte(strings.Join(os.Environ(), "\n")), 0o644)
		}
		if err := serve.Serve(envPlugin{}, os.Stdin, os.Stdout); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// leakyEnvironment sets variables a plugin must not see unless told.
func leakyEnvironment(t *testing.T) {
	t.Helper()
	for _, v := range []string{"CI", "GITHUB_ACTIONS", "TF_BUILD"} {
		t.Setenv(v, "")
	}
	t.Setenv("SMITH_HOME", t.TempDir())
	t.Setenv("NPM_TOKEN", "npm-secret")
	t.Setenv("AWS_ACCESS_KEY_ID", "aws-secret")
	t.Setenv("SMITH_CACHE_KEY", "cache-secret")
	t.Setenv("SMITH_RELEASE_BRANCH", "main")
	t.Setenv("UNRELATED_SECRET", "other")
	pluginadmit.SetPrompter(nil)
}

func dumped(t *testing.T, file string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("the plugin did not run: %v", err)
	}
	env := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		if k, v, ok := strings.Cut(line, "="); ok {
			env[k] = v
		}
	}
	return env
}

func TestStartRefusesAnUntrustedExternalPluginAndGivesATrustedOneAMinimalEnvironment(t *testing.T) {
	leakyEnvironment(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ws, out := t.TempDir(), filepath.Join(t.TempDir(), "env.txt")
	t.Setenv("ENVDUMP_OUT", out)
	cfg := workspace.Config{}
	cfg.Plugins.Enabled = []string{BuiltinNames()[0]}
	cfg.Plugins.External = map[string]workspace.ExternalPlugin{
		"envdump": {Command: []string{self, "__envdump"}, Env: []string{"NPM_TOKEN", "ENVDUMP_OUT"}},
	}
	opts := Options{WorkspaceRoot: ws, Config: cfg, Version: "test"}

	if _, err := Start(opts); !errors.Is(err, pluginadmit.ErrUntrusted) {
		t.Fatalf("an untrusted plugin started: %v", err)
	}
	if _, err := os.Stat(out); err == nil {
		t.Fatal("the untrusted plugin ran")
	}

	store, err := plugintrust.OpenStore()
	if err != nil {
		t.Fatal(err)
	}
	fp, err := plugintrust.Measure([]string{self, "__envdump"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Trust(ws, "envdump", fp); err != nil {
		t.Fatal(err)
	}
	set, err := Start(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer set.Close()

	env := dumped(t, out)
	if env["NPM_TOKEN"] != "npm-secret" || env["SMITH_PLUGIN"] != "1" || env["PATH"] == "" && env["Path"] == "" {
		t.Errorf("the plugin lacks what it declared or the baseline: %v", names(env))
	}
	for _, leaked := range []string{"AWS_ACCESS_KEY_ID", "SMITH_CACHE_KEY", "SMITH_RELEASE_BRANCH", "UNRELATED_SECRET"} {
		if _, ok := env[leaked]; ok {
			t.Errorf("%s leaked to an external plugin", leaked)
		}
	}
}

func TestStartGrantsSmithSecretsOnlyFromSmithToml(t *testing.T) {
	leakyEnvironment(t)
	self, _ := os.Executable()
	ws, out := t.TempDir(), filepath.Join(t.TempDir(), "env.txt")
	t.Setenv("ENVDUMP_OUT", out)
	cfg := workspace.Config{}
	cfg.Plugins.Enabled = []string{BuiltinNames()[0]}
	cfg.Plugins.External = map[string]workspace.ExternalPlugin{
		"envdump": {Command: []string{self, "__envdump"}, Env: []string{"ENVDUMP_OUT", "SMITH_CACHE_KEY"}},
	}
	store, _ := plugintrust.OpenStore()
	fp, _ := plugintrust.Measure([]string{self, "__envdump"})
	if err := store.Trust(ws, "envdump", fp); err != nil {
		t.Fatal(err)
	}
	set, err := Start(Options{WorkspaceRoot: ws, Config: cfg, Version: "test"})
	if err != nil {
		t.Fatal(err)
	}
	defer set.Close()
	env := dumped(t, out)
	if env["SMITH_CACHE_KEY"] != "cache-secret" {
		t.Errorf("smith.toml's grant was not honoured: %v", names(env))
	}
	if _, ok := env["SMITH_RELEASE_BRANCH"]; ok {
		t.Error("a secret nobody granted leaked")
	}
}

func names(env map[string]string) []string {
	out := make([]string, 0, len(env))
	for k := range env {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}
