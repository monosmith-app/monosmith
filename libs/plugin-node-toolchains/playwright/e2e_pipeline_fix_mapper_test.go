package playwright

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAMissingE2EPipelineHasASmithFix(t *testing.T) {
	onPlatform(t, "darwin", false)
	realLook := lookPath
	t.Cleanup(func() { lookPath = realLook })
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	ws := t.TempDir()
	writeProject(t, ws, "packages/studio", `{"dependencies": {"playwright": "^1.63.0"}}`)
	projects := []protocol.Project{{Name: "studio", Root: "packages/studio", Tasks: map[string]protocol.Task{TaskName: e2eScript}}}
	f := finding(t, Doctor(ws, protocol.DoctorParams{Projects: projects}), "a pipeline runs the e2e tasks")
	if f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 1 || strings.Join(f.Fix.Steps[0].Argv, " ") != "smith pipeline e2e" {
		t.Fatalf("the remedy names smith pipeline e2e, so the fix runs it: %+v", f.Fix)
	}
}
