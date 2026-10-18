package registries

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CheckNpmrc checks that the workspace's .npmrc carries smith's block for
// the registries smith.toml names, and offers the fix that writes it. The
// fix only rewrites smith's block, so it is safe. ok is false when there is
// nothing to check: no registry named and no block to remove.
func CheckNpmrc(workspaceRoot string, projects []protocol.Project) (protocol.Finding, bool) {
	path := filepath.Join(workspaceRoot, ".npmrc")
	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return protocol.Finding{Check: ".npmrc points at the registries smith.toml names", Detail: err.Error(), Remedy: "make .npmrc readable"}, true
	}
	existing := string(raw)
	block := Block(Of(projects), Scope(projects))
	want := Merge(existing, block)
	if block == "" && want == existing {
		return protocol.Finding{}, false
	}
	f := protocol.Finding{Check: ".npmrc points at the registries smith.toml names", OK: want == existing}
	if !f.OK {
		f.Detail = "the smith:registry block of .npmrc does not match [registries] in smith.toml"
		f.Remedy = "run `smith doctor --fix`"
		f.Fix = &protocol.Fix{Description: "write the registries into .npmrc", Safe: true, Files: []protocol.FileEdit{{Path: ".npmrc", Content: want}}}
	}
	return f, true
}
