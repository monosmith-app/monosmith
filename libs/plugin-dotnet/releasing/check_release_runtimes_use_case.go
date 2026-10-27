package releasing

import (
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckReleaseRuntimes is doctor's look at an app's release: every platform it
// lists has a .NET runtime identifier, and the runtime packs the publish needs
// are in the NuGet cache. A pack that is not cached is a warning, not an
// error: the first publish downloads it, which only fails without network.
func CheckReleaseRuntimes(p protocol.Project) []protocol.Finding {
	if p.ReleaseAssets == nil {
		return nil
	}
	var findings []protocol.Finding
	for _, platform := range p.ReleaseAssets.Platforms {
		rid, err := RuntimeIdentifier(platform)
		if err != nil {
			findings = append(findings, protocol.Finding{
				Check: p.Name + " releases for " + platform, Detail: err.Error(),
				Remedy: "change platforms in [projects." + p.Name + ".release] in smith.toml to ones .NET can publish for",
			})
			continue
		}
		cached := RuntimePackCached(rid)
		findings = append(findings, protocol.Finding{
			Check: "the .NET runtime for " + rid + " is downloaded for " + p.Name, OK: cached, Warning: !cached,
			Detail: "the " + platform + " archive carries the " + rid + " runtime, which is not in the NuGet cache yet: the release downloads it from NuGet, so it needs network",
			Remedy: "run: dotnet publish -r " + rid + " --self-contained in " + p.Root + " once with network, or restore the NuGet cache in CI",
		})
	}
	return findings
}
