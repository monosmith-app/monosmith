package releasing

import (
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-rust/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckReleaseTargets is doctor's look at an app's release: every platform it
// lists has a Rust target, this machine can build each (the target is
// installed, a linker is there where one is needed), and an app that cannot
// be built here is marked native. A native app is built by the runner of each
// system, so only its platforms' names are checked.
func CheckReleaseTargets(workspaceRoot string, p protocol.Project) []protocol.Finding {
	if p.ReleaseAssets == nil {
		return nil
	}
	native := slices.Contains(p.Tags, discovery.NativeTag)
	installed, haveRustup := InstalledTargets(workspaceRoot)
	var findings []protocol.Finding
	for _, platform := range p.ReleaseAssets.Platforms {
		triple, err := Target(platform)
		if err != nil {
			findings = append(findings, protocol.Finding{
				Check: p.Name + " releases for " + platform, Detail: err.Error(),
				Remedy: "change platforms in [projects." + p.Name + ".release] in smith.toml to ones Rust can build for",
			})
			continue
		}
		if native {
			continue
		}
		switch CrossReach(hostPlatform, platform) {
		case OnItsOwnSystem:
			findings = append(findings, protocol.Finding{
				Check:  p.Name + " can build the " + platform + " archive here",
				Detail: "this machine (" + hostPlatform + ") has no linker or system libraries for " + platform + ": a Rust program is built on the system it is for",
				Remedy: "add native = true under [package.metadata.smith] in " + p.Root + "/Cargo.toml: `smith ci native` then builds each archive on a runner of its own system",
			})
			continue
		case WithLinker:
			program, pkg := Linker(platform)
			ok := LinkerAvailable(workspaceRoot, platform, triple)
			findings = append(findings, protocol.Finding{
				Check: p.Name + " has a linker for " + platform, OK: ok,
				Detail: program + " is not on PATH and " + LinkerEnv(triple) + " is not set, so cargo cannot link a " + platform + " program on " + hostPlatform,
				Remedy: "install it (apt install " + pkg + "), or set " + LinkerEnv(triple) + " to a linker, or mark the app native",
			})
		}
		if !haveRustup {
			continue
		}
		ok := slices.Contains(installed, triple)
		f := protocol.Finding{
			Check: "the Rust target " + triple + " is installed for " + p.Name, OK: ok,
			Detail: "the " + platform + " archive is built with --target " + triple + ", which rustup has not installed",
			Remedy: "run: rustup target add " + triple,
		}
		if !ok {
			f.Fix = &protocol.Fix{
				Description: "install the Rust target for a release platform",
				Steps:       []protocol.FixStep{{Description: "rustup target add " + triple, Argv: []string{"rustup", "target", "add", triple}}},
				AppliesTo:   []string{p.Name},
			}
		}
		findings = append(findings, f)
	}
	return findings
}
