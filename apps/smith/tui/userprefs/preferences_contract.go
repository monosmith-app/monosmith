package userprefs

// Preferences are what one person chose for the interface.
type Preferences struct {
	// Theme is dark, light or plain; "" (or auto) lets the terminal decide.
	Theme string `json:"theme,omitempty"`
	// Icons is unicode or ascii; "" (or auto) lets the terminal decide.
	Icons string `json:"icons,omitempty"`
	// SettingsTipSeen is true once the Settings tab was opened, so Home stops
	// pointing at it.
	SettingsTipSeen bool `json:"settingsTipSeen,omitempty"`
	// Splits are the sizes the person gave the panels by moving the dividers
	// between them, in cells, by divider (for example "doctor.detail").
	Splits map[string]int `json:"splits,omitempty"`
	// CommitChoices are the answers to "commit each change so it can be
	// undone?" that are kept for the person only, by workspace root: "no" or
	// "ask". A "yes" is written in smith.toml, where smith reads it.
	CommitChoices map[string]string `json:"commitChoices,omitempty"`
}

// FileName is the file in the smith home.
const FileName = "tui.json"
