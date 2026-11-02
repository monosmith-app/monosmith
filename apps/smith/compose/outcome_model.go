package compose

// Outcome is something a person wants to build, in plain words ("a website",
// "a function app"), and what it means for the stack. The Add wizard asks for
// an outcome first and the language only where the outcome leaves a choice.
type Outcome struct {
	// ID is the short name of the outcome: web, api, function, desktop, library
	// or package.
	ID string `json:"id"`
	// Title is the answer in plain words ("A website or web app").
	Title string `json:"title"`
	// Kind is the kind of project smith add makes for it: app, lib or package.
	Kind string `json:"kind"`
	// Layer is the layer of the catalogue whose items name the outcome (framework,
	// shell or host); empty when the outcome has no such layer.
	Layer string `json:"layer,omitempty"`
	// Typical lists the things people pick for it, for the line of the list.
	Typical []string `json:"typical"`
	// Consequence says in one sentence what the person gets.
	Consequence string `json:"consequence"`
	// Languages are the languages that fit the outcome, in catalogue order. A
	// list of one means there is no language to ask.
	Languages []string `json:"languages"`
	// Words find the outcome besides its title: its own words and the names and
	// aliases of the catalogue items that name it ("react", "electron", "lambda").
	Words []string `json:"words"`
	// Items are the catalogue items of the outcome's layer a person can pick
	// (React, Vue and Svelte for a website); empty for an outcome with none.
	Items []Item `json:"items,omitempty"`
}
