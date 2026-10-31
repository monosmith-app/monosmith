package convertplan

import "github.com/monosmith-app/monosmith/apps/smith/workspace"

// applyTagFormat replaces the tag_prefix of each project that follows the
// workspace's common tag format with [release] tag_format, and keeps a
// tag_prefix only on the projects that do not follow it. It changes nothing
// when no format pays off (see chooseTagFormat).
func applyTagFormat(plan *Plan) {
	format, exceptions := chooseTagFormat(plan.Config.Workspace.Scope, plan.tagChoices)
	if format == "" {
		return
	}
	plan.Config.Release.TagFormat = format
	for _, c := range plan.tagChoices {
		o := plan.Override(c.project)
		if prefix, exception := exceptions[c.project]; exception {
			o.TagPrefix = prefix
		} else {
			o.TagPrefix = ""
		}
		plan.SetOverride(c.project, o)
	}
	plan.Add(Mapped, "", "release tags", "the tags follow "+workspace.ExpandTagFormat(format, plan.Config.Workspace.Scope, "<project>")+"<version>: [release] tag_format = "+`"`+format+`"`+" names them once, and a project whose tags differ keeps its own tag_prefix")
}
