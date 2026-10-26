package check

// Report is the outcome of a check. Problems fail it; warnings ask for a
// look.
type Report struct {
	Problems []string
	Warnings []string
}

// OK reports whether nothing breaks the rules.
func (r Report) OK() bool { return len(r.Problems) == 0 }

func (r *Report) problem(path, why string) { r.Problems = append(r.Problems, path+" - "+why) }

func (r *Report) warn(path, why string) { r.Warnings = append(r.Warnings, path+" - "+why) }
