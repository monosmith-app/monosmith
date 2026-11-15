package pipelinetemplate

// Template is one pipeline smith pipeline can add.
type Template struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Providers   []string `json:"providers"`
	// NeedsProject is set for templates that build or deploy one app.
	NeedsProject bool `json:"needsProject"`
}

// Catalog lists the templates.
var Catalog = []Template{
	{Name: "e2e", Description: "run every project's e2e task, one at a time, on a push to the release branch, on pull requests (GitHub) and on demand", Providers: []string{"github", "azure"}},
	{Name: "package-zip", Description: "pack the projects (smith ci pack) and keep dist/drop as a build artifact, the input of a classic release pipeline", Providers: []string{"github", "azure"}},
	{Name: "deploy-pages", Description: "build an app and deploy its build output to GitHub Pages", Providers: []string{"github"}, NeedsProject: true},
	{Name: "deploy-azure-function", Description: "build a function app and deploy it to an Azure Function App", Providers: []string{"github", "azure"}, NeedsProject: true},
}

// Find returns the template named.
func Find(name string) (Template, bool) {
	for _, t := range Catalog {
		if t.Name == name {
			return t, true
		}
	}
	return Template{}, false
}
