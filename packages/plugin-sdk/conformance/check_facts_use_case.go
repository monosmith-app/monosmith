package conformance

import (
	"sort"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// factsProbe is the project a plugin is asked to contribute to, with shared
// facts in its request, to show that it accepts the field and publishes only
// valid facts.
var factsProbe = protocol.Project{
	Name: "conformance-probe", Root: "conformance-probe", Kind: protocol.ProjectLib,
	Facts: map[string]any{protocol.FactSourceRoots: []string{"src"}},
}

// checkFacts asks a plugin that contributes for one probe project that
// carries facts, and checks every fact in the answer: its key is namespaced
// and a well-known key holds its documented type. A plugin that refuses the
// probe project is not judged here: it may serve only projects it discovered.
func checkFacts(c *host.Client, r *Report) {
	if !c.Supports(protocol.MethodContribute) {
		return
	}
	var res protocol.ContributeResult
	if err := c.Call(protocol.MethodContribute, protocol.ContributeParams{Projects: []protocol.Project{factsProbe}}, &res, time.Minute); err != nil {
		return
	}
	names := make([]string, 0, len(res.Projects))
	for name := range res.Projects {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		err := protocol.ValidateFacts(res.Projects[name].Facts)
		r.add("contribute publishes valid facts for "+name, err == nil, "%v", err)
	}
}
