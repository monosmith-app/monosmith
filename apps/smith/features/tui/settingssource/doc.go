// Package settingssource is the Settings tab's backend: it reads the workspace
// and makes every change by calling what the matching `smith` command calls
// (plugin enable|disable, plugin trust|untrust, config set|unset) in this
// process, so the tab leaves smith.toml byte for byte as the command would.
package settingssource
