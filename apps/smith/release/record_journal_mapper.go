package release

import "time"

// JournalEntries turns what one package's release did into journal entries:
// every step it got through, and the failure that stopped it.
func JournalEntries(res ReportResult, at time.Time) []JournalEntry {
	var out []JournalEntry
	add := func(step, target, outcome string) {
		out = append(out, JournalEntry{Time: at, Version: res.Next, Step: step, Target: target, Outcome: outcome})
	}
	if res.Published {
		add("publish", res.Registry, OutcomePublished)
	}
	for _, p := range res.Packages {
		add("package", p, OutcomePublished)
	}
	for _, p := range res.ExistingPackages {
		add("package", p, OutcomeExisting)
	}
	for _, image := range res.Images {
		add("image", ImageLine(image), OutcomePublished)
	}
	switch res.Status {
	case "released", "already-released":
		add("tag", res.Tag, OutcomeDone)
	}
	if res.Pushed {
		add("push", res.Tag, OutcomeDone)
	}
	if res.Mirror != "" {
		add("mirror", res.Mirror, OutcomeDone)
	}
	if res.GitHubRelease != "" {
		add("github-release", res.GitHubRelease, OutcomeDone)
	}
	if res.PubDispatch != "" {
		add("pub-dispatch", res.PubDispatch, OutcomeDone)
	}
	if res.Status == "failed" {
		out = append(out, JournalEntry{Time: at, Version: res.Next, Step: "release", Target: res.Tag, Outcome: OutcomeFailed, Error: res.Error})
	}
	return out
}
