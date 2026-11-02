package compose

import (
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
)

// Language is a language plugin as the catalogue lists it.
type Language struct {
	// Name is the plugin name and the value of --lang.
	Name string
	// Title and Description are the words shown to a person.
	Title, Description string
}

// WorkspacePlugin is a plugin that is turned on for the workspace (an
// architecture or a CI and deploy plugin), listed but not part of --with.
type WorkspacePlugin struct {
	// Name is the plugin name.
	Name string
	// Layer is "architecture" or "ci".
	Layer string
	// Description says what the plugin does, in plain words.
	Description string
}

// Source is what the catalogue is built from. Everything in it comes from the
// loaded plugins.
type Source struct {
	// Offers are the choices of all loaded plugins.
	Offers []choiceselect.Offer
	// Languages are the loaded language plugins that can generate.
	Languages []Language
	// Plugins are the loaded structure and deploy plugins.
	Plugins []WorkspacePlugin
}

// Build makes the catalogue from the source: one item for each language, each
// choice and each workspace plugin, placed on its layer, with the problem of a
// choice that cannot be used here.
//
// Parameters: src is what the loaded plugins publish.
// Returns: the catalogue, with the layers that hold an item in stack order and
// the items in stack order, then by title.
func Build(src Source) Catalogue {
	loaded := map[string]bool{}
	var items []Item
	for _, l := range src.Languages {
		loaded[l.Name] = true
		items = append(items, Item{
			ID: KindLanguage + "/" + l.Name, Kind: KindLanguage, Name: l.Name, Title: titleOr(l.Title, l.Name),
			Description: l.Description, Layer: "language", Language: l.Name,
		})
	}
	problems := choiceselect.Problems(src.Offers, func(language string) bool { return loaded[language] })
	for _, o := range src.Offers {
		items = append(items, Item{
			ID: o.Language + "/" + o.Name, Kind: KindChoice, Name: o.Name, Title: titleOr(o.Title, o.Name),
			Description: o.Description, Aliases: o.Aliases, Layer: LayerOf(o.Layer, o.Axis), Axis: o.Axis,
			Language: o.Language, Default: o.Default, Suggested: o.Suggested, Needs: o.Needs,
			Requires: o.Requires, Conflicts: o.Conflicts, Many: o.Many, Roles: rolesOf(o), Plugin: o.Plugin,
			Problem: problems[choiceselect.ProblemKey(o)],
		})
	}
	for _, p := range src.Plugins {
		items = append(items, Item{
			ID: KindPlugin + "/" + p.Name, Kind: KindPlugin, Name: p.Name, Title: p.Name, Description: p.Description,
			Layer: p.Layer, Route: "smith plugin enable " + p.Name,
		})
	}
	sort.SliceStable(items, func(i, j int) bool {
		a, b := items[i], items[j]
		if ra, rb := layerRank(a.Layer), layerRank(b.Layer); ra != rb {
			return ra < rb
		}
		if a.Layer != b.Layer {
			return a.Layer < b.Layer
		}
		if ta, tb := strings.ToLower(a.Title), strings.ToLower(b.Title); ta != tb {
			return ta < tb
		}
		return a.Language < b.Language
	})
	return Catalogue{Layers: layersFor(items), Items: items}
}

// titleOr is the title, or the name when the plugin gave none.
func titleOr(title, name string) string {
	if title != "" {
		return title
	}
	return name
}

// rolesOf is the roles of an offer as text.
func rolesOf(o choiceselect.Offer) []string {
	var out []string
	for _, r := range o.Roles {
		out = append(out, string(r))
	}
	return out
}
