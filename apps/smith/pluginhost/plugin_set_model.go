package pluginhost

import (
	"context"
	"fmt"
	"github.com/monosmith-app/monosmith/apps/smith/nearmatch"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// PluginSet is the running plugins of one smith invocation.
type PluginSet struct {
	Clients []*host.Client
}

// Get returns the plugin with a name, or nil.
func (s *PluginSet) Get(name string) *host.Client {
	for _, c := range s.Clients {
		if c.Info.Plugin.Name == name {
			return c
		}
	}
	return nil
}

// ForLanguage returns the language plugin answering to a --lang value.
func (s *PluginSet) ForLanguage(lang string) (*host.Client, error) {
	lang = strings.ToLower(strings.TrimSpace(lang))
	for _, c := range s.Clients {
		if !HasRole(c, protocol.RoleLanguage) {
			continue
		}
		if c.Info.Plugin.Name == lang || slices.Contains(c.Info.Languages, lang) {
			return c, nil
		}
	}
	return nil, fmt.Errorf("no plugin handles language %q; available: %s.%s", lang, strings.Join(s.Languages(), ", "), nearmatch.Suggestion(lang, s.Languages()))
}

// Languages lists every language plugin that can generate.
func (s *PluginSet) Languages() []string {
	var out []string
	for _, c := range s.Clients {
		if c.Info.Capabilities.Generate && HasRole(c, protocol.RoleLanguage) {
			out = append(out, c.Info.Plugin.Name)
		}
	}
	sort.Strings(out)
	return out
}

// HasRole reports whether a plugin can play a role in a stack.
func HasRole(c *host.Client, role protocol.Role) bool {
	return slices.Contains(c.Info.EffectiveRoles(), role)
}

// Call calls a method on a plugin if it supports it. ok is false when the
// plugin does not advertise the method.
func (s *PluginSet) Call(plugin, method string, params, result any, timeout time.Duration) (ok bool, err error) {
	c := s.Get(plugin)
	if c == nil {
		return false, fmt.Errorf("plugin %q is not loaded", plugin)
	}
	if !c.Supports(method) {
		return false, nil
	}
	return true, c.Call(method, params, result, timeout)
}

// CallContext is Call that gives up when ctx is done (Ctrl+C, a cancelled
// task): a plugin that advertised cancel is asked to stop and has a grace
// period to answer, and any plugin that does not is killed (see
// host.Client.CallContext).
func (s *PluginSet) CallContext(ctx context.Context, plugin, method string, params, result any, timeout time.Duration) (ok bool, err error) {
	c := s.Get(plugin)
	if c == nil {
		return false, fmt.Errorf("plugin %q is not loaded", plugin)
	}
	if !c.Supports(method) {
		return false, nil
	}
	return true, c.CallContext(ctx, method, params, result, timeout)
}

// Close shuts every plugin down.
func (s *PluginSet) Close() {
	for _, c := range s.Clients {
		_ = c.Close()
	}
}

// Offers lists the choices every loaded plugin publishes, each with the
// plugin that owns it, in the order the plugins load.
func (s *PluginSet) Offers() []choiceselect.Offer {
	var out []choiceselect.Offer
	for _, c := range s.Clients {
		for _, choice := range c.Info.Choices {
			out = append(out, choiceselect.Offer{Choice: choice, Plugin: c.Info.Plugin.Name})
		}
	}
	return out
}
