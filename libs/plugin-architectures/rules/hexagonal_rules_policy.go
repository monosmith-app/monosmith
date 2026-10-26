package rules

// hexagonal is ports and adapters: the core knows nothing of the outside,
// adapters plug into it and never into each other.
var hexagonal = Ruleset{
	Name: "hexagonal",
	Layers: []Layer{
		{Name: "core", Rank: 0, Folders: []string{"domain", "application", "core", "ports"},
			Meaning: "the business rules and the ports (interfaces) they need; no framework, database or transport"},
		{Name: "adapters", Rank: 1, Folders: []string{"adapters", "infrastructure"},
			Meaning: "what plugs into the ports: web handlers, databases, queues, clients (inbound and outbound)"},
	},
	IsolatedLayer: "adapters",
	Summary:       "the core never imports an adapter; one adapter never imports another",
}
