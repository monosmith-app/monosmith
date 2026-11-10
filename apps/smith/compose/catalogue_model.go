package compose

import (
	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Kinds of item.
const (
	// KindLanguage is a language plugin: the first layer of a stack.
	KindLanguage = "language"
	// KindChoice is a choice a plugin publishes; it goes into --with.
	KindChoice = "choice"
	// KindPlugin is a workspace plugin (architecture, CI and deploy) that is
	// turned on with smith plugin, not with --with.
	KindPlugin = "plugin"
)

// Item is one thing the catalogue lists.
type Item struct {
	// ID is unique in the catalogue: language/name, because a choice name is
	// unique per language only.
	ID string `json:"id"`
	// Kind is KindLanguage, KindChoice or KindPlugin.
	Kind string `json:"kind"`
	// Name is what goes into --with (or --lang for a language).
	Name string `json:"name"`
	// Title is the name shown to a person.
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	// Layer is where the item sits in the stack (see Layers).
	Layer string `json:"layer"`
	// Axis is the question a choice answers; empty for a language or a plugin.
	Axis string `json:"axis,omitempty"`
	// Language is the language the item is for; empty for a plugin that serves
	// every language.
	Language string `json:"language,omitempty"`
	// Default and Suggested are as in a choice (default: added when its axis is
	// needed; suggested: pre-selected when asked).
	Default   bool     `json:"default,omitempty"`
	Suggested bool     `json:"suggested,omitempty"`
	Needs     []string `json:"needs,omitempty"`
	Requires  []string `json:"requires,omitempty"`
	Conflicts []string `json:"conflicts,omitempty"`
	// Many, Roles and Plugin are what the solver needs to check a pick again
	// where the plugins are not loaded (the terminal screen): see OffersOf.
	Many   bool     `json:"many,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	Plugin string   `json:"plugin,omitempty"`
	// Problem says why the item cannot be used here; empty when it can.
	Problem string `json:"problem,omitempty"`
	// Route is set when the item is not composed with --with: the command that
	// turns it on, with the reason.
	Route string `json:"route,omitempty"`
}

// Layer is one level of the stack, from the language up to the pipelines.
type Layer struct {
	// ID is the value of Item.Layer.
	ID    string `json:"id"`
	Title string `json:"title"`
	// About says in plain words what sits on the layer.
	About string `json:"about"`
}

// Catalogue is the layers and the items of a workspace's plugins.
type Catalogue struct {
	// Layers are in stack order and only those that hold an item.
	Layers []Layer `json:"layers"`
	Items  []Item  `json:"items"`
}

// Find returns the item with the id, or the only item with the name.
//
// Parameters: key is an ID, or a name that one item has.
// Returns: the item and whether exactly one matched.
func (c Catalogue) Find(key string) (Item, bool) {
	var found []Item
	for _, it := range c.Items {
		if it.ID == key {
			return it, true
		}
		if it.Name == key {
			found = append(found, it)
		}
	}
	if len(found) == 1 {
		return found[0], true
	}
	return Item{}, false
}

// OffersOf turns the choice items of a catalogue back into offers, so a screen
// that holds only the catalogue checks the picks with the same solver as smith
// add. Languages and plugins are not offers.
//
// Parameters: items are the items of a catalogue.
// Returns: one offer for each choice item, in order.
func OffersOf(items []Item) []choiceselect.Offer {
	var out []choiceselect.Offer
	for _, it := range items {
		if it.Kind != KindChoice {
			continue
		}
		var roles []protocol.Role
		for _, r := range it.Roles {
			roles = append(roles, protocol.Role(r))
		}
		out = append(out, choiceselect.Offer{Plugin: it.Plugin, Choice: protocol.Choice{
			Axis: it.Axis, Name: it.Name, Title: it.Title, Description: it.Description, Aliases: it.Aliases,
			Layer: it.Layer, Language: it.Language, Roles: roles, Needs: it.Needs, Default: it.Default,
			Suggested: it.Suggested, Many: it.Many, Requires: it.Requires, Conflicts: it.Conflicts,
		}})
	}
	return out
}
