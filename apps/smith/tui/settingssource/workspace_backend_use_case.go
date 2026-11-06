package settingssource

import (
	"bytes"

	"github.com/monosmith-app/monosmith/apps/smith/architecture"
	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginadmit"
	"github.com/monosmith-app/monosmith/apps/smith/pluginhost"
	"github.com/monosmith-app/monosmith/apps/smith/pluginusage"
	"github.com/monosmith-app/monosmith/apps/smith/tui/settingsview"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
	"github.com/monosmith-app/monosmith/apps/smith/tui/userprefs"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// New is the backend for the workspace that holds dir. home is the smith home
// (SMITH_HOME, or ~/.smith) where the person's own preferences are kept;
// getenv reads the variables that win over them.
func New(dir, home, version string, getenv func(string) string) settingsview.Backend {
	root := func() (string, error) { return workspace.FindRootFrom(dir) }
	prefs, _ := userprefs.Load(home)
	b := settingsview.Backend{
		Load: func() settingsview.Snapshot {
			r, err := root()
			if err != nil {
				return settingsview.Snapshot{Err: err}
			}
			return ReadSnapshot(r, version)
		},
		SetPlugin: func(name string, enable bool) (string, error) {
			r, err := root()
			if err != nil {
				return "", err
			}
			if _, err = pluginhost.EnablePlugins(r, []string{name}, enable); err != nil {
				return "", err
			}
			if !enable {
				warning := pluginhost.SwitchOffWarning([]string{name})
				// The same count `smith plugin disable` shows.
				if used, err := pluginusage.Projects(r, version, []string{name}); err == nil {
					if count := pluginusage.Warning(used, []string{name}); count != "" {
						warning = words.Say("settingssource.disable.hides", "count", count, "warning", warning)
					}
				}
				return warning, nil
			}
			return "", nil
		},
		SetArchitecture: func(name string) error {
			r, err := root()
			if err != nil {
				return err
			}
			return architecture.Set(r, name)
		},
		Review: func(name string) (settingsview.TrustReview, error) {
			r, cfg, err := load(root)
			if err != nil {
				return settingsview.TrustReview{}, err
			}
			review, err := pluginhost.ReviewForTrust(r, cfg, name)
			if err != nil {
				return settingsview.TrustReview{}, err
			}
			var text bytes.Buffer
			pluginadmit.Review(&pluginadmit.Prompter{Out: &text}, review.Status)
			return settingsview.TrustReview{Text: text.String(), Trusted: review.AlreadyTrusted}, nil
		},
		Trust: func(name string, pin bool) (string, error) {
			r, cfg, err := load(root)
			if err != nil {
				return "", err
			}
			review, err := pluginhost.ReviewForTrust(r, cfg, name)
			if err != nil {
				return "", err
			}
			_, digest, err := pluginhost.RecordTrust(r, review, pin)
			if err != nil {
				return "", err
			}
			note := words.Say("settingssource.trust.done", "name", name)
			if pin {
				note = words.Say("settingssource.trust.pinned", "note", note, "digest", digest)
			}
			return note, nil
		},
		Untrust: func(name string) (string, error) {
			r, err := root()
			if err != nil {
				return "", err
			}
			removed, err := pluginhost.ForgetTrust(r, name)
			if err != nil {
				return "", err
			}
			if !removed {
				return words.Say("settingssource.untrust.none", "name", name), nil
			}
			return words.Say("settingssource.untrust.done", "name", name), nil
		},
		SetSetting: func(path string, values []string) (string, error) {
			r, err := root()
			if err != nil {
				return "", err
			}
			if err := configedit.Set(r, path, values); err != nil {
				return "", err
			}
			value, _, _ := configedit.Get(r, path)
			return words.Say("settingssource.setting.set", "path", path, "value", value), nil
		},
		UnsetSetting: func(path string) (string, error) {
			r, err := root()
			if err != nil {
				return "", err
			}
			if err := configedit.Unset(r, path); err != nil {
				return "", err
			}
			return words.Say("settingssource.setting.unset", "path", path), nil
		},
		Style: settingsview.StylePrefs{
			Theme: prefs.Theme, Icons: prefs.Icons, Path: userprefs.Path(home), Getenv: getenv, Dark: theme.DarkBackground(getenv),
			Save: func(themeName, icons string) error {
				p, _ := userprefs.Load(home)
				p.Theme, p.Icons = themeName, icons
				return userprefs.Save(home, p)
			},
		},
		TipSeen: prefs.SettingsTipSeen,
		MarkTipSeen: func() {
			p, _ := userprefs.Load(home)
			p.SettingsTipSeen = true
			_ = userprefs.Save(home, p)
		},
	}
	return b
}

func load(root func() (string, error)) (string, workspace.Config, error) {
	r, err := root()
	if err != nil {
		return "", workspace.Config{}, err
	}
	cfg, err := workspace.Load(r)
	return r, cfg, err
}
