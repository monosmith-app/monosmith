package pipeline

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestCIInstallsTheTargetsAndLinkersOfCrossBuiltReleases(t *testing.T) {
	spec := &protocol.AssetSpec{Platforms: []string{"linux/amd64", "linux/arm64", "windows/amd64", "darwin/arm64"}}
	app := protocol.Project{Name: "cli", Root: "apps/cli", ReleaseAssets: spec}
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{app}})
	inputs := got.Toolchains[0].GitHub.Inputs
	if last := inputs[len(inputs)-1]; last.Name != "targets" || last.Value != "aarch64-unknown-linux-gnu" {
		t.Fatalf("only what the Linux runner can build is installed: %+v", inputs)
	}
	if !strings.Contains(got.Toolchains[0].Install, "--target aarch64-unknown-linux-gnu") {
		t.Errorf("rustup installs it elsewhere: %s", got.Toolchains[0].Install)
	}
	if len(got.Setup) != 2 || !strings.Contains(got.Setup[1].Command[2], "gcc-aarch64-linux-gnu") || got.Setup[1].Blocking {
		t.Errorf("the linker: %+v", got.Setup)
	}
}

func TestCIOfANativeAppAddsNoTargets(t *testing.T) {
	spec := &protocol.AssetSpec{Platforms: []string{"linux/arm64", "windows/amd64"}}
	app := protocol.Project{Name: "cli", Root: "apps/cli", Tags: []string{"native"}, ReleaseAssets: spec}
	got := CI(t.TempDir(), protocol.CIParams{Projects: []protocol.Project{app}})
	if len(got.Setup) != 1 || len(got.Toolchains[0].GitHub.Inputs) != 2 {
		t.Errorf("a native app is built on its own runner: %+v", got)
	}
}
