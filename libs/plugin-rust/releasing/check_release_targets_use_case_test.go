package releasing

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func findingFor(findings []protocol.Finding, check string) *protocol.Finding {
	for i := range findings {
		if strings.Contains(findings[i].Check, check) {
			return &findings[i]
		}
	}
	return nil
}

func TestDoctorFlagsAMissingTargetAndOffersToInstallIt(t *testing.T) {
	fakeToolchain(t, "darwin/arm64", "aarch64-apple-darwin")
	p := protocol.Project{Name: "cli", Root: "apps/cli", ReleaseAssets: spec("darwin/arm64", "darwin/amd64")}
	got := CheckReleaseTargets(t.TempDir(), p)
	missing := findingFor(got, "x86_64-apple-darwin")
	if missing == nil || missing.OK || missing.Remedy != "run: rustup target add x86_64-apple-darwin" || missing.Fix == nil {
		t.Fatalf("%+v", got)
	}
	if have := findingFor(got, "aarch64-apple-darwin"); have == nil || !have.OK {
		t.Errorf("an installed target passes: %+v", got)
	}
}

func TestDoctorFlagsAMissingLinkerAndAnotherSystem(t *testing.T) {
	fakeToolchain(t, "linux/amd64", "x86_64-unknown-linux-gnu", "aarch64-unknown-linux-gnu")
	p := protocol.Project{Name: "cli", Root: "apps/cli", ReleaseAssets: spec("linux/arm64", "windows/amd64")}
	got := CheckReleaseTargets(t.TempDir(), p)
	linker := findingFor(got, "linker for linux/arm64")
	if linker == nil || linker.OK || !strings.Contains(linker.Remedy, "apt install gcc-aarch64-linux-gnu") {
		t.Fatalf("%+v", got)
	}
	system := findingFor(got, "windows/amd64 archive here")
	if system == nil || system.OK || !strings.Contains(system.Remedy, "native = true") || !strings.Contains(system.Remedy, "apps/cli/Cargo.toml") {
		t.Fatalf("%+v", got)
	}
}

func TestDoctorOfANativeAppChecksOnlyTheNames(t *testing.T) {
	fakeToolchain(t, "linux/amd64")
	p := protocol.Project{Name: "cli", Root: "apps/cli", Tags: []string{"native"}, ReleaseAssets: spec("windows/arm64", "darwin/amd64", "beos/ppc")}
	got := CheckReleaseTargets(t.TempDir(), p)
	if len(got) != 1 || got[0].OK || !strings.Contains(got[0].Check, "beos/ppc") {
		t.Fatalf("%+v", got)
	}
}

func TestDoctorSaysNothingForAnAppWithoutRelease(t *testing.T) {
	if got := CheckReleaseTargets(t.TempDir(), protocol.Project{Name: "cli"}); got != nil {
		t.Fatalf("%+v", got)
	}
}
