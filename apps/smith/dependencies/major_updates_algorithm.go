package dependencies

// Majors are the updates that move a package to a new major version, which
// can break the project that uses it.
//
// Parameters: updates are the updates found.
// Returns: the major ones, in order.
func Majors(updates []Update) []Update {
	var out []Update
	for _, u := range updates {
		if u.Level == "major" {
			out = append(out, u)
		}
	}
	return out
}

// WithoutMajors drops the major updates, which need a choice of their own.
//
// Parameters: updates are the updates found.
// Returns: the patch, minor and preset updates, in order.
func WithoutMajors(updates []Update) []Update {
	var out []Update
	for _, u := range updates {
		if u.Level != "major" {
			out = append(out, u)
		}
	}
	return out
}
