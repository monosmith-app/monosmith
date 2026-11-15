package compose

import "slices"

// backendLanguages are the languages that make a service, in the order people
// look for them; an API outcome keeps those the workspace can generate.
var backendLanguages = []string{"node", "python", "go", "dotnet", "jvm", "rust", "php", "ruby"}

// outcomeSpec is the fixed part of an outcome: the words of the product. What
// it holds (languages, frameworks, hosts) comes from the catalogue.
type outcomeSpec struct {
	id, title, kind, layer, consequence string
	// needsHost keeps only the items of the layer that ask a trigger or runtime.
	needsHost bool
	words     []string
	// typical is used when the catalogue has nothing to list.
	typical []string
}

var outcomeSpecs = []outcomeSpec{
	{id: "web", title: "A website or web app", kind: "app", layer: "framework",
		consequence: "People open it in a browser. smith adds a dev server, tests and a build.",
		words:       []string{"website", "web app", "web", "frontend", "front end", "page", "browser", "site"}},
	{id: "api", title: "An API or backend service", kind: "app",
		consequence: "A program that answers requests on a server. You deploy it; it has no screen.",
		words:       []string{"api", "backend", "back end", "server", "service", "rest", "microservice", "http"}},
	{id: "function", title: "A function app (Azure, AWS, Google)", kind: "app", layer: "host", needsHost: true,
		consequence: "Small code that the cloud runs when something happens, such as a request or a timer.",
		words:       []string{"function", "functions", "serverless", "cloud", "lambda", "faas", "trigger"}},
	{id: "desktop", title: "A desktop app (Tauri, Electron)", kind: "app", layer: "shell",
		consequence: "A program people install on a computer. smith adds the tasks that package it.",
		words:       []string{"desktop", "windows", "mac", "linux", "native", "installer"}},
	{id: "library", title: "A library for my other projects", kind: "lib",
		consequence: "Shared code that the projects of this workspace import. It stays here and is not published.",
		words:       []string{"library", "lib", "shared", "internal", "common", "core", "reuse"}},
	{id: "package", title: "A package I publish", kind: "package",
		consequence: "Code that other people install from a registry. smith adds the publish steps.",
		typical:     []string{"npm", "PyPI", "NuGet", "crates.io"},
		words:       []string{"package", "publish", "npm", "pypi", "nuget", "crates", "registry", "sdk"}},
}

// Outcomes lists what a person can build, from the catalogue. The languages of
// a website, a function app or a desktop app are those that have an item on its
// layer, so a new plugin changes the list with no change here. An outcome that
// needs a layer the catalogue has no item on is left out.
//
// Parameters: items are the items of the catalogue (languages and choices).
// Returns: the outcomes in the order of the list, each with its languages,
// words and items.
func Outcomes(items []Item) []Outcome {
	var languages []string
	titles := map[string]string{}
	for _, it := range items {
		if it.Kind == KindLanguage {
			languages = append(languages, it.Name)
			titles[it.Name] = it.Title
		}
	}
	var out []Outcome
	for _, spec := range outcomeSpecs {
		o := Outcome{ID: spec.id, Title: spec.title, Kind: spec.kind, Layer: spec.layer, Consequence: spec.consequence, Words: slices.Clone(spec.words), Typical: spec.typical}
		switch {
		case spec.layer != "":
			for _, it := range items {
				if it.Kind != KindChoice || it.Layer != spec.layer || it.Problem != "" || (spec.needsHost && len(it.Needs) == 0) {
					continue
				}
				o.Items = append(o.Items, it)
				if !slices.Contains(o.Languages, it.Language) {
					o.Languages = append(o.Languages, it.Language)
				}
				o.Words = append(o.Words, it.Name, it.Title)
				o.Words = append(o.Words, it.Aliases...)
			}
			o.Typical = typicalTitles(o.Items)
		case spec.id == "api":
			for _, l := range backendLanguages {
				if slices.Contains(languages, l) {
					o.Languages = append(o.Languages, l)
					o.Typical = append(o.Typical, titles[l])
				}
			}
		default:
			o.Languages = slices.Clone(languages)
		}
		if len(o.Languages) > 0 {
			out = append(out, o)
		}
	}
	return out
}

// typicalTitles are the distinct titles of the items, in order.
func typicalTitles(items []Item) []string {
	var out []string
	for _, it := range items {
		if !slices.Contains(out, it.Title) {
			out = append(out, it.Title)
		}
	}
	return out
}
