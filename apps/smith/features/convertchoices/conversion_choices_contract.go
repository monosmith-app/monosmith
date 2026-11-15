package convertchoices

// Level is how far a conversion goes.
type Level string

// The levels. Conservative keeps what works, maps what can be mapped and
// lists the rest as things to check. Full makes smith the only tool: it
// replaces the old tooling the choices retire and deletes what is left over,
// asking first.
const (
	LevelConservative Level = "conservative"
	LevelFull         Level = "full"
)

// The values a choice takes.
const (
	// Smith is smith's own: its ESLint config, its stack build, its pipeline.
	Smith = "smith"
	// Keep leaves what the workspace has.
	Keep   = "keep"
	Jest   = "jest"
	Vitest = "vitest"
)

// Choices are the decisions of a full conversion. The JSON names are the
// flags' names, so a wizard and the command line speak the same words.
type Choices struct {
	Level Level `json:"level"`
	// Lint is Smith (smith's ESLint config with the vertical-slice rules) or Keep.
	Lint string `json:"lint"`
	// Test is Jest, Vitest or Keep.
	Test string `json:"test"`
	// Build is Smith (the build smith defines for the stack) or Keep.
	Build string `json:"build"`
	// CI is Smith (the pipeline smith generates) or Keep.
	CI string `json:"ci"`
}

// Given records which choices the user already made, on the command line.
type Given struct {
	Lint, Test, Build, CI bool
}

// Any reports whether a choice was given.
func (g Given) Any() bool { return g.Lint || g.Test || g.Build || g.CI }

// Option is one answer to a question.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// Question is one choice put to the user: the flag that answers it, the
// words that ask it, and its options, the suggested one first.
type Question struct {
	Key      string   `json:"key"`
	Flag     string   `json:"flag"`
	Question string   `json:"question"`
	Options  []Option `json:"options"`
}
