package toolinstall

import (
	"context"
	"sync"
	"time"
)

// DetectTimeout is how long detection of all tools may take. A program that
// does not answer in time counts as missing, so detection never makes a
// command slow.
const DetectTimeout = 3 * time.Second

// Probe is what a detector found for one tool.
type Probe struct {
	// Program is the program that answered; empty when none did.
	Program string
	// Version is the version it printed; empty when it printed none.
	Version string
	// Folder, when Program is empty, is a folder off PATH that holds the
	// program: where winget unpacked a portable package.
	Folder string
}

// Detector asks the machine whether a tool is there. The real one runs the
// tool's program (ExecDetector); tests give a fake.
type Detector interface {
	// Detect runs the version command of the tool. It returns a Probe with an
	// empty Program when no program of the tool runs before ctx ends.
	Detect(ctx context.Context, tool Need) Probe
}

// DetectAll detects every need in parallel within DetectTimeout and reads the
// version range each workspace pin gives.
//
// Parameters: needs are the tools to detect; d is the detector.
// Returns: one status per need, in the same order.
func DetectAll(parent context.Context, d Detector, needs []Need) []Status {
	ctx, cancel := context.WithTimeout(parent, DetectTimeout)
	defer cancel()
	out := make([]Status, len(needs))
	var wg sync.WaitGroup
	for i, n := range needs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			out[i] = detectOne(ctx, d, n)
		}()
	}
	wg.Wait()
	return out
}

// detectOne detects one tool and judges its version against the pin.
func detectOne(ctx context.Context, d Detector, n Need) Status {
	want, from := FindRequirement(n.Pins, n.Dirs)
	if n.Exact && n.Suggested != "" {
		want, from = Requirement{Prefix: n.Suggested, Min: n.Suggested}, "the version CI installs"
	}
	st := Status{Need: n, Wanted: want, WantedFrom: from}
	probe := d.Detect(ctx, n)
	switch {
	case probe.Program == "":
		st.State, st.OffPath, st.Version = Missing, probe.Folder, probe.Version
	case !want.Satisfied(probe.Version):
		st.State, st.Program, st.Version = Outdated, probe.Program, probe.Version
	default:
		st.State, st.Program, st.Version = Found, probe.Program, probe.Version
	}
	return st
}
