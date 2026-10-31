package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAppliesDefaults(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"acme\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Schema != 1 || cfg.CI.Providers[0] != "github" || cfg.Release.Branch != "main" || cfg.Cache.Dir != ".smith/cache" || cfg.Cache.MaxSizeMB != 2048 {
		t.Fatalf("defaults %+v", cfg)
	}
}

func TestParseReadsAnExternalPluginsEnvAndPin(t *testing.T) {
	pin := strings.Repeat("ab", 32)
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[plugins.external.kiwi]\ncommand = [\"kiwi\"]\nenv = [\"NPM_TOKEN\", \"AWS_*\"]\nsha256 = \"" + pin + "\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	if k := cfg.Plugins.External["kiwi"]; len(k.Env) != 2 || k.SHA256 != pin {
		t.Fatalf("got %+v", k)
	}
	for input, want := range map[string]string{
		"[workspace]\nname = \"a\"\n[plugins.external.kiwi]\nsha256 = \"abc\"\n":    "sha256 must be 64 hexadecimal",
		"[workspace]\nname = \"a\"\n[plugins.external.kiwi]\nenv = [\"A\", \"\"]\n": "env has an empty name",
	} {
		if _, err := Parse([]byte(input)); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: %v, want %q", input, err, want)
		}
	}
}

func TestParseRejectsMistakes(t *testing.T) {
	for input, want := range map[string]string{
		"[workspace]\nname = \"a\"\nnmae = \"typo\"\n": "nmae",
		"[workspace]\n": "name is required",
		"[workspace]\nname = \"a\"\n[ci]\naudit_policies=[\"cobol\"]":                                                       "unknown audit policy",
		"[workspace]\nname = \"a\"\n[ci]\naudit_max_unsafe=-1":                                                              "audit_max_unsafe",
		"[workspace]\nname = \"a\"\n[ci]\nproviders=[\"x\"]":                                                                "unknown CI provider",
		"[workspace]\nname = \"a\"\n[release]\nrubygems_trusted_publishing = true\n[ci]\nproviders=[\"github\",\"gitlab\"]": "works on GitHub Actions only",
		"[workspace]\nname = \"a\"\n[release]\nvsce_entra = true\n[ci]\nproviders=[\"github\",\"azure\"]":                   "vsce_entra works on GitHub Actions only",
		"[workspace]\nname = \"a\"\n[release]\napi_check = \"strict\"\n": "api_check must be",
		"schema = 9\n[workspace]\nname = \"a\"\n":                                                                         "upgrade smith",
		"[workspace\nname = \"a\"\n":                                                                                        "smith.toml:",
		"[workspace]\nname = \"a\"\n[conventions.vertical-slices]\nregistry = \"ghcr.io/a\"\n":                              "settings of the docker convention",
		"[workspace]\nname = \"a\"\n[projects.api.docker]\nimage = \"api\"\n":                                               "needs [conventions.docker]",
		"[workspace]\nname = \"a\"\nrecipe = \"http://example.com/team.toml\"\n":                                            "[workspace] recipe",
		"[workspace]\nname = \"a\"\nrecipe = \"../team.toml\"\n":                                                            "[workspace] recipe",
		"[workspace]\nname = \"a\"\n[recipes]\nservice = \"team.toml\"\n":                                                   "the kind is app, lib or package",
		"[workspace]\nname = \"a\"\n[recipes]\napp = \"C:\\\\team.toml\"\n":                                                 "[recipes] app",
	} {
		_, err := Parse([]byte(input))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("input %q: want error containing %q, got %v", input, want, err)
		}
	}
}

func TestParseTellsANewerSmithFromATypo(t *testing.T) {
	for input, newer := range map[string]bool{
		"[workspace]\nname = \"a\"\nnmae = \"typo\"\n":                     false,
		"[workspace]\nname = \"a\"\n[future]\nk = 1\n":                     true,
		"[workspace]\nname = \"a\"\n[ci]\nfoo = \"x\"\nother = 1\n": true,
		"schema = 99\n[workspace]\nname = \"a\"\nnmae = \"x\"\n":           true,
		"schema = 99\n[workspace]\nname = \"a\"\n":                         true,
	} {
		_, err := Parse([]byte(input))
		if _, got := AsNewerSmith(err); got != newer {
			t.Errorf("input %q: newer smith = %v, want %v (%v)", input, got, newer, err)
		}
	}
}

