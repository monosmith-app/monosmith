package conformance

import (
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// declaredName is the project the checks declare, at a folder that does not exist.
const declaredName = "conformance-declared"

// checkDeclared checks what every language plugin does the same with a project
// smith.toml declares (DiscoverParams.declared): one that describes declared
// projects (the declared capability) refuses a folder that does not exist,
// with an error, and invents no project for it; one that does not describe them
// ignores the declaration and answers as it would without it. Either way a
// declaration never makes discover crash.
func checkDeclared(c *host.Client, r *Report) {
	if !c.Supports(protocol.MethodDiscover) {
		return
	}
	var res protocol.DiscoverResult
	err := c.Call(protocol.MethodDiscover, protocol.DiscoverParams{
		Manifests: []string{},
		Declared:  []protocol.DeclaredProject{{Name: declaredName, Root: "conformance/no-such-folder"}},
	}, &res, time.Minute)
	invented := false
	for _, p := range res.Projects {
		if p.Name == declaredName {
			invented = true
		}
	}
	if c.Info.Capabilities.Declared {
		r.add("a declared folder that does not exist is refused, not invented", err != nil && !invented, "it must answer with an error and describe no project for a folder that is not there")
		return
	}
	r.add("a plugin that describes no declared project ignores a declaration", err == nil && !invented, "it must answer as it would without the declaration: %v", err)
}

// checkCapabilityPairs checks the capabilities that only make sense together.
func checkCapabilityPairs(c *host.Client, r *Report) {
	caps := c.Info.Capabilities
	if caps.Channels {
		r.add("a plugin that publishes channels publishes", caps.Publish, "the channels capability changes publish: it needs the publish capability")
	}
	if caps.Declared {
		r.add("a plugin that describes declared projects discovers", caps.Discover, "the declared capability changes discover: it needs the discover capability")
	}
}
