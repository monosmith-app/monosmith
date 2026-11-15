// Package userprefs keeps the person's own interface preferences (the colour
// theme, the symbol set, whether the Settings tab was introduced) in the
// smith home, as tui.json. They are a personal choice, so they never go into
// a workspace's smith.toml; the SMITH_THEME, NO_COLOR and SMITH_ASCII
// environment variables still win over them.
package userprefs
