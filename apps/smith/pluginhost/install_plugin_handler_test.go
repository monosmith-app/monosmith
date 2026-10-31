package pluginhost

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/plugintrust"
	"github.com/monosmith-app/monosmith/apps/smith/problem"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// pluginSource is a plugin directory whose program is a copy of this test
// binary, which answers `__envdump` as a real plugin (see TestMain).
func pluginSource(t *testing.T) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exe := ""
	if runtime.GOOS == "windows" {
		exe = ".exe"
	}
	dir := t.TempDir()
	program := filepath.Join(dir, "bin", "envdump-"+runtime.GOOS+"-"+runtime.GOARCH+exe)
	if err := os.MkdirAll(filepath.Dir(program), 0o755); err != nil {
		t.Fatal(err)
	}
	src, err := os.Open(self)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = src.Close() }()
	dst, err := os.OpenFile(program, os.O_WRONLY|os.O_CREATE, 0o755)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		t.Fatal(err)
	}
	if err := dst.Close(); err != nil {
		t.Fatal(err)
	}
	manifest := `{"name":"envdump","version":"0.1.0","command":["./bin/envdump-{os}-{arch}{exe}","__envdump"]}`
	if err := os.WriteFile(filepath.Join(dir, ManifestFile), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func emptyWorkspace(t *testing.T) string {
	t.Helper()
	leakyEnvironment(t)
	t.Setenv("SMITH_WORKSPACE_ROOT", "")
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, workspace.FileName), []byte("[workspace]\nname = \"w\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(root)
	return root
}

func runPack(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := packCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := run(t, cmd, args...)
	return out.String(), err
}

func runInstall(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := installCommand()
	var out bytes.Buffer
	cmd.SetOut(&out)
	err := run(t, cmd, args...)
	return out.String(), err
}

func TestPackThenInstallRunsInCIOnlyThroughThePinAndNeverRunsAtInstall(t *testing.T) {
	root := emptyWorkspace(t)
	source := pluginSource(t)
	dist := t.TempDir()
	marker := filepath.Join(t.TempDir(), "ran.txt")
	t.Setenv("ENVDUMP_OUT", marker)

	out, err := runPack(t, source, "--out", dist)
	if err != nil {
		t.Fatal(err)
	}
	archive := filepath.Join(dist, "envdump-0.1.0.smith-plugin.zip")
	if !strings.Contains(out, "Packed envdump 0.1.0") || !strings.Contains(out, "program  bin/envdump-") {
		t.Fatalf("pack output: %s", out)
	}
	digest, err := os.ReadFile(archive + ".sha256")
	if err != nil {
		t.Fatal(err)
	}

	out, err = runInstall(t, archive, "--sha256", strings.Fields(string(digest))[0], "--pin")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("installing started the plugin")
	}
	for _, want := range []string{"Installed envdump 0.1.0", "matches the digest", "every file matches checksums.sha256", "was not started", "Pinned sha256"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".smith", "plugins", "envdump", ManifestFile)); err != nil {
		t.Fatal("not installed under .smith/plugins")
	}

	cfg, err := workspace.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	pin := cfg.Plugins.External["envdump"].SHA256
	found, err := FindPlugin(root, "envdump")
	if err != nil {
		t.Fatal(err)
	}
	fp, err := plugintrust.Measure(found.Argv)
	if err != nil || pin != plugintrust.PinnedDigest(fp) {
		t.Fatalf("smith.toml pins %q, the installed file is %q (%v)", pin, plugintrust.PinnedDigest(fp), err)
	}

	// On a person's machine the pin alone does not vouch: no --trust, no start.
	cfg.Plugins.Enabled = []string{BuiltinNames()[0]}
	granted := cfg.Plugins.External["envdump"]
	granted.Env = []string{"ENVDUMP_OUT"} // the test's own grant, so the program can report that it ran
	cfg.Plugins.External["envdump"] = granted
	opts := Options{WorkspaceRoot: root, Config: cfg, Version: "test"}
	if _, err := Start(opts); !errors.Is(err, pluginadmit.ErrUntrusted) {
		t.Fatalf("a plugin that was only installed started: %v", err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("the untrusted plugin ran")
	}

	// In CI the pin is what lets it run.
	t.Setenv("CI", "true")
	set, err := Start(opts)
	if err != nil {
		t.Fatalf("in CI the pinned plugin was refused: %v", err)
	}
	set.Close()
	if _, err := os.Stat(marker); err != nil {
		t.Fatal("the pinned plugin did not run in CI")
	}
}

