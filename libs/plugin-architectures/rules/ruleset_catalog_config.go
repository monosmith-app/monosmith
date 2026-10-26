package rules

// all is every architecture the checker knows, in the order they are
// offered. Adding one is a ruleset file and a row here, and a row in the
// architecture catalog of smith.
var all = []Ruleset{hexagonal, clean, ddd, modularMonolith}

// For returns an architecture's rules.
func For(name string) (Ruleset, bool) {
	for _, r := range all {
		if r.Name == name {
			return r, true
		}
	}
	return Ruleset{}, false
}

// Names lists the architectures that have rules, in the order they are
// offered.
func Names() []string {
	names := make([]string, len(all))
	for i, r := range all {
		names[i] = r.Name
	}
	return names
}
