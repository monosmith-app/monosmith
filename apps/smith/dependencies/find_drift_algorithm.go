package dependencies

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Drift is one dependency the projects of one plugin declare differently.
type Drift struct {
	Plugin string `json:"plugin"`
	Name   string `json:"name"`
	// Target is the spec every declaration converges on.
	Target string `json:"target"`
	// Reason is "installed" (the operator most declarations use, at the
	// version installed) or "highest" (the highest version declared).
	Reason string `json:"reason"`
	// Fixable are the declarations sync rewrites; Blocked are those it must
	// leave alone and reports.
	Fixable []protocol.Declaration `json:"fixable"`
	// Crossing are declarations whose major version differs from the target:
	// sync rewrites them only when told to cross majors.
	Crossing []protocol.Declaration `json:"crossing,omitempty"`
	Blocked  []protocol.Declaration `json:"blocked,omitempty"`
	// Sites are every declaration of the dependency, the ones that already
	// agree with the target too.
	Sites []protocol.Declaration `json:"sites,omitempty"`
	// Installed is the version the package manager has installed, "" when unknown.
	Installed string `json:"installed,omitempty"`
	// pickTarget is smith's own target, kept when the person picks another.
	pickTarget string
}

// classify sets the target of a drift and sorts its sites into the ones sync
// rewrites, the ones that would cross a major version, and the ones it must
// leave alone.
func classify(d Drift, target string) Drift {
	d.Target = target
	d.Fixable, d.Crossing, d.Blocked = nil, nil, nil
	for _, s := range d.Sites {
		switch {
		case s.Spec == target:
		case s.Rewritable && crossesMajor(s.Spec, target):
			d.Crossing = append(d.Crossing, s)
		case s.Rewritable:
			d.Fixable = append(d.Fixable, s)
		default:
			d.Blocked = append(d.Blocked, s)
		}
	}
	return d
}

var (
	operator = regexp.MustCompile(`^[\^~><=!]*`)
	version  = regexp.MustCompile(`^\d+(\.\d+){0,2}([-.+]?[A-Za-z][\w.]*)?$`)
)

// splitSpec splits a single-constraint spec into its operator and version;
// ok is false for anything else (a tag, URL, compound range).
func splitSpec(spec string) (op, v string, ok bool) {
	spec = strings.TrimSpace(spec)
	op = operator.FindString(spec)
	v = strings.TrimSpace(spec[len(op):])
	return op, v, version.MatchString(v)
}

// newer reports whether version a is higher than b: numeric parts compared
// in turn (a missing part is 0), and a release above its pre-releases.
func newer(a, b string) bool {
	parse := func(s string) ([3]int, string) {
		var n [3]int
		core, pre := s, ""
		if i := strings.IndexFunc(s, func(r rune) bool { return r != '.' && (r < '0' || r > '9') }); i >= 0 {
			core, pre = s[:i], strings.TrimLeft(s[i:], "-.+")
		}
		for i, part := range strings.SplitN(core, ".", 3) {
			n[i], _ = strconv.Atoi(part)
		}
		return n, pre
	}
	an, ap := parse(a)
	bn, bp := parse(b)
	for i := range an {
		if an[i] != bn[i] {
			return an[i] > bn[i]
		}
	}
	switch {
	case ap == bp:
		return false
	case ap == "":
		return true
	case bp == "":
		return false
	}
	return ap > bp
}

// FindDrift finds the dependencies one plugin's manifests declare at more
// than one spec. Peer declarations and the workspace's own packages (local)
// do not count.
func FindDrift(plugin string, declared protocol.DeclaredResult, local map[string]bool) []Drift {
	byName := map[string][]protocol.Declaration{}
	for _, d := range declared.Declarations {
		if d.Peer || local[d.Name] {
			continue
		}
		byName[d.Name] = append(byName[d.Name], d)
	}
	var out []Drift
	for name, sites := range byName {
		specs := map[string]bool{}
		for _, s := range sites {
			specs[s.Spec] = true
		}
		if len(sites) < 2 || len(specs) < 2 {
			continue
		}
		target, reason := targetOf(sites, declared.Installed[name])
		if target == "" {
			continue
		}
		out = append(out, classify(Drift{Plugin: plugin, Name: name, Reason: reason, Sites: sites, Installed: declared.Installed[name]}, target))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// targetOf is the operator most declarations use at the installed version,
// or, when nothing says what is installed, the highest spec declared.
func targetOf(sites []protocol.Declaration, installed string) (string, string) {
	if installed != "" {
		counts := map[string]int{}
		for _, s := range sites {
			if op, _, ok := splitSpec(s.Spec); ok {
				counts[op]++
			}
		}
		best, bestCount := "", -1
		for op, n := range counts {
			if n > bestCount || (n == bestCount && (op == "^" || (best != "^" && op < best))) {
				best, bestCount = op, n
			}
		}
		if bestCount > 0 {
			return best + installed, "installed"
		}
		return "", ""
	}
	best, bestVersion := "", ""
	for _, s := range sites {
		if _, v, ok := splitSpec(s.Spec); ok && (bestVersion == "" || newer(v, bestVersion)) {
			best, bestVersion = strings.TrimSpace(s.Spec), v
		}
	}
	return best, "highest"
}

// majorOf is the first number of the version a spec names, or -1.
func majorOf(spec string) int {
	_, v, ok := splitSpec(spec)
	if !ok {
		return -1
	}
	n, err := strconv.Atoi(strings.SplitN(v, ".", 2)[0])
	if err != nil {
		return -1
	}
	return n
}

// crossesMajor reports whether moving a declaration from spec to target
// changes the major version. A spec with no readable major never crosses.
func crossesMajor(spec, target string) bool {
	a, b := majorOf(spec), majorOf(target)
	return a >= 0 && b >= 0 && a != b
}
