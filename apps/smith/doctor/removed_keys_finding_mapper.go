package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// RemovedKeysFinding reports the tables smith.toml still uses under a name smith
// renamed, with the fix that renames them in the text (comments and the rest of
// the file stay). It returns false when the text uses none. The fix is safe: it
// changes names only, and the diff shows every line it touches.
func RemovedKeysFinding(raw []byte) (protocol.Finding, bool) {
	removed := workspace.FindRemovedKeys(raw)
	if len(removed) == 0 {
		return protocol.Finding{}, false
	}
	err := &workspace.RemovedKeysError{Keys: removed}
	finding := protocol.Finding{
		Check:  "smith.toml uses current key names",
		Detail: err.Error(),
		Remedy: "run `smith doctor --fix`: it renames the keys and keeps your comments",
	}
	if text, ok := workspace.RenameRemovedKeys(raw); ok {
		finding.Fix = &protocol.Fix{
			Description: "rename the keys smith.toml uses to their current names",
			Safe:        true,
			Files:       []protocol.FileEdit{{Path: workspace.FileName, Content: string(text)}},
		}
	}
	return finding, true
}
