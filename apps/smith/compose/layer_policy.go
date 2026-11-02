package compose

import "slices"

// The layers smith knows, in stack order. A choice whose plugin names another
// layer (or whose axis has no layer here) gets a layer of that name, after
// these, so a new plugin needs no change in the host.
var knownLayers = []Layer{
	{ID: "language", Title: "Language", About: "what the project is written in"},
	{ID: "framework", Title: "Framework", About: "how the app or library is written: a web framework, or a whole starter"},
	{ID: "shell", Title: "Desktop shell", About: "what wraps the app as a program you install on a computer"},
	{ID: "host", Title: "Cloud host", About: "where the code runs in the cloud, with the trigger and the runtime it asks for"},
	{ID: "library", Title: "Libraries", About: "extra libraries the app draws on"},
	{ID: "style", Title: "Style", About: "how the app looks: a CSS library"},
	{ID: "test", Title: "Test", About: "the tool that runs the tests"},
	{ID: "lint", Title: "Lint", About: "the tool that checks and formats the code"},
	{ID: "architecture", Title: "Architecture", About: "the rule for how the code is laid out in folders"},
	{ID: "ci", Title: "CI and deploy", About: "the pipelines that build, test and deploy the project"},
}

// AboutLayer says in plain words what sits on a layer ("what wraps the app as a
// program you install on a computer"), for a screen that meets the word
// "framework", "shell" or "host" before a person has read anything.
//
// Parameters: id is a layer id.
// Returns: the sentence, or "" for a layer smith does not know.
func AboutLayer(id string) string {
	for _, l := range knownLayers {
		if l.ID == id {
			return l.About
		}
	}
	return ""
}

// LayerOf is the layer of a choice: the layer its plugin names, else its axis.
//
// Parameters: layer is protocol.Choice.Layer; axis is protocol.Choice.Axis.
// Returns: the layer id.
func LayerOf(layer, axis string) string {
	if layer != "" {
		return layer
	}
	return axis
}

// layersFor returns the layers that hold an item, in stack order: the known
// ones first, then the plugin's own, by name.
func layersFor(items []Item) []Layer {
	used := map[string]bool{}
	for _, it := range items {
		used[it.Layer] = true
	}
	var out []Layer
	for _, l := range knownLayers {
		if used[l.ID] {
			out = append(out, l)
			delete(used, l.ID)
		}
	}
	var extra []string
	for id := range used {
		extra = append(extra, id)
	}
	slices.Sort(extra)
	for _, id := range extra {
		out = append(out, Layer{ID: id, Title: id, About: "choices on the " + id + " axis"})
	}
	return out
}

// layerRank is the place of a layer in stack order; unknown layers come last.
func layerRank(id string) int {
	for i, l := range knownLayers {
		if l.ID == id {
			return i
		}
	}
	return len(knownLayers)
}
