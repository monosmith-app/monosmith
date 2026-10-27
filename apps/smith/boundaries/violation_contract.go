package boundaries

// Violation is one dependency that breaks a rule.
type Violation struct {
	// From is the project that depends; To is the project it depends on.
	From string `json:"from"`
	To   string `json:"to"`
	// Rule says which rule the dependency breaks, with its reason if the
	// rule gave one.
	Rule string `json:"rule"`
	// Via is "import" or "schema": whether From imports To or consumes a
	// schema To owns.
	Via string `json:"via"`
	// Warning is true for the built-in rule while smith.toml has not turned
	// it on: reported, but it does not fail a check.
	Warning bool `json:"warning,omitempty"`
	// Hint says how to clear it.
	Hint string `json:"hint"`
}

// Edge is "from -> to".
func (v Violation) Edge() string { return v.From + " -> " + v.To }
