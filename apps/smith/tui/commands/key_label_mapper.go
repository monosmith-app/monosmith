package commands

import "strings"

var keyNames = map[string]string{
	" ": "Space", "enter": "Enter", "esc": "Esc", "tab": "Tab", "shift+tab": "Shift+Tab",
	"up": "Up", "down": "Down", "left": "Left", "right": "Right", "pgup": "PgUp", "pgdown": "PgDn",
	"home": "Home", "end": "End", "delete": "Delete",
}

// KeyLabel is a Bubble Tea key name as it is shown: "ctrl+p" as "Ctrl+P",
// "f1" as "F1", a letter as itself.
func KeyLabel(key string) string {
	if n, ok := keyNames[key]; ok {
		return n
	}
	if len(key) < 2 {
		return key
	}
	parts := strings.Split(key, "+")
	for i, p := range parts {
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	if last := parts[len(parts)-1]; len(last) > 1 && !strings.HasPrefix(last, "F") {
		return strings.Join(parts, "+")
	}
	parts[len(parts)-1] = strings.ToUpper(parts[len(parts)-1])
	return strings.Join(parts, "+")
}

// keysLabel spells the first two keys of a command, "Ctrl+P / :".
func keysLabel(keys []string) string {
	var labels []string
	for _, k := range keys[:min(len(keys), 2)] {
		labels = append(labels, KeyLabel(k))
	}
	return strings.Join(labels, " / ")
}
