package convertplan

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"path"
	"strings"
)

// DropDir is where smith ci pack collects deliverables.
const DropDir = "dist/drop"

// ownDropDir gives a package task its own directory under dist/drop when it
// declares the shared dist/drop as its output. Every package task of the old
// workspace wrote its zip into dist/drop, and a cache restore replaces the
// output directories of the task it restores, deleting what another package
// task had just written there. The command's dist/drop paths and the outputs
// move to dist/drop/<project dir>; it reports that directory, and whether
// anything changed.
func ownDropDir(projectRoot string, o *workspace.TaskOverride) (string, bool) {
	shared := -1
	for i, out := range o.Outputs {
		if path.Clean(path.Join(projectRoot, out)) == DropDir {
			shared = i
		}
	}
	if shared < 0 {
		return "", false
	}
	own := DropDir + "/" + path.Base(projectRoot)
	o.Outputs[shared] = relativeTo(projectRoot, own)
	for i, arg := range o.Command {
		o.Command[i] = moveDropPaths(arg, own)
	}
	return own, true
}

// moveDropPaths rewrites each dist/drop in text to own, unless it names a
// longer path (dist/drop-old) or is already below own.
func moveDropPaths(text, own string) string {
	var b strings.Builder
	for {
		i := strings.Index(text, DropDir)
		if i < 0 {
			b.WriteString(text)
			return b.String()
		}
		end := i + len(DropDir)
		rest := text[end:]
		b.WriteString(text[:i])
		switch {
		case strings.HasPrefix(rest, own[len(DropDir):]+"/") || rest == own[len(DropDir):]:
			b.WriteString(DropDir)
		case rest != "" && strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-.", rune(rest[0])):
			b.WriteString(DropDir)
		default:
			b.WriteString(own)
		}
		text = rest
	}
}
