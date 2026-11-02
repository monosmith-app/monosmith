package npmpropagation

import (
	"fmt"
	"strings"
	"time"
)

// Await probes every package until all are fetchable on Confirmations
// consecutive probes, or Max has passed.
//
// Parameters: reg answers the probes; clock tells the time and pauses; cfg
// tunes the wait; names are the packages; version is the one all were
// published at.
// Returns: the outcome; Ready is false when Max passed first.
func Await(reg Registry, clock Clock, cfg Config, names []string, version string) Outcome {
	cfg = withDefaults(cfg)
	start := clock.Now()
	lastProgress := start
	delay := cfg.FirstDelay / 2
	passed := 0
	for {
		missing := probe(reg, names, version)
		elapsed := clock.Now().Sub(start)
		if len(missing) == 0 {
			passed++
			if passed >= cfg.Confirmations || cfg.Max <= 0 {
				return Outcome{Ready: true, Waited: elapsed}
			}
			delay = cfg.Spacing
		} else {
			passed = 0
			delay = min(delay*2, cfg.MaxDelay)
		}
		// A confirming probe is always a full Spacing after the last one.
		if cfg.Max <= 0 || elapsed >= cfg.Max || (passed > 0 && elapsed+delay > cfg.Max) {
			return Outcome{Missing: missing, Waited: elapsed}
		}
		if cfg.Progress != nil && clock.Now().Sub(lastProgress) >= cfg.ProgressEvery {
			lastProgress = clock.Now()
			cfg.Progress(progressLine(missing, passed, cfg.Confirmations, elapsed))
		}
		clock.Sleep(min(delay, cfg.Max-elapsed))
	}
}

// probe returns the packages that cannot be fetched now.
func probe(reg Registry, names []string, version string) []string {
	var missing []string
	for _, n := range names {
		found, tarball := reg.Lookup(n, version)
		if !found || tarball == "" || !reg.Fetchable(tarball) {
			missing = append(missing, n)
		}
	}
	return missing
}

func progressLine(missing []string, passed, need int, elapsed time.Duration) string {
	if len(missing) > 0 {
		return fmt.Sprintf("waiting for npm to spread the release (%s): not fetchable yet: %s", elapsed.Round(time.Second), strings.Join(missing, ", "))
	}
	return fmt.Sprintf("waiting for npm to spread the release (%s): every package answered %d of %d times", elapsed.Round(time.Second), passed, need)
}

func withDefaults(c Config) Config {
	if c.Confirmations <= 0 {
		c.Confirmations = DefaultConfirmations
	}
	if c.Spacing <= 0 {
		c.Spacing = DefaultSpacing
	}
	if c.FirstDelay <= 0 {
		c.FirstDelay = DefaultFirstDelay
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = DefaultMaxDelay
	}
	if c.ProgressEvery <= 0 {
		c.ProgressEvery = DefaultProgressEvery
	}
	return c
}

// ParseWait reads a propagation_wait setting.
//
// Parameters: s is a duration such as "15m", "90s" or "0"; empty takes the default.
// Returns: the duration. Errors: s is not a duration, or is negative.
func ParseWait(s string) (time.Duration, error) {
	if strings.TrimSpace(s) == "" {
		return DefaultMax, nil
	}
	d, err := time.ParseDuration(strings.TrimSpace(s))
	if err != nil || d < 0 {
		return 0, fmt.Errorf("propagation_wait must be a duration such as \"15m\" or \"0\" (do not wait), not %q", s)
	}
	return d, nil
}
