package pluginarchive

import (
	"fmt"
	"regexp"
)

var (
	pluginNamePattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	pluginVersionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,63}$`)
)

// ValidatePluginName accepts the names smith documents for plugins:
// lowercase words joined by dashes. The name becomes a directory name when a
// plugin is installed, so nothing else is accepted.
func ValidatePluginName(name string) error {
	if !pluginNamePattern.MatchString(name) {
		return fmt.Errorf("plugin name %q must be lowercase letters, digits and dashes, starting with a letter or digit", name)
	}
	return nil
}

// ValidatePluginVersion accepts a version that is safe in a file name.
func ValidatePluginVersion(version string) error {
	if !pluginVersionPattern.MatchString(version) {
		return fmt.Errorf("plugin version %q must be letters, digits, dots, dashes, plus signs and underscores (it is part of the archive's file name)", version)
	}
	return nil
}
