package turboworkspace

// ExtendsMarker, inside a list of a package configuration's task, means "the
// parent's entries go here"; without it the list replaces the parent's.
const ExtendsMarker = "$TURBO_EXTENDS$"

// MergeTask lays a child task definition over its parent, as Turborepo does
// for a package configuration (turbo.json beside a package.json, extending
// "//"): a field the child sets replaces the parent's, unless the child's
// list holds $TURBO_EXTENDS$, which stands for the parent's entries.
func MergeTask(parent, child TaskConfig) TaskConfig {
	out := parent
	if child.Extends != nil && !*child.Extends {
		out = TaskConfig{}
	}
	out.DependsOn = mergeList(out.DependsOn, child.DependsOn)
	out.Outputs = mergeList(out.Outputs, child.Outputs)
	out.Inputs = mergeList(out.Inputs, child.Inputs)
	out.Env = mergeList(out.Env, child.Env)
	out.PassThroughEnv = mergeList(out.PassThroughEnv, child.PassThroughEnv)
	out.With = mergeList(out.With, child.With)
	if child.Cache != nil {
		out.Cache = child.Cache
	}
	if child.Persistent != nil {
		out.Persistent = child.Persistent
	}
	if child.Interactive != nil {
		out.Interactive = child.Interactive
	}
	if child.OutputLogs != "" {
		out.OutputLogs = child.OutputLogs
	}
	if child.OutputMode != "" {
		out.OutputMode = child.OutputMode
	}
	out.Keys = append(append([]string(nil), parent.Keys...), child.Keys...)
	return out
}

func mergeList(parent, child []string) []string {
	if child == nil {
		return parent
	}
	out := []string{}
	for _, v := range child {
		if v != ExtendsMarker {
			out = append(out, v)
			continue
		}
		out = append(out, parent...)
	}
	return out
}
