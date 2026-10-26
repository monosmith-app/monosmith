package rules

// mvc is model, view, controller: the model knows neither of the others, a
// view reads the model, a controller connects the two.
var mvc = Ruleset{
	Name: "mvc",
	Layers: []Layer{
		{Name: "model", Rank: 0, Folders: []string{"models", "model"},
			Meaning: "the data and the rules about it; it knows nothing of how it is shown or requested"},
		{Name: "view", Rank: 1, Folders: []string{"views", "view", "templates"},
			Meaning: "what is shown: pages, templates and view components; it reads the model"},
		{Name: "controller", Rank: 2, Folders: []string{"controllers", "controller"},
			Meaning: "takes a request, works with the model and picks the view"},
	},
	Summary: "the model imports neither views nor controllers, and a view never imports a controller",
}
