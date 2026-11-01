package overridelearn

// Kind says which way a suggestion points.
type Kind int

const (
	// Lift means several projects set the same value: make it the workspace default.
	Lift Kind = iota
	// Replace means nearly every project overrides the workspace default: the default is probably wrong.
	Replace
)

// Suggestion is one thing the overrides in smith.toml say about the defaults.
// It is a report only: smith never applies it by itself.
type Suggestion struct {
	Kind Kind
	// Task is the task name ("test") and Field the setting ("cache", "env.CI").
	Task, Field string
	// Value is the value as smith.toml writes it.
	Value string
	// Projects are the projects that set it, sorted.
	Projects []string
	// Total is the number of projects in the workspace.
	Total int
}