func TestParseRecipesAndProjectSettings(t *testing.T) {
	cfg, err := Parse([]byte(`[workspace]
name = "a"
recipe = "recipes/team.recipe.toml"
[recipes]
package = "https://example.com/pkg.recipe.toml"
[projects.api.settings.dotnet]
target_framework = "net10.0"
[projects.api.settings.dotnet.nested]
on = true
`))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Workspace.Recipe != "recipes/team.recipe.toml" || cfg.Recipes["package"] != "https://example.com/pkg.recipe.toml" {
		t.Fatalf("recipes %+v %+v", cfg.Workspace, cfg.Recipes)
	}
	dotnet := cfg.Projects["api"].Settings["dotnet"]
	if dotnet["target_framework"] != "net10.0" {
		t.Fatalf("settings %+v", cfg.Projects["api"].Settings)
	}
	if nested, _ := dotnet["nested"].(map[string]any); nested["on"] != true {
		t.Fatalf("nested settings %+v", dotnet)
	}
	if got := RecipeFor(cfg, "app"); got != "recipes/team.recipe.toml" {
		t.Errorf("recipe for app: %q", got)
	}
}

func TestParseOverrides(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[tasks.test]\ncache = false\n[projects.web.tasks.build]\ndepends_on = [\"^build\", \"codegen\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if *cfg.Tasks["test"].Cache || cfg.Projects["web"].Tasks["build"].DependsOn[1] != "codegen" {
		t.Fatalf("overrides %+v", cfg)
	}
}

func TestFindRoot(t *testing.T) {
	t.Setenv("SMITH_WORKSPACE_ROOT", "") // set when the test runs as a smith task
	root := t.TempDir()
	deep := filepath.Join(root, "apps", "web", "src")
	_ = os.MkdirAll(deep, 0o755)
	_ = os.WriteFile(filepath.Join(root, FileName), []byte("[workspace]\nname=\"a\"\n"), 0o644)
	got, err := FindRoot(deep)
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.EvalSymlinks(root)
	got, _ = filepath.EvalSymlinks(got)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
	if _, err := FindRoot(t.TempDir()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestParseProjectReleaseAndTagPrefix(t *testing.T) {
	cfg, err := Parse([]byte("[workspace]\nname = \"a\"\n[release]\ngithub_releases = true\n[projects.cli]\ntag_prefix = \"@a/cli@\"\n[projects.cli.release]\nproduct = \"acme\"\n[projects.cli.tasks.slice]\ncommand = [\"node\", \"check.cjs\"]\ncwd = \".\"\n"))
	if err != nil {
		t.Fatal(err)
	}
	cli := cfg.Projects["cli"]
	if !cfg.Release.GitHubReleases || cli.TagPrefix != "@a/cli@" || cli.Tasks["slice"].Cwd != "." {
		t.Fatalf("%+v", cfg)
	}
	release, err := ResolveRelease("cli", cli, nil, true)
	if err != nil || release.Product != "acme" || len(release.Platforms) != 6 || release.AssetName != "{product}_{version}_{os}_{arch}.{ext}" {
		t.Fatalf("release defaults %+v %v", release, err)
	}
	bad, err := Parse([]byte("[workspace]\nname = \"a\"\n[projects.cli.release]\nplatforms = [\"linux\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ResolveRelease("cli", bad.Projects["cli"], nil, false); err == nil || !strings.Contains(err.Error(), "os/arch") {
		t.Fatalf("a bad platform must be rejected: %v", err)
	}
}

func TestFindRootFromIgnoresTheWorkspaceRootOfTheRunningTask(t *testing.T) {
	task, copyDir := t.TempDir(), t.TempDir()
	for _, dir := range []string{task, copyDir} {
		if err := os.WriteFile(filepath.Join(dir, FileName), []byte("schema = 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("SMITH_WORKSPACE_ROOT", task)
	if got, err := FindRoot(copyDir); err != nil || got != task {
		t.Fatalf("FindRoot follows the task's workspace: got %q, %v", got, err)
	}
	if got, err := FindRootFrom(copyDir); err != nil || got != copyDir {
		t.Fatalf("FindRootFrom must stay on the directory it was given: got %q, %v", got, err)
	}
}

func TestCacheSizeLimitIsSetByTheFileAndZeroMeansNoLimit(t *testing.T) {
	const head = "[workspace]\nname = \"a\"\n"
	for input, want := range map[string]int{
		head:                                  2048, // absent: the default
		head + "[cache]\ndir = \"x\"\n":       2048, // the table without the key
		head + "[cache]\nmax_size_mb = 512\n": 512,
		head + "[cache]\nmax_size_mb = 0\n":   -1, // written as 0: no limit
		head + "[cache]\nmax_size_mb = -5\n":  -5, // negative: no limit
	} {
		cfg, err := Parse([]byte(input))
		if err != nil || cfg.Cache.MaxSizeMB != want {
			t.Errorf("%q: MaxSizeMB = %d (%v), want %d", input, cfg.Cache.MaxSizeMB, err, want)
		}
	}
}
