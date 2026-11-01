package configedit

import (
	"fmt"
	"strings"
)

// SettingWhy answers "why is this setting what it is": its value and where the
// value comes from, what it does, why the default is what it is, and how to
// change and undo it. It is the data behind `smith why <setting>`, and the
// shape an editor can show beside a setting.
type SettingWhy struct {
	Path string `json:"path"`
	// Value is the value as smith.toml writes it; empty when the setting is
	// not set.
	Value string `json:"value,omitempty"`
	// Source is "smith.toml" or "default".
	Source string `json:"source"`
	// Default is what the setting takes when smith.toml does not set it.
	Default string `json:"default,omitempty"`
	// Explanation is what the setting does.
	Explanation string `json:"explanation"`
	// Reason is why the default is what it is; empty when no reason is recorded.
	Reason string `json:"reason,omitempty"`
	// Change and Undo are commands.
	Change string `json:"change"`
	Undo   string `json:"undo"`
}

// ExplainSetting reads one setting of the workspace at root.
//
// Parameters: root is the workspace root; path is "table.key" (ci.devcontainer).
// Returns the answer; an error when path is not a setting smith config edits,
// or smith.toml cannot be read.
func ExplainSetting(root, path string) (SettingWhy, error) {
	f, err := Find(path)
	if err != nil {
		return SettingWhy{}, err
	}
	value, set, err := Get(root, path)
	if err != nil {
		return SettingWhy{}, err
	}
	w := SettingWhy{Path: f.Path, Default: f.Default(), Explanation: f.Explanation(), Reason: f.Reason, Source: "default"}
	if set {
		w.Source, w.Value = "smith.toml", value
	}
	w.Change = "smith config set " + f.Path + " <value>"
	w.Undo = "smith config unset " + f.Path
	return w, nil
}

// Text writes the answer as a few plain lines.
func (w SettingWhy) Text() string {
	var b strings.Builder
	if w.Source == "smith.toml" {
		fmt.Fprintf(&b, "%s = %s (set in smith.toml)\n", w.Path, w.Value)
	} else if w.Default != "" {
		fmt.Fprintf(&b, "%s is not set: it takes its default, %s.\n", w.Path, w.Default)
	} else {
		fmt.Fprintf(&b, "%s is not set: it takes its default.\n", w.Path)
	}
	fmt.Fprintf(&b, "What it does: %s\n", w.Explanation)
	if w.Reason != "" {
		fmt.Fprintf(&b, "Why the default: %s\n", w.Reason)
	} else {
		b.WriteString("Why the default: no reason is recorded for this one.\n")
	}
	fmt.Fprintf(&b, "Change it: %s\n", w.Change)
	if w.Source == "smith.toml" {
		fmt.Fprintf(&b, "Go back to the default: %s\n", w.Undo)
	}
	return b.String()
}
