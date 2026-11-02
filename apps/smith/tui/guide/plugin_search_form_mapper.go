package guide

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PluginHit is one plugin as `smith plugin search --json` lists it.
type PluginHit struct {
	Name        string   `json:"name"`
	Newest      string   `json:"newest"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

// ParsePluginHits reads the output of `smith plugin search --json`.
//
// text is that output. It returns an error when the text is not the JSON list
// smith prints.
func ParsePluginHits(text string) ([]PluginHit, error) {
	var hits []PluginHit
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &hits); err != nil {
		return nil, fmt.Errorf("smith plugin search did not print a list of plugins: %w", err)
	}
	return hits, nil
}

// The answers of the trust question of a plugin install.
const (
	pluginTrustAsk    = "ask"
	pluginTrustPinned = "pinned"
)

// PluginTrustStep is the question that follows the choice of a plugin: install
// only (the default), or install, trust and pin it. A plugin runs with your
// privileges and smith does not sandbox it, so the second answer says what it
// gives away. smith plugin install checks the archive's SHA-256 either way.
func PluginTrustStep() Step {
	return Step{Key: "trust", Prompt: "Do you trust this plugin? A plugin runs with your privileges and smith does not sandbox it", Options: []Option{
		{Label: "Install it only (default)", Value: pluginTrustAsk, Hint: "nothing runs now; smith asks the first time the plugin is used"},
		{Label: "Install, trust and pin this version", Value: pluginTrustPinned, Hint: "it may run at once; smith writes the SHA-256 of its program to smith.toml, and CI refuses a changed file (--trust --pin)"},
	}}
}

// PluginSearchTitle is the title of the form that asks what plugin to look for.
const PluginSearchTitle = "Find a plugin in the marketplace"

// PluginInstallTitle is the title of the form that picks a found plugin.
const PluginInstallTitle = "Install a plugin from the marketplace"

// PluginInstallAction is a form that asks which of the found plugins to
// install; its answer is the command `smith plugin install <name>`, which the
// shell confirms before it runs.
//
// hits are the plugins the search found.
func PluginInstallAction(hits []PluginHit) Action {
	step := Step{Key: "plugin", Prompt: "Which plugin? (type to narrow the list)", Filter: true}
	for _, h := range hits {
		hint := h.Description
		if h.Newest != "" {
			hint = "v" + h.Newest + " · " + hint
		}
		if len(h.Roles) > 0 {
			hint += " [" + strings.Join(h.Roles, ", ") + "]"
		}
		step.Options = append(step.Options, Option{Label: h.Name, Value: h.Name, Hint: hint})
	}
	return Action{
		Title: PluginInstallTitle, Presents: "plugin install", Hint: "smith plugin install", NeedsWorkspace: true,
		Steps: []Step{step, PluginTrustStep()},
		Argv: func(a map[string][]string) []string {
			argv := []string{"plugin", "install", first(a, "plugin")}
			if first(a, "trust") == pluginTrustPinned {
				argv = append(argv, "--trust", "--pin")
			}
			return argv
		},
	}
}
