package rules

// ddd is domain-driven design: one folder per bounded context, each holding
// its own domain, application and infrastructure, and contexts reaching each
// other only through what sits at their root.
var ddd = Ruleset{
	Name: "ddd",
	Layers: []Layer{
		{Name: "domain", Rank: 0, Folders: []string{"domain"},
			Meaning: "the model of the context: entities, value objects, aggregates, domain services and events"},
		{Name: "application", Rank: 1, Folders: []string{"application"},
			Meaning: "the use cases of the context: they orchestrate the domain and own the transactions"},
		{Name: "infrastructure", Rank: 2, Folders: []string{"infrastructure", "infra"},
			Meaning: "databases, messaging and clients that implement what the domain and application declare"},
	},
	Contexts: true,
	Summary:  "domain imports nothing above it, application imports only the domain, and one context never imports another's layers",
}
