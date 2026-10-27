package cipipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func installModel(version string, tools ...string) Model {
	var c Contributions
	for _, name := range tools {
		c.Add(protocol.CIResult{Toolchains: []protocol.Toolchain{{Name: name, Version: "1", Install: "install-" + name}}})
	}
	return NewModel(c, version, "main")
}

func TestSmithInstallPicksNodeThenPythonThenDotnetElseTheScript(t *testing.T) {
	for _, tc := range []struct {
		tools  []string
		script bool
		hosted string
	}{
		{[]string{"go", "node", "python"}, false, "npm install --global --registry https://registry.npmjs.org/ monosmith@1.5.0"},
		{[]string{"python", "dotnet"}, false, "python -m pip install --disable-pip-version-check --index-url https://pypi.org/simple/ monosmith==1.5.0"},
		{[]string{"dotnet", "rust"}, false, "dotnet tool update --global monosmith --version 1.5.0"},
		{[]string{"go", "rust", "java"}, true, ""},
		{nil, true, ""},
	} {
		got := SmithInstallFor(installModel("1.5.0", tc.tools...))
		if got.Script != tc.script || got.HostedCommand() != tc.hosted {
			t.Errorf("%v: script %v command %q", tc.tools, got.Script, got.HostedCommand())
		}
	}
}

func TestSmithInstallFromARegistryIsUnpinnedForADevBuild(t *testing.T) {
	for tool, want := range map[string]string{"node": "monosmith\n", "python": "monosmith\n", "dotnet": "monosmith\n"} {
		got := SmithInstallFor(installModel("latest", tool)).HostedCommand() + "\n"
		if !strings.HasSuffix(got, want) || strings.ContainsAny(got, "@=") || strings.Contains(got, "--version") {
			t.Errorf("%s: %q", tool, got)
		}
	}
}

func TestPipDoesNotInstallPerUserOnHostedAgents(t *testing.T) {
	i := SmithInstallFor(installModel("1.5.0", "python"))
	if strings.Contains(i.HostedCommand(), "--user") || !strings.Contains(i.ContainerCommand(), "python3 -m pip install") || !strings.Contains(i.ContainerCommand(), "--user") {
		t.Fatalf("hosted %q container %q", i.HostedCommand(), i.ContainerCommand())
	}
}

func TestEveryProviderPutsTheDotnetToolOnPath(t *testing.T) {
	m := installModel("1.5.0", "dotnet")
	gh := RenderGitHub(m)
	if !strings.Contains(gh, "dotnet tool update --global monosmith --version 1.5.0\n          echo \"${USERPROFILE:-$HOME}/.dotnet/tools\" >> \"$GITHUB_PATH\"") {
		t.Errorf("github:\n%s", gh)
	}
	az := RenderAzure(m)
	if !strings.Contains(az, "echo \"##vso[task.prependpath]${USERPROFILE:-$HOME}/.dotnet/tools\"") {
		t.Errorf("azure:\n%s", az)
	}
	if gl := RenderGitLab(m); !strings.Contains(gl, "    - 'dotnet tool update --global monosmith --version 1.5.0'\n    - 'export PATH=\"$HOME/.dotnet/tools:$PATH\"'\n") {
		t.Errorf("gitlab:\n%s", gl)
	}
	if cc := RenderCircleCI(m); !strings.Contains(cc, "echo 'export PATH=\"$HOME/.dotnet/tools:$PATH\"' >> \"$BASH_ENV\"") {
		t.Errorf("circleci:\n%s", cc)
	}
	if dc := RenderDevcontainer("x", m); !strings.Contains(dc, `PATH=\"$HOME/.dotnet/tools\":\"$PATH\" smith ci setup`) {
		t.Errorf("devcontainer:\n%s", dc)
	}
}

func TestARegistryInstallWritesNoScriptAndPrunesAGeneratedOne(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	script := filepath.Join(root, "tools", "smith", "install.sh")
	if _, _, err := Write(root, cfg, scriptModel()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("a Go workspace keeps the script: %v", err)
	}
	node := model()
	if drift := Drift(root, cfg, node); len(drift) != 2 || !strings.Contains(strings.Join(drift, ";"), InstallScriptPath+" is no longer configured") {
		t.Fatalf("drift %v", drift)
	}
	changed, _, err := Write(root, cfg, node)
	if err != nil || !strings.Contains(strings.Join(changed, ";"), InstallScriptPath) {
		t.Fatalf("changed %v err %v", changed, err)
	}
	if _, err := os.Stat(script); !os.IsNotExist(err) {
		t.Fatalf("the generated script should be removed: %v", err)
	}
	if drift := Drift(root, cfg, node); len(drift) != 0 {
		t.Fatalf("drift %v", drift)
	}
}

func TestATeamsOwnInstallScriptIsKept(t *testing.T) {
	root := t.TempDir()
	cfg := workspace.Config{CI: workspace.CISection{Providers: []string{"github"}}}
	script := filepath.Join(root, "tools", "smith", "install.sh")
	if err := os.MkdirAll(filepath.Dir(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script, []byte("#!/bin/sh\necho mine\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Write(root, cfg, model()); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(script); err != nil {
		t.Fatalf("a script without smith's header is the team's: %v", err)
	}
}
