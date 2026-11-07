package release

import (
	"fmt"
	"os"
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
		if res.Manual {
			fmt.Fprintf(&b, "Manual release: no commit releases %s.\n", res.Project)
		}
		for _, image := range res.Images {
			fmt.Fprintf(&b, "Image pushed for %s: %s\n", res.Project, ImageLine(image))
		}
		for _, f := range res.Packed {
			fmt.Fprintf(&b, "Attached to GitHub Release %s: %s (sha256 %s)\n", res.Tag, f.Name, f.SHA256[:12])
		}
		if res.PubDispatch != "" {
			fmt.Fprintf(&b, "Queued the pub.dev publish for %s (%s). It is not finished yet.\n", res.PubDispatch, PubWorkflow)
		}
		if res.Mirror != "" {
			verb := "Mirrored"
			if r.DryRun {
				verb = "Would mirror"
			}
			fmt.Fprintf(&b, "%s %s to %s\n", verb, res.Project, res.Mirror)
		}
	}
	if lines := r.SummaryLines(asciiWanted()); len(lines) > 0 {
		b.WriteString("\n")
		for _, line := range lines {
			b.WriteString(line.Text + "\n")
		}
	}
	if pending := r.PendingTags(); len(pending) > 0 {
		b.WriteString("\nWarning: npm did not serve every platform package in time, so a dist-tag was not moved.\nThe packages are published under the dist-tag \"pending\". When npm serves them, run:\n")
		for _, command := range pending {
			fmt.Fprintf(&b, "  %s\n", command)
		}
		b.WriteString("The release exits with code 2 so that CI shows a warning.\n")
	}
	if r.Error != "" && r.DryRun {
		b.WriteString("\nThe real release would fail for:\n")
		for _, res := range r.Results {
			if res.Status == "failed" {
				fmt.Fprintf(&b, "  %s: %s\n", res.Project, res.Error)
			}
		}
	} else if r.Error != "" {
		fmt.Fprintf(&b, "\nStopped: %s\nRe-run the release: finished packages are kept and the rest resume.\n", r.Error)
	}
	if r.DryRun && len(r.Results) > 0 && r.Error == "" {
		b.WriteString("\n" + PublishHintLine + "\n")
	}
	return b.String()
}

// PublishHintLine ends the report of a preview: it says how the release goes on.
// The shell leaves it out, because its own screen has the button that releases.
const PublishHintLine = "→ Push to the release branch and CI releases these; or run smith release here to publish them yourself."

// asciiWanted reports whether the terminal needs the ASCII banner: the same
// signals the TUI icon set reads (SMITH_ASCII=1, TERM=dumb or linux).
func asciiWanted() bool {
	term := os.Getenv("TERM")
	return os.Getenv("SMITH_ASCII") == "1" || term == "dumb" || term == "linux"
}
