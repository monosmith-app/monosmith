package release

import (
	"fmt"
	"strings"
)

// Text renders the end of a release run for a terminal. The plan (each
// package's old and new version and why) was printed before the run, so a
// dry run adds only what to do next; a real run lists what happened to each
// package. Markdown is for a CI step summary or a PR comment, not a terminal.
func (r Report) Text() string {
	var b strings.Builder
	if !r.DryRun {
		for _, res := range r.Results {
			from := res.Current
			if from == "" {
				from = "new"
			}
			fmt.Fprintf(&b, "  %-20s %s → %s  %s\n", res.Project, from, res.Next, res.Status)
		}
	}
	for _, res := range r.Results {
		for _, image := range res.Images {
			fmt.Fprintf(&b, "Image pushed for %s: %s\n", res.Project, ImageLine(image))
		}
		if res.PubDispatch != "" {
			fmt.Fprintf(&b, "Dispatched pub.dev publish for %s.\n", res.PubDispatch)
		}
		if res.Mirror != "" {
			verb := "Mirrored"
			if r.DryRun {
				verb = "Would mirror"
			}
			fmt.Fprintf(&b, "%s %s to %s\n", verb, res.Project, res.Mirror)
		}
	}
	if r.Error != "" {
		fmt.Fprintf(&b, "\nStopped: %s\nRe-run the release: finished packages are kept and the rest resume.\n", r.Error)
	}
	if r.DryRun && len(r.Results) > 0 && r.Error == "" {
		b.WriteString("\n→ Push to the release branch and CI releases these; or run smith release here to publish them yourself.\n")
	}
	return b.String()
}
