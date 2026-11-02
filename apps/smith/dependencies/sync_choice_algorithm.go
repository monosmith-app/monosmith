package dependencies

import (
	"fmt"
	"slices"
	"strings"
)

// The ways to say which version wins when projects disagree.
const (
	// PreferPick is smith's own pick: the installed version, else the highest declared.
	PreferPick = ""
	// PreferHighest takes the highest version any project declares.
	PreferHighest = "highest"
	// PreferLowest takes the lowest version any project declares.
	PreferLowest = "lowest"
	// PreferInstalled takes the version the package manager has installed.
	PreferInstalled = "installed"
)

// Candidate is one version a dependency can converge on.
type Candidate struct {
	// Spec is the declaration every project would get.
	Spec string
	// Why says in plain words where the version comes from.
	Why string
}

// Candidates lists the versions the projects can agree on: smith's pick, the
// highest and the lowest declared, the installed one, and what each project
// declares now. A version listed twice keeps its first reason.
//
// Parameters: d is a dependency whose projects disagree.
// Returns: the candidates, smith's pick first.
func Candidates(d Drift) []Candidate {
	var out []Candidate
	add := func(spec, why string) {
		if spec == "" || slices.ContainsFunc(out, func(c Candidate) bool { return c.Spec == spec }) {
			return
		}
		out = append(out, Candidate{Spec: spec, Why: why})
	}
	pick := d.Target
	if d.pickTarget != "" {
		pick = d.pickTarget
	}
	add(pick, "smith's pick: the installed version, else the highest declared")
	if s, ok := extreme(d, true); ok {
		add(s, "the highest declared")
	}
	if s, ok := extreme(d, false); ok {
		add(s, "the lowest declared")
	}
	if s, ok := installedSpec(d); ok {
		add(s, "the version installed")
	}
	for _, site := range d.Sites {
		add(strings.TrimSpace(site.Spec), "what "+where(site)+" declares now")
	}
	return out
}

// extreme is the highest (or lowest) of the specs the projects declare.
func extreme(d Drift, highest bool) (string, bool) {
	best, bestVersion := "", ""
	for _, s := range d.Sites {
		_, v, ok := splitSpec(s.Spec)
		if !ok {
			continue
		}
		if bestVersion == "" || (highest && newer(v, bestVersion)) || (!highest && newer(bestVersion, v)) {
			best, bestVersion = strings.TrimSpace(s.Spec), v
		}
	}
	return best, best != ""
}

// installedSpec is the installed version with the operator most projects use.
func installedSpec(d Drift) (string, bool) {
	if d.Installed == "" {
		return "", false
	}
	counts := map[string]int{}
	for _, s := range d.Sites {
		if op, _, ok := splitSpec(s.Spec); ok {
			counts[op]++
		}
	}
	best, n := "", -1
	for op, c := range counts {
		if c > n || (c == n && op < best) {
			best, n = op, c
		}
	}
	if n < 0 {
		return "", false
	}
	return best + d.Installed, true
}

// Prefer retargets a dependency by a rule: highest, lowest or installed.
//
// Parameters: d is the dependency; rule is a Prefer constant.
// Returns: the dependency with its new target, and false when the rule cannot
// apply (no installed version known, or no readable version declared).
func Prefer(d Drift, rule string) (Drift, bool) {
	var spec string
	var ok bool
	switch rule {
	case PreferPick:
		return d, true
	case PreferHighest:
		spec, ok = extreme(d, true)
	case PreferLowest:
		spec, ok = extreme(d, false)
	case PreferInstalled:
		spec, ok = installedSpec(d)
	}
	if !ok {
		return d, false
	}
	return Retarget(d, spec), true
}

// Retarget makes spec the version every project converges on.
//
// Parameters: d is the dependency; spec is the new target.
// Returns: the dependency with its sites sorted again for that target.
func Retarget(d Drift, spec string) Drift {
	if d.pickTarget == "" {
		d.pickTarget = d.Target
	}
	d.Reason = "chosen"
	return classify(d, spec)
}

// ValidatePrefer checks the value of --prefer.
func ValidatePrefer(rule string) error {
	switch rule {
	case PreferPick, PreferHighest, PreferLowest, PreferInstalled:
		return nil
	}
	return fmt.Errorf("--prefer %q: use highest, lowest or installed", rule)
}
