package doctor

import (
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// defaultsVintageFinding reports which defaults vintage the workspace uses
// against the latest this smith knows (ADR 0054). Being behind is a warning
// with one remedy: a newer smith never moves a workspace by itself.
//
// Parameters: recorded is [workspace] defaults.
func defaultsVintageFinding(recorded string) protocol.Finding {
	latest := workspace.LatestDefaultsVintage()
	f := protocol.Finding{Check: "the workspace uses the latest defaults"}
	switch workspace.DefaultsVintageState(recorded) {
	case workspace.VintageCurrent:
		f.OK = true
	case workspace.VintageAhead:
		f.Warning = true
		f.Detail = "this workspace uses the defaults of " + recorded + ", newer than this smith knows (" + latest + ")"
		f.Remedy = "upgrade smith"
	default:
		f.Warning = true
		if recorded == "" {
			f.Detail = "this workspace records no defaults vintage; the latest is " + latest
		} else {
			f.Detail = "this workspace uses the defaults of " + recorded + "; the latest is " + latest
		}
		f.Remedy = "`smith upgrade --defaults` shows the plan and moves the workspace; until then smith keeps the defaults it has"
	}
	return f
}
