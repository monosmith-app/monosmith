package pluginhost

import (
	"fmt"
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// Toggle returns the enabled list after enabling or disabling names. An empty
// current list means every built-in plugin.
func Toggle(current, names []string, enable bool) ([]string, error) {
	if len(current) == 0 {
		current = BuiltinNames()
	}
	set := map[string]bool{}
	for _, n := range current {
		set[n] = true
	}
	for _, n := range names {
		if _, ok := Builtins[n]; !ok {
			if _, addon := BuiltinAddons[n]; addon {
				return nil, fmt.Errorf("%s is an add-on plugin, not a language: it is always loaded and acts only on projects whose stack names it, so it cannot be disabled. Only the language plugins can: %s. To stop it acting, change the stack of the project (smith stacks) or remove [conventions.%s] from smith.toml", n, strings.Join(BuiltinNames(), ", "), n)
			}
			return nil, fmt.Errorf("no built-in plugin %q (built-in: %s)", n, strings.Join(BuiltinNames(), ", "))
		}
		set[n] = enable
	}
	var out []string
	for _, n := range BuiltinNames() {
		if set[n] {
			out = append(out, n)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one plugin must stay enabled")
	}
	return out, nil
}

// EnablePlugins turns built-in plugins on or off in the smith.toml at root:
// what `smith plugin enable|disable` does. It returns the list now enabled;
// the file changes only on its `enabled = [...]` line, and an error (an
// unknown plugin, the last one switched off) leaves it alone.
func EnablePlugins(root string, names []string, enable bool) ([]string, error) {
	cfg, err := workspace.Load(root)
	if err != nil {
		return nil, err
	}
	enabled, err := Toggle(cfg.Plugins.Enabled, names, enable)
	if err != nil {
		return nil, err
	}
	// A list that names every built-in plugin would freeze the set: a plugin a
	// later smith adds would stay off. Only a switched-off plugin is written.
	if err := writeEnabled(root, enabled); err != nil {
		return nil, err
	}
	return enabled, nil
}

// writeEnabled stores the enabled list: the explicit list only when a plugin is
// off, and no list at all when every built-in plugin is on.
func writeEnabled(root string, enabled []string) error {
	if slices.Equal(enabled, BuiltinNames()) {
		return workspace.ClearEnabledPlugins(root)
	}
	return workspace.SetEnabledPlugins(root, enabled)
}

// SwitchOffWarning is what to tell a person who switched off built-in plugins:
// what they detect disappears from smith until they are switched on again.
// It returns "" when names is empty.
func SwitchOffWarning(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return "switched off " + strings.Join(names, ", ") + ": the projects, tasks and checks it detects are hidden from smith until you switch it on again"
}
