package rules

// packageByLayer is the layout of many existing codebases: one folder per
// technical layer, each layer calling only the ones below it.
var packageByLayer = Ruleset{
	Name: "package-by-layer",
	Layers: []Layer{
		{Name: "model", Rank: 0, Folders: []string{"models", "model", "entities", "entity"},
			Meaning: "plain data objects"},
		{Name: "repository", Rank: 1, Folders: []string{"repositories", "repository", "dao", "persistence"},
			Meaning: "reading and writing the stored data"},
		{Name: "service", Rank: 2, Folders: []string{"services", "service"},
			Meaning: "the business operations; they use repositories and models"},
		{Name: "controller", Rank: 3, Folders: []string{"controllers", "controller", "web", "api", "rest"},
			Meaning: "the entry points: web handlers and API endpoints; they use services and models"},
	},
	Summary: "each layer imports only the layers below it: controller, service, repository, model",
}
