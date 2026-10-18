package registries

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// FileName is the NuGet configuration smith writes at the workspace root.
const FileName = "nuget.config"

// CheckNuGetConfig checks that the workspace's nuget.config lists the feeds
// smith.toml names. smith owns a nuget.config it wrote (it carries Marker),
// so rewriting that one is a safe fix; a nuget.config of the team's gets the
// same content as a fix to apply by hand. ok is false when there is nothing
// to check.
func CheckNuGetConfig(workspaceRoot string, projects []protocol.Project) (protocol.Finding, bool) {
	const check = "nuget.config lists the feeds smith.toml names"
	raw, err := os.ReadFile(filepath.Join(workspaceRoot, FileName))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return protocol.Finding{Check: check, Detail: err.Error(), Remedy: "make nuget.config readable"}, true
	}
	existing, ours := string(raw), err != nil || strings.Contains(string(raw), Marker)
	want := Config(Of(projects))
	if want == "" {
		// No feed named: a nuget.config smith wrote is stale; the team's is
		// none of smith's business.
		if err == nil && ours {
			return protocol.Finding{Check: check, Detail: "nuget.config was written for [registries] smith.toml no longer names", Remedy: "run `smith doctor --fix`",
				Fix: &protocol.Fix{Description: "remove smith's nuget.config", Safe: true, Files: []protocol.FileEdit{{Path: FileName, Delete: true}}}}, true
		}
		return protocol.Finding{}, false
	}
	f := protocol.Finding{Check: check, OK: existing == want}
	if f.OK {
		return f, true
	}
	if ours {
		f.Detail = "nuget.config does not match [registries] in smith.toml"
		f.Remedy = "run `smith doctor --fix`"
		f.Fix = &protocol.Fix{Description: "write the feeds into nuget.config", Safe: true, Files: []protocol.FileEdit{{Path: FileName, Content: want}}}
		return f, true
	}
	f.Detail = "nuget.config is the team's own (it lacks smith's marker), so smith does not rewrite it"
	f.Remedy = "add the feeds from smith.toml [registries] to it, or delete it and run `smith doctor --fix` to have smith write it"
	f.Fix = &protocol.Fix{Description: "replace nuget.config with smith's", Safe: false, Files: []protocol.FileEdit{{Path: FileName, Content: want}}}
	return f, true
}
