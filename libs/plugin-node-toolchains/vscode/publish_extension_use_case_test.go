package vscode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func extensionWorkspace(t *testing.T, built bool) string {
	t.Helper()
	ws := t.TempDir()
	manifest := `{"name":"ext","version":"0.10.0","publisher":"acme","engines":{"vscode":"^1.90.0"},"main":"../../dist/apps/ext/main.js"}`
	for dir, ok := range map[string]bool{"apps/ext": true, "dist/apps/ext": built} {
		if !ok {
			continue
		}
		if err := os.MkdirAll(filepath.Join(ws, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(ws, filepath.FromSlash(dir), "package.json"), []byte(manifest), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return ws
}

func recordRuns(t *testing.T) *[]string {
	t.Helper()
	real := run
	t.Cleanup(func() { run = real })
	var got []string
	run = func(_ string, argv []string) error { got = append(got, strings.Join(argv, " ")); return nil }
	return &got
}

func extensionApp() protocol.PublishParams {
	return protocol.PublishParams{Project: protocol.Project{Name: "ext", Root: "apps/ext", Kind: protocol.ProjectApp}, Version: "0.11.0"}
}

func TestPublishStampsTheVersionPacksAndPublishesToBothRegistries(t *testing.T) {
	ws := extensionWorkspace(t, true)
	t.Setenv("VSCE_PAT", "token")
	t.Setenv("OVSX_PAT", "token")
	t.Setenv("VSCE_AZURE_CREDENTIAL", "")
	runs := recordRuns(t)
	result, err := publishExtension(ws, extensionApp())
	if err != nil || !result.Published {
		t.Fatalf("published %v: %v", result, err)
	}
	raw, _ := os.ReadFile(filepath.Join(ws, "dist", "apps", "ext", "package.json"))
	if !strings.Contains(string(raw), `"version": "0.11.0"`) {
		t.Errorf("the build's manifest is not stamped:\n%s", raw)
	}
	if src, _ := os.ReadFile(filepath.Join(ws, "apps", "ext", "package.json")); !strings.Contains(string(src), "0.10.0") {
		t.Error("the source manifest changed")
	}
	if len(*runs) != 3 || !strings.Contains((*runs)[0], "vsce package") || !strings.Contains((*runs)[1], "vsce publish --packagePath") || !strings.HasSuffix((*runs)[1], "--skip-duplicate") || !strings.Contains((*runs)[2], "ovsx publish") {
		t.Errorf("commands: %v", *runs)
	}
	if strings.Contains((*runs)[1], "--azure-credential") {
		t.Error("a token publish used the Azure credential")
	}
	if !strings.Contains(result.Registry, "open-vsx.org") {
		t.Errorf("registry %q", result.Registry)
	}
}

func TestPublishWithEntraAndWithoutOpenVSX(t *testing.T) {
	ws := extensionWorkspace(t, true)
	t.Setenv("VSCE_PAT", "")
	t.Setenv("OVSX_PAT", "")
	t.Setenv("VSCE_AZURE_CREDENTIAL", "true")
	runs := recordRuns(t)
	if _, err := publishExtension(ws, extensionApp()); err != nil {
		t.Fatal(err)
	}
	if len(*runs) != 2 || !strings.Contains((*runs)[1], "--azure-credential") {
		t.Errorf("commands: %v", *runs)
	}
}

func TestPublishNeedsACredentialAndABuild(t *testing.T) {
	t.Setenv("OVSX_PAT", "")
	t.Setenv("VSCE_AZURE_CREDENTIAL", "")
	t.Setenv("VSCE_PAT", "")
	runs := recordRuns(t)
	if _, err := publishExtension(extensionWorkspace(t, true), extensionApp()); err == nil {
		t.Error("no credential is an error")
	}
	t.Setenv("VSCE_PAT", "token")
	if _, err := publishExtension(extensionWorkspace(t, false), extensionApp()); err == nil || !strings.Contains(err.Error(), "build") {
		t.Errorf("an unbuilt extension: %v", err)
	}
	if len(*runs) != 0 {
		t.Errorf("ran %v", *runs)
	}
}

func TestPublishLeavesOtherProjectsAndDryRunsAlone(t *testing.T) {
	ws := extensionWorkspace(t, true)
	runs := recordRuns(t)
	pkg := extensionApp()
	pkg.Project.Kind = protocol.ProjectPackage
	if result, err := publishExtension(ws, pkg); err != nil || result.Published || result.Skipped != "" || result.Registry != "" {
		t.Errorf("a package: %v %v", result, err)
	}
	dry := extensionApp()
	dry.DryRun = true
	if result, err := publishExtension(ws, dry); err != nil || result.Skipped != "dry run" || result.Published {
		t.Errorf("dry run: %v %v", result, err)
	}
	if len(*runs) != 0 {
		t.Errorf("ran %v", *runs)
	}
}

func TestOnlyAnExtensionAppIsContributedAndNeedsTokens(t *testing.T) {
	ws := extensionWorkspace(t, true)
	app := protocol.Project{Name: "ext", Root: "apps/ext", Kind: protocol.ProjectApp}
	pkg := protocol.Project{Name: "ext", Root: "apps/ext", Kind: protocol.ProjectPackage}
	if got := contributeExtensions(ws, protocol.ContributeParams{Projects: []protocol.Project{app}}).Projects["ext"]; len(got.DependsOn["package"]) != 1 {
		t.Errorf("contribution %+v", got)
	}
	if got := contributeExtensions(ws, protocol.ContributeParams{Projects: []protocol.Project{pkg}}); len(got.Projects) != 0 {
		t.Errorf("a package is contributed: %v", got)
	}
	creds := ciExtensions(ws, protocol.CIParams{Projects: []protocol.Project{app}}).Credentials
	if len(creds) != 2 || creds[0].Env != "VSCE_PAT" || !creds[0].Required || creds[1].Env != "OVSX_PAT" || creds[1].Required {
		t.Errorf("credentials %v", creds)
	}
	if got := ciExtensions(ws, protocol.CIParams{Projects: []protocol.Project{pkg}}).Credentials; len(got) != 0 {
		t.Errorf("a package needs tokens: %v", got)
	}
}

func TestANativeExtensionIsPublishedForEachPlatform(t *testing.T) {
	ws := extensionWorkspace(t, true)
	t.Setenv("VSCE_PAT", "token")
	t.Setenv("OVSX_PAT", "")
	t.Setenv("VSCE_AZURE_CREDENTIAL", "")
	runs := recordRuns(t)
	native := extensionApp()
	native.Project.Tags = []string{"native"}
	if result, err := publishExtension(ws, native); err != nil || result.Published || !strings.Contains(result.Skipped, "native jobs") || len(*runs) != 0 {
		t.Fatalf("the release job: %v %v %v", result, err, *runs)
	}
	native.Platform = "linux/arm64"
	result, err := publishExtension(ws, native)
	if err != nil || !result.Published {
		t.Fatalf("published %v: %v", result, err)
	}
	if len(*runs) != 2 || !strings.Contains((*runs)[0], "--target linux-arm64") || !strings.Contains((*runs)[1], "ext-linux-arm64.vsix") {
		t.Errorf("commands: %v", *runs)
	}
}
