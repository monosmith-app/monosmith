package check

import (
	"fmt"

	"github.com/monosmith-app/monosmith/libs/plugin-architectures/rules"
)

// forbiddenImport says why a file in place `from` may not import what `to`
// names, or "" when the architecture allows it.
func forbiddenImport(rs rules.Ruleset, from, to Place, sets folderSets) string {
	if rs.Modules && to.InModules && to.Module != from.Module && !to.isPublic(sets) {
		return fmt.Sprintf("reaches into the internals of the %s module; import what sits directly in its folder, or in its api folder, instead", to.Module)
	}
	if !from.HasLayer || !to.HasLayer {
		return ""
	}
	if rs.Contexts && from.Context != "" && to.Context != "" && from.Context != to.Context {
		return fmt.Sprintf("reaches into the %s layer of the %s context; import what sits at the root of that context instead", to.Layer.Name, to.Context)
	}
	if to.Layer.Rank > from.Layer.Rank {
		return fmt.Sprintf("%s code must not depend on %s code (%s)", from.Layer.Name, to.Layer.Name, rs.Summary)
	}
	if rs.IsolatedLayer != "" && from.Layer.Name == rs.IsolatedLayer && to.Layer.Name == rs.IsolatedLayer &&
		from.Sub != "" && to.Sub != "" && from.Sub != to.Sub {
		return fmt.Sprintf("%s %s must not depend on %s %s; they meet only through the core's ports", rs.IsolatedLayer, from.Sub, rs.IsolatedLayer, to.Sub)
	}
	return ""
}
