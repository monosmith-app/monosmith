package sources

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/doctor"
	"github.com/monosmith-app/monosmith/apps/smith/doctorfix"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ConfigRepair is the fix `smith doctor --fix` makes to a smith.toml that
// uses a renamed key, with what the fix wizard needs to show and apply it.
type ConfigRepair struct {
	// Fix is the planned rename: the same one doctor plans.
	Fix doctorfix.Planned
	// Show writes the review of Fix: its diff against the file on disk.
	Show func(doctorfix.Planned) (string, error)
	// Apply writes the file. It runs only after the user confirms.
	Apply func(ctx context.Context, fix protocol.Fix, out io.Writer) error
}

// ConfigRepairOf plans the repair of the smith.toml found from dir. It only
// reads: nothing is written. It returns false when the file cannot be read or
// has a problem doctor cannot fix.
func ConfigRepairOf(dir, version string) (ConfigRepair, bool) {
	root, err := workspace.FindRoot(dir)
	if err != nil {
		return ConfigRepair{}, false
	}
	raw, err := os.ReadFile(filepath.Join(root, workspace.FileName))
	if err != nil {
		return ConfigRepair{}, false
	}
	finding, ok := doctor.RemovedKeysFinding(raw)
	if !ok || finding.Fix == nil {
		return ConfigRepair{}, false
	}
	planned := doctorfix.Group([]doctorfix.Item{{Check: finding.Check, Fix: finding.Fix}})
	if len(planned) != 1 {
		return ConfigRepair{}, false
	}
	s := WorkspaceSources(dir, version)
	return ConfigRepair{Fix: planned[0], Show: s.ShowFix, Apply: s.ApplyFix}, true
}
