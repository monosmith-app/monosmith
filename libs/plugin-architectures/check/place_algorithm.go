package check

import "github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"

// Place is where a file, or what an import names, sits in an architecture.
type Place struct {
	// Layer is the layer it is in, when a folder on its path names one.
	Layer    rules.Layer
	HasLayer bool
	// Context is the bounded context around the layer ("" when there is none).
	Context string
	// Module is the module it is in ("" outside modules/); ModuleRest is the
	// path inside the module, and InModules says there is a module.
	Module     string
	ModuleRest []string
	InModules  bool
	// Sub is the folder inside the layer that groups its parts (the adapter's
	// name), "" when the code sits directly in the layer.
	Sub string
}

// folderSets are the names of the folders a project has below its layers and
// modules, which tell a folder from a file in an import such as
// "adapters/postgres".
type folderSets struct {
	// layerChildren are folders directly inside a layer folder.
	layerChildren map[string]bool
	// moduleChildren are folders directly inside a module folder.
	moduleChildren map[string]bool
}

// genericRoots are folders that hold a project's code and say nothing about
// a context.
var genericRoots = map[string]bool{"src": true, "main": true, "java": true, "kotlin": true, "internal": true, "lib": true, "app": true, "pkg": true, "cmd": true, "source": true, "contexts": true, "boundedcontexts": true}

// directions are the folders that split a layer into inbound and outbound
// halves; the adapter is the folder inside them.
var directions = map[string]bool{"inbound": true, "outbound": true, "driving": true, "driven": true, "primary": true, "secondary": true, "in": true, "out": true}

// classify places a path. segs are its folders and, for an import, the last
// name; isFile says segs are a file's folders only, with its name left off.
func classify(rs rules.Ruleset, segs []string, isFile bool, sets folderSets) Place {
	var p Place
	layerAt := -1
	for i, s := range segs {
		if l, ok := rs.LayerOf(s); ok {
			p.Layer, p.HasLayer, layerAt = l, true, i
			break
		}
	}
	if layerAt >= 0 {
		if rs.Contexts && layerAt > 0 && !genericRoots[rules.Normalise(segs[layerAt-1])] {
			p.Context = segs[layerAt-1]
		}
		p.Sub = subOf(segs[layerAt+1:], isFile, sets.layerChildren)
	}
	if rs.Modules {
		for i, s := range segs {
			if rules.Normalise(s) == "modules" && i+1 < len(segs) {
				p.InModules, p.Module, p.ModuleRest = true, segs[i+1], segs[i+2:]
				break
			}
		}
	}
	return p
}

// subOf is the first folder below a layer, past the inbound/outbound
// folders, when it is a folder and not the file or name an import ends in.
func subOf(rest []string, isFile bool, children map[string]bool) string {
	for len(rest) > 0 && directions[rules.Normalise(rest[0])] {
		rest = rest[1:]
	}
	if len(rest) == 0 {
		return ""
	}
	if isFile || len(rest) > 1 || children[rules.Normalise(rest[0])] {
		return rest[0]
	}
	return ""
}

// isPublic reports whether what an import names inside a module is the
// module's public face: what sits directly in its folder, or in api/.
func (p Place) isPublic(sets folderSets) bool {
	if len(p.ModuleRest) == 0 {
		return true
	}
	first := rules.Normalise(p.ModuleRest[0])
	if first == "api" || first == "public" {
		return true
	}
	return len(p.ModuleRest) == 1 && !sets.moduleChildren[first]
}
