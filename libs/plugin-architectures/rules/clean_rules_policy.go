package rules

// clean is the dependency rule: source code depends only inward, from
// frameworks to adapters to use cases to entities.
var clean = Ruleset{
	Name: "clean",
	Layers: []Layer{
		{Name: "entities", Rank: 0, Folders: []string{"entities", "entity", "enterprise"},
			Meaning: "the business objects and rules that hold in every application"},
		{Name: "use-cases", Rank: 1, Folders: []string{"use-cases", "usecase", "interactors"},
			Meaning: "the rules of this application: one class or function per thing a person can do"},
		{Name: "interface-adapters", Rank: 2, Folders: []string{"interface-adapters", "adapters", "controllers", "presenters", "gateways"},
			Meaning: "converts between the use cases and the outside: controllers, presenters, gateways"},
		{Name: "frameworks", Rank: 3, Folders: []string{"frameworks", "drivers", "frameworks-and-drivers", "infrastructure"},
			Meaning: "the database, the web framework, devices and every other tool"},
	},
	Summary: "code depends only inward: frameworks, then interface adapters, then use cases, then entities",
}
