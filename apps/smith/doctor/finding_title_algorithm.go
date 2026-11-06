package doctor

// Title is how a list names a finding: what is wrong when it failed and says
// so, else the rule it checks. A failed finding that gives no Problem keeps
// listing its rule until its plugin does.
func (f Finding) Title() string {
	if !f.OK && f.Problem != "" {
		return f.Problem
	}
	return f.Check
}

// Located is Title with the project in front ("web: compiles into ..."), for a
// flat list that is not filed under project headings. A finding about no
// project is its Title alone.
func (f Finding) Located() string {
	if f.Project == "" {
		return f.Title()
	}
	return f.Project + ": " + f.Title()
}
