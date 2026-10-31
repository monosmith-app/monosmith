package derive

import (
	"reflect"

	"github.com/monosmith-app/monosmith/apps/smith/derivelock"
	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Sync records and writes the derived files that need no decision, as
// `smith derive` does, without a prompt or a report. A new project calls it, so
// the files the scaffold wrote are in smith.lock from the first commit. It does
// nothing when [derive] is off. A file that needs a choice stays as it is, for
// `smith derive` to report. It returns the paths it wrote or recorded, and an
// error when a plugin fails or a write fails.
func Sync(s *session.Session) ([]string, error) {
	if !s.Config.Derive.Enabled {
		return nil, nil
	}
	plan, rec, err := BuildPlan(s)
	if err != nil {
		return nil, err
	}
	before := append([]derivelock.Entry(nil), rec.Files...)
	out, err := Apply(s.Root, plan, &rec, Decision{})
	if err != nil {
		return nil, err
	}
	if !reflect.DeepEqual(before, rec.Files) {
		if err := derivelock.Write(s.Root, rec); err != nil {
			return nil, err
		}
	}
	return append(out.Written, out.Recorded...), nil
}