func TestInstallTrustFlagRecordsTheDecisionAndForceIsNeededToReinstall(t *testing.T) {
	root := emptyWorkspace(t)
	source := pluginSource(t)

	if out, err := runInstall(t, source, "--trust"); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	store, _ := plugintrust.OpenStore()
	if _, ok, _ := store.Find(root, "envdump"); !ok {
		t.Fatal("--trust did not record the decision")
	}
	_, err := runInstall(t, source)
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "already installed") {
		t.Fatalf("got %v", err)
	}
	if out, err := runInstall(t, source, "--force"); err != nil || !strings.Contains(out, "Replaced envdump") {
		t.Fatalf("%v\n%s", err, out)
	}
}

func TestInstallGlobalGoesToSmithHomeAndPinsNothing(t *testing.T) {
	root := emptyWorkspace(t)
	source := pluginSource(t)
	out, err := runInstall(t, source, "--global")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if _, err := os.Stat(filepath.Join(Home(), "plugins", "envdump", ManifestFile)); err != nil {
		t.Fatal("not installed in the smith home")
	}
	cfg, _ := workspace.Load(root)
	if _, declared := cfg.Plugins.External["envdump"]; declared || !strings.Contains(out, "Nothing was pinned") {
		t.Fatalf("a global install changed smith.toml or did not say so:\n%s", out)
	}
}

func TestInstallRefusesAWrongDigestAndLeavesNothingBehind(t *testing.T) {
	root := emptyWorkspace(t)
	dist := t.TempDir()
	if _, err := runPack(t, pluginSource(t), "--out", dist); err != nil {
		t.Fatal(err)
	}
	_, err := runInstall(t, filepath.Join(dist, "envdump-0.1.0.smith-plugin.zip"), "--sha256", strings.Repeat("0", 64))
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "sha256") || !strings.Contains(p.Remedy[0], "nothing was installed") {
		t.Fatalf("got %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, ".smith", "plugins")); len(entries) != 0 {
		t.Fatalf("left behind: %v", entries)
	}
	if cfg, _ := workspace.Load(root); len(cfg.Plugins.External) != 0 {
		t.Fatal("smith.toml was changed")
	}
}

