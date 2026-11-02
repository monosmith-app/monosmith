package tauri

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAMissingTauriCLIHasASmithFixForTheFirstTauriApp(t *testing.T) {
	ws := t.TempDir()
	write(t, ws, "apps/studio/src-tauri/tauri.conf.json", "{}")
	res := Doctor(ws, protocol.DoctorParams{Projects: []protocol.Project{{Name: "web", Root: "apps/web"}, {Name: "studio", Root: "apps/studio"}}})
	for _, f := range res.Findings {
		if f.Check != "the tauri CLI is installed" {
			continue
		}
		if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 1 || strings.Join(f.Fix.Steps[0].Argv, " ") != "smith install -w studio -D @tauri-apps/cli --yes" {
			t.Fatalf("%+v", f.Fix)
		}
		return
	}
	t.Fatalf("no tauri CLI finding: %+v", res.Findings)
}
