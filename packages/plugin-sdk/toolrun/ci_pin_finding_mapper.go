package toolrun

import (
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// CIPinFinding compares the version a tool runs at on this machine with the
// version the generated CI workflow and dev container install (ci, as the
// pipeline contribution wrote it: "24", "1.90", "10.0"). The comparison looks
// only at as many leading components as ci has, so "1.90" accepts 1.90.2.
//
//   - the same version is a pass;
//   - a newer local one is a warning: it builds, but can accept what CI's
//     older tool rejects;
//   - an older local one fails: the tool may refuse what CI builds, with an
//     error that reads like a dependency problem.
//
// A ci that is not a number ("stable", "nightly") gives no finding (ok is
// false): there is nothing to compare. The caller adds the remedy, and a Fix
// where it can apply one; name is the tool as the user knows it ("Rust").
func CIPinFinding(name, have, ci string) (f protocol.Finding, ok bool) {
	order, ok := comparePin(have, ci)
	if !ok {
		return protocol.Finding{}, false
	}
	f = protocol.Finding{
		Check: name + " is " + ci + ", as in CI",
		OK:    order == 0, Warning: order > 0,
		Detail: "found " + name + " " + have,
	}
	switch {
	case order < 0:
		f.Detail += ", older than the " + ci + " the generated CI and dev container install: builds that pass there can fail here"
	case order > 0:
		f.Detail += ", newer than the " + ci + " the generated CI and dev container install: it can accept what CI rejects"
	}
	return f, true
}

// comparePin orders have against ci over ci's components: -1 older, 0 the
// same, 1 newer. ok is false when either is not a dotted number.
func comparePin(have, ci string) (order int, ok bool) {
	h, hok := numbers(have)
	c, cok := numbers(ci)
	if !hok || !cok {
		return 0, false
	}
	for i, want := range c {
		got := 0
		if i < len(h) {
			got = h[i]
		}
		if got != want {
			if got < want {
				return -1, true
			}
			return 1, true
		}
	}
	return 0, true
}

// numbers reads "v1.90.2" as 1, 90, 2; it fails on anything else.
func numbers(version string) ([]int, bool) {
	version = strings.TrimPrefix(strings.TrimSpace(version), "v")
	if version == "" {
		return nil, false
	}
	var out []int
	for _, part := range strings.Split(version, ".") {
		n, err := strconv.Atoi(part)
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, true
}
