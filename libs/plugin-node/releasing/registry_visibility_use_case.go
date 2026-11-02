package releasing

import "time"

// Visibility checks, for the registry the release goes to, that a package
// version can be read back. npm needs seconds to minutes between a publish and
// the moment every mirror answers for it.
type Visibility struct {
	// Visible reports whether name@version can be read from the registry now.
	Visible func(name, version string) bool
	// Wait is how long to keep trying in all; Every is the pause between tries.
	Wait, Every time.Duration
	// Sleep pauses; time.Sleep when nil. Tests replace it with a fake clock.
	Sleep func(time.Duration)
}

// Await tries each package until it is visible or Wait has passed.
//
// Parameters: names are the packages; version is the one every name was
// published at.
// Returns: the names still not visible after Wait (empty when all are).
func (v Visibility) Await(names []string, version string) []string {
	sleep := v.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	pending := append([]string(nil), names...)
	for waited := time.Duration(0); ; waited += v.Every {
		var still []string
		for _, n := range pending {
			if !v.Visible(n, version) {
				still = append(still, n)
			}
		}
		pending = still
		if len(pending) == 0 || waited >= v.Wait || v.Every <= 0 {
			return pending
		}
		sleep(v.Every)
	}
}
