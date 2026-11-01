package adoption

import "github.com/monosmith-app/monosmith/apps/smith/releaselineage"

// What an item stops.
const (
	// BlocksAdoption stops smith adopt --init.
	BlocksAdoption = "adoption"
	// BlocksRelease would make the first release wrong.
	BlocksRelease = "release"
)

// Report is `smith adopt --json`.
type Report struct {
	Root string `json:"root"`
	// Name, Scope and Providers are what --init would write.
	Name      string   `json:"name"`
	Scope     string   `json:"scope"`
	Providers []string `json:"providers"`
	// Clean is true when git has no uncommitted changes; the steps need it.
	Clean    bool          `json:"clean"`
	Projects []ProjectLine `json:"projects"`
	// Declared are the projects --init declares in smith.toml for a repository
	// whose Go code is one module at the root (ADR 0052).
	Declared []DeclaredProject `json:"declared,omitempty"`
	Blockers []Item            `json:"blockers"`
	Warnings []Item            `json:"warnings"`
}

// Item is one finding and the step that clears it.
type Item struct {
	Message string `json:"message"`
	// Step is the usual way out; Steps lists every way, Step first.
	Step  string   `json:"step"`
	Steps []string `json:"steps,omitempty"`
	// Blocks is BlocksAdoption or BlocksRelease, for blockers.
	Blocks string `json:"blocks,omitempty"`
	// Baseline is the tag --tags creates for it.
	Baseline *releaselineage.Baseline `json:"baseline,omitempty"`
}

// ProjectLine is one project smith would see.
type ProjectLine struct {
	Name     string `json:"name"`
	Root     string `json:"root"`
	Language string `json:"language"`
	Kind     string `json:"kind"`
	Stack    string `json:"stack,omitempty"`
}

// Adoptable reports whether nothing stops --init.
func (r Report) Adoptable() bool {
	for _, b := range r.Blockers {
		if b.Blocks == BlocksAdoption {
			return false
		}
	}
	return true
}
