package doctor

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/overridelearn"
	"github.com/monosmith-app/monosmith/apps/smith/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// overrideSuggestionChecks reports what the per-project task overrides say
// about the defaults (ADR 0056): a value several projects set, and a default
// nearly every project overrides. Both are warnings with a remedy; nothing is
// changed. A workspace with no such pattern has none.
func overrideSuggestionChecks(s *session.Session) []protocol.Finding {
	var out []protocol.Finding
	for _, sg := range overridelearn.Analyse(s.Config, len(s.Graph.Projects), overridelearn.DefaultThreshold) {
		out = append(out, overrideSuggestionFinding(sg))
	}
	return out
}

func overrideSuggestionFinding(sg overridelearn.Suggestion) protocol.Finding {
	setting := fmt.Sprintf("[tasks.%s] %s", sg.Task, sg.Field)
	names := strings.Join(sg.Projects, ", ")
	if len(sg.Projects) > 5 {
		names = strings.Join(sg.Projects[:5], ", ") + fmt.Sprintf(" and %d more", len(sg.Projects)-5)
	}
	if sg.Kind == overridelearn.Replace {
		return protocol.Finding{
			Check:   "a workspace default is not overridden by nearly every project",
			Warning: true,
			Detail:  fmt.Sprintf("%d of %d projects override %s = %s (%s)", len(sg.Projects), sg.Total, setting, sg.Value, names),
			Remedy:  fmt.Sprintf("the default is probably wrong: change %s in smith.toml to what they set and remove the overrides they repeat", setting),
		}
	}
	return protocol.Finding{
		Check:   "repeated project overrides are a workspace default",
		Warning: true,
		Detail:  fmt.Sprintf("%d projects set %s = %s (%s)", len(sg.Projects), setting, sg.Value, names),
		Remedy:  fmt.Sprintf("make it the default for all projects: set %s = %s under [tasks.%s] in smith.toml, then remove it from the projects that set it", sg.Field, sg.Value, sg.Task),
	}
}
