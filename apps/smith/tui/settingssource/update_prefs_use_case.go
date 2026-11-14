package settingssource

import (
	"context"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/updatenotice"
)

// UpdateInfoOf is the update as the Update section shows it.
//
// Parameters: u is what updatenotice knows.
// Returns: the same facts in the tab's own type.
func UpdateInfoOf(u updatenotice.Update) settingsview.UpdateInfo {
	return settingsview.UpdateInfo{
		Current: u.Current, Latest: u.Latest, Command: u.Command,
		Available: u.Available, Runnable: u.Runnable(), Muted: u.Muted,
	}
}

// updatePrefs is the Update section's access to smith update: what is known
// now, a check on request, and the reminder switch.
func updatePrefs(version string) settingsview.UpdatePrefs {
	o := updatenotice.ForThisProcess(version, 0)
	known, _ := updatenotice.Available(o)
	return settingsview.UpdatePrefs{
		Info: UpdateInfoOf(known),
		Check: func() (settingsview.UpdateInfo, error) {
			u, err := updatenotice.Check(context.Background(), o, 10*time.Second)
			if err != nil {
				return settingsview.UpdateInfo{}, err
			}
			return UpdateInfoOf(u), nil
		},
		SetMuted: func(muted bool) error {
			updatenotice.SetMuted(o.Home, muted)
			return nil
		},
	}
}