func TestPackNeedsAManifestWithAVersion(t *testing.T) {
	emptyWorkspace(t)
	empty := t.TempDir()
	_, err := runPack(t, empty)
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "has no plugin.json") {
		t.Fatalf("got %v", err)
	}
	if err := os.WriteFile(filepath.Join(empty, ManifestFile), []byte(`{"name":"w","command":["w"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runPack(t, empty); !errors.As(err, &p) || !strings.Contains(p.Problem, "no version") {
		t.Fatalf("got %v", err)
	}
}

func TestInstallPinsOnlyWithThePinFlag(t *testing.T) {
	root := emptyWorkspace(t)
	if out, err := runInstall(t, pluginSource(t)); err != nil || !strings.Contains(out, "Nothing was pinned") {
		t.Fatalf("%v\n%s", err, out)
	}
	if cfg, _ := workspace.Load(root); len(cfg.Plugins.External) != 0 {
		t.Fatalf("smith.toml was written without --pin: %v", cfg.Plugins.External)
	}
	if out, err := runInstall(t, pluginSource(t), "--force", "--pin"); err != nil || !strings.Contains(out, "Pinned sha256") {
		t.Fatalf("%v\n%s", err, out)
	}
	if cfg, _ := workspace.Load(root); cfg.Plugins.External["envdump"].SHA256 == "" {
		t.Fatal("--pin did not write the sha256")
	}
}

func TestInstallPinNeedsAWorkspaceAndInstallsNothingOtherwise(t *testing.T) {
	emptyWorkspace(t)
	_, err := runInstall(t, pluginSource(t), "--global", "--pin")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "--pin") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(filepath.Join(Home(), "plugins", "envdump")); err == nil {
		t.Fatal("installed although --pin could not be honoured")
	}
}

func TestInstallRefusesAnArchiveWithoutADigest(t *testing.T) {
	root := emptyWorkspace(t)
	dist := t.TempDir()
	if _, err := runPack(t, pluginSource(t), "--out", dist); err != nil {
		t.Fatal(err)
	}
	_, err := runInstall(t, filepath.Join(dist, "envdump-0.1.0.smith-plugin.zip"))
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "no sha256") || !strings.Contains(strings.Join(p.Remedy, " "), "--sha256") {
		t.Fatalf("got %v", err)
	}
	if entries, _ := os.ReadDir(filepath.Join(root, ".smith", "plugins")); len(entries) != 0 {
		t.Fatalf("left behind: %v", entries)
	}
}

func TestInstallRefusesAPluginWithoutAProgramForThisPlatformEvenGlobally(t *testing.T) {
	emptyWorkspace(t)
	src := t.TempDir()
	manifest := `{"name":"elsewhere","version":"1.0.0","command":["./bin/elsewhere-plan9-mips"]}`
	if err := os.WriteFile(filepath.Join(src, ManifestFile), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := runInstall(t, src, "--global")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "no program for this platform") {
		t.Fatalf("got %v", err)
	}
	if _, err := os.Stat(filepath.Join(Home(), "plugins", "elsewhere")); err == nil {
		t.Fatal("installed a plugin that cannot run here")
	}
}

// marketplaceFor packs the test plugin, serves it with an index that lists it
// under indexDigest ("" means its real digest), points SMITH_MARKETPLACE at
// the index and returns the server.
func marketplaceFor(t *testing.T, indexDigest string) *httptest.Server {
	t.Helper()
	dist := t.TempDir()
	if out, err := runPack(t, pluginSource(t), "--out", dist); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	raw, err := os.ReadFile(filepath.Join(dist, "envdump-0.1.0.smith-plugin.zip.sha256"))
	if err != nil {
		t.Fatal(err)
	}
	if indexDigest == "" {
		indexDigest = strings.Fields(string(raw))[0]
	}
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	mux.HandleFunc("/index.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"schema":1,"plugins":{"envdump":{"description":"d","license":"MIT","versions":{"0.1.0":{"url":"%s/envdump-0.1.0.smith-plugin.zip","sha256":"%s"}}}}}`, srv.URL, indexDigest)
	})
	mux.Handle("/", http.FileServer(http.Dir(dist)))
	t.Setenv("SMITH_MARKETPLACE", srv.URL+"/index.json")
	return srv
}

func TestInstallByNameResolvesThroughTheMarketplaceAndDoesNotTrust(t *testing.T) {
	root := emptyWorkspace(t)
	marketplaceFor(t, "")

	out, err := runInstall(t, "envdump@0.1.0", "--pin")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	for _, want := range []string{"Found envdump@0.1.0 in the marketplace", "Installed envdump 0.1.0", "matches the digest", "was not started", "Pinned sha256"} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
	store, _ := plugintrust.OpenStore()
	if _, ok, _ := store.Find(root, "envdump"); ok {
		t.Fatal("installing by name trusted the plugin")
	}
}

func TestInstallByNameRefusesAnArchiveThatIsNotTheOneTheIndexListsAndGivenSha256(t *testing.T) {
	emptyWorkspace(t)
	marketplaceFor(t, strings.Repeat("0", 64))

	_, err := runInstall(t, "envdump")
	var p *problem.Error
	if !errors.As(err, &p) || !strings.Contains(p.Problem, "sha256") {
		t.Fatalf("got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(".smith", "plugins", "envdump")); statErr == nil {
		t.Fatal("a plugin that failed its digest was installed")
	}
	if _, err := runInstall(t, "envdump", "--sha256", strings.Repeat("0", 64)); err == nil || !strings.Contains(err.Error(), "--sha256 cannot be used") {
		t.Fatalf("got %v", err)
	}
}
