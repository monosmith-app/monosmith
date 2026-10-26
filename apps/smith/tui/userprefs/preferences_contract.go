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
}

// FileName is the file in the smith home.
const FileName = "tui.json"
