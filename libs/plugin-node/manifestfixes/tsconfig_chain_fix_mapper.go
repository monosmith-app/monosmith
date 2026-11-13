package manifestfixes

import (
	"path"

	"github.com/monosmith-app/monosmith/libs/plugin-node/tsconfigchain"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/lineendings"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TsconfigChainDescription names the fix on every project it applies to, so the
// host folds them into one.
const TsconfigChainDescription = "extend tsconfig.base.json in tsconfig.json"

// TsconfigChainFix is the fix for a project whose tsconfig.json does not reach
// the workspace's tsconfig.base.json. It asks first: extend the base (the edit
// of the "extends" member, shown as a diff), or leave the file as it is. It
// returns nil when the file cannot be read or edited.
func TsconfigChainFix(workspaceRoot string, p protocol.Project) *protocol.Fix {
	file := path.Join(p.Root, "tsconfig.json")
	raw, err := readFile(workspaceRoot, file)
	if err != nil {
		return nil
	}
	edited, err := tsconfigchain.Rewire(raw, p.Root)
	if err != nil {
		return nil
	}
	return &protocol.Fix{
		Description: TsconfigChainDescription,
		AppliesTo:   []string{p.Name},
		Choices: []protocol.FixChoice{{
			Question: p.Name + ": " + file + " does not extend tsconfig.base.json, so a change of the base never reaches it. Extending sets \"extends\" and keeps every option the file has (its own values still win). The next screen shows the changed line.",
			Options: []protocol.FixChoiceOption{
				{Label: "extend " + tsconfigchain.Extends(p.Root, tsconfigchain.Base), Fix: protocol.Fix{Files: []protocol.FileEdit{{Path: file, Content: lineendings.Match(string(raw), string(edited))}}}},
				{Label: "leave " + file + " as it is (nothing changes)"},
			},
		}},
	}
}
