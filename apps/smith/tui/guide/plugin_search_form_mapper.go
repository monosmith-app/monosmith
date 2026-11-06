package guide

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
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
		return nil, fmt.Errorf("%s: %w", words.Say("guide.plugin_search.unreadable"), err) // words:literal
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
	return Step{Key: "trust", Prompt: words.Say("guide.plugin_search_form.do_you_trust_this_plugin_a"), Options: []Option{
		{Label: words.Say("guide.plugin_search_form.install_it_only_default"), Value: pluginTrustAsk, Hint: words.Say("guide.plugin_search_form.nothing_runs_now_smith_asks_the")},
		{Label: words.Say("guide.plugin_search_form.install_trust_and_pin_this_version"), Value: pluginTrustPinned, Hint: words.Say("guide.plugin_search_form.it_may_run_at_once_smith")},
	}}
}

// PluginSearchTitle is the title of the form that asks what plugin to look for.
var PluginSearchTitle = words.Say("guide.plugin_search.title")

// PluginInstallTitle is the title of the form that picks a found plugin.
var PluginInstallTitle = words.Say("guide.plugin_search.install_title")

// PluginInstallAction is a form that asks which of the found plugins to
// install; its answer is the command `smith plugin install <name>`, which the
// shell confirms before it runs.
//
// hits are the plugins the search found.
func PluginInstallAction(hits []PluginHit) Action {
	step := Step{Key: "plugin", Prompt: words.Say("guide.plugin_search_form.which_plugin_type_to_narrow_the"), Filter: true}
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
		Title: PluginInstallTitle, Presents: words.Say("guide.plugin_search_form.plugin_install"), Hint: words.Say("guide.plugin_search_form.smith_plugin_install"), NeedsWorkspace: true,
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
