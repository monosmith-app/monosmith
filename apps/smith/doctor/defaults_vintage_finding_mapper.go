package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/smithfix"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// defaultsVintageFinding reports which defaults vintage the workspace uses
// against the latest this smith knows (ADR 0054). Being behind is a warning
// with one remedy: a newer smith never moves a workspace by itself.
//
// A workspace that records no vintage follows the defaults of today's vintage.
// Recording it keeps them exactly as they are, so doctor offers that as a fix
// that writes one line in smith.toml, next to the move to the latest (which
// shows a plan).
//
// Parameters: recorded is [workspace] defaults; tomlSrc is the text of
// smith.toml, from which the record fix is made ("" offers no record fix).
func defaultsVintageFinding(recorded, tomlSrc string) protocol.Finding {
	latest := workspace.LatestDefaultsVintage()
	f := protocol.Finding{Check: "the workspace uses the latest defaults", Area: AreaConfig}
	switch workspace.DefaultsVintageState(recorded) {
	case workspace.VintageCurrent:
		f.OK = true
	case workspace.VintageAhead:
		f.Warning = true
		f.Problem = "the workspace uses the defaults of " + recorded + ", newer than this smith knows (" + latest + ")"
		f.Detail = "this workspace uses the defaults of " + recorded + ", newer than this smith knows (" + latest + ")"
		f.Remedy = "upgrade smith"
	default:
		f.Warning = true
		if recorded == "" {
			f.Problem = "the workspace does not record which defaults it uses"
			f.Detail = "this workspace follows the defaults of " + latest + " but does not record it. Recording the vintage keeps them exactly as they are and changes nothing else"
		} else {
			f.Problem = "the workspace uses the defaults of " + recorded + ", the latest is " + latest
			f.Detail = "this workspace uses the defaults of " + recorded + "; the latest is " + latest
		}
		f.Remedy = "`smith upgrade --defaults` shows the plan and moves the workspace; until then smith keeps the defaults it has"
		move := smithfix.Step("Move the workspace to the "+latest+" defaults; the plan above lists each change", "upgrade", "--defaults")
		f.Fix = smithfix.Fix("Move the workspace to the "+latest+" defaults", move)
		if recorded == "" && tomlSrc != "" {
			f.Remedy = "choose: keep the current defaults (record " + latest + " in smith.toml, nothing else changes) or move to the latest (`smith upgrade --defaults` shows a plan)"
			record := protocol.Fix{Description: "keep the current defaults", Files: []protocol.FileEdit{{
				Path: workspace.FileName, Content: configedit.SetKey(tomlSrc, "workspace", "defaults", configedit.Quote(latest)),
			}}}
			f.Fix = &protocol.Fix{
				Description: "Record the defaults vintage of the workspace",
				Choices: []protocol.FixChoice{{
					Question: "This workspace follows the defaults of " + latest + ". What do you want?",
					Options: []protocol.FixChoiceOption{
						{Label: "keep the current defaults (record " + latest + " in smith.toml; changes nothing else)", Fix: record},
						{Label: "move to the latest defaults (shows a plan first)", Fix: protocol.Fix{Description: "Move the workspace to the " + latest + " defaults", Steps: []protocol.FixStep{move}}},
					},
				}},
			}
		}
	}
	return f
}
