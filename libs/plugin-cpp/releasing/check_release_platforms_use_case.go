package releasing

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/monosmith-app/monosmith/libs/plugin-cpp/discovery"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

var installRule = regexp.MustCompile(`(?i)\binstall\s*\(`)

// CheckReleasePlatforms is doctor's look at an app's release: every platform
// it lists is one the release can name, an app that cannot be built here is
// marked native, and the project has a program to put in the archive (an
// add_executable target, or install rules). A native app is built by the
// runner of each system, so only its platforms' names are checked.
func CheckReleasePlatforms(workspaceRoot string, p protocol.Project) []protocol.Finding {
	if p.ReleaseAssets == nil {
		return nil
	}
	native := slices.Contains(p.Tags, discovery.NativeTag)
	var findings []protocol.Finding
	for _, platform := range p.ReleaseAssets.Platforms {
		if why := CheckPlatform(platform); why != "" {
			findings = append(findings, protocol.Finding{
				Check: p.Name + " releases for " + platform, Detail: why,
				Remedy: "change platforms in [projects." + p.Name + ".release] in smith.toml",
			})
			continue
		}
		if !native && platform != hostPlatform {
			findings = append(findings, protocol.Finding{
				Check:  p.Name + " can build the " + platform + " archive here",
				Detail: "this machine (" + hostPlatform + ") cannot build " + platform + ": a C++ program is built on the system it is for",
				Remedy: "add set(SMITH_NATIVE ON) to " + p.Root + "/CMakeLists.txt: `smith ci native` then builds each archive on a runner of its own system",
			})
		}
	}
	if src, err := os.ReadFile(filepath.Join(workspaceRoot, filepath.FromSlash(p.Root), "CMakeLists.txt")); err == nil {
		lists := discovery.ParseLists(string(src))
		ok := lists.Executable != "" || installRule.Match(src)
		findings = append(findings, protocol.Finding{
			Check: p.Name + " has a program to release", OK: ok,
			Detail: p.Root + "/CMakeLists.txt has no add_executable and no install rule, so the archive would be empty",
			Remedy: "add an add_executable target (the app's program), or install(TARGETS ...) rules",
		})
	}
	return findings
}
