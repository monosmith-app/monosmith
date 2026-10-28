package release

import (
	"fmt"
	"strings"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Report is the machine-readable record of a release run, written to
// .smith/release-report.json and printed by --json.
type Report struct {
	DryRun   bool           `json:"dryRun"`
	Started  time.Time      `json:"started"`
	Finished time.Time      `json:"finished"`
	Results  []ReportResult `json:"results"`
	// Unchanged lists packages with nothing to release.
	Unchanged []string `json:"unchanged,omitempty"`
	// Error is set when the release stopped early.
	Error string `json:"error,omitempty"`
}

// ReportResult is one package's outcome.
type ReportResult struct {
	Item
	// Status is planned (dry run), released, already-released or failed.
	Status    string   `json:"status"`
	Files     []string `json:"files,omitempty"`
	Published bool     `json:"published"`
	Publish   string   `json:"publish,omitempty"`
	Registry  string   `json:"registry,omitempty"`
	Pushed    bool     `json:"pushed"`
	// Assets are the files built for an app's release.
	Assets []string `json:"assets,omitempty"`
	// Packages are the npm packages, PyPI wheels, NuGet packages and gems an
	// app was published as (name@version, wheel, .nupkg or gem file name), and
	// those the registries already had.
	Packages         []string `json:"packages,omitempty"`
	ExistingPackages []string `json:"existingPackages,omitempty"`
	// Images are the container images the release pushed, one per registry,
	// with their tags and the manifest digest the registry holds: what a
	// deploy step or a GitOps manifest refers to.
	Images []protocol.PublishedImage `json:"images,omitempty"`
	// GitHubRelease is the URL of the tag's GitHub Release.
	GitHubRelease string `json:"githubRelease,omitempty"`
	// PubDispatch names the pub.dev publish the release started (package,
	// version and tag), which the pub workflow finishes.
	PubDispatch string `json:"pubDispatch,omitempty"`
	// Mirror is the split repository and tag the release was pushed to
	// (owner/repo@vX.Y.Z), when the project has one.
	Mirror string `json:"mirror,omitempty"`
	// TookMs is how long the package took to release, in milliseconds.
	TookMs int64  `json:"tookMs,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Markdown renders the report for a CI step summary or a PR comment.
func (r Report) Markdown() string {
	var b strings.Builder
	title := "Release"
	if r.DryRun {
		title = "Release preview"
	}
	fmt.Fprintf(&b, "### %s\n\n", title)
	if len(r.Results) == 0 {
		b.WriteString("Nothing to release.\n")
	} else {
		b.WriteString("| Package | Version | Why | Status |\n|---|---|---|---|\n")
		for _, res := range r.Results {
			from := res.Current
			if from == "" {
				from = "new"
			}
			fmt.Fprintf(&b, "| %s | %s → **%s** | %s | %s |\n", res.Project, from, res.Next, res.Reason, res.Status)
		}
	}
	for _, res := range r.Results {
		if len(res.Images) > 0 {
			fmt.Fprintf(&b, "\nImages pushed for %s:\n\n", res.Project)
			for _, image := range res.Images {
				fmt.Fprintf(&b, "- `%s`\n", ImageLine(image))
			}
		}
		if res.PubDispatch != "" {
			fmt.Fprintf(&b, "\nDispatched pub.dev publish for %s.\n", res.PubDispatch)
		}
		if res.Mirror != "" {
			verb := "Mirrored"
			if r.DryRun {
				verb = "Would mirror"
			}
			fmt.Fprintf(&b, "\n%s %s to %s\n", verb, res.Project, res.Mirror)
		}
	}
	if len(r.Unchanged) > 0 {
		fmt.Fprintf(&b, "\nUnchanged: %s\n", strings.Join(r.Unchanged, ", "))
	}
	if r.Error != "" && r.DryRun {
		b.WriteString("\n**The real release would fail for:**\n\n")
		for _, res := range r.Results {
			if res.Status == "failed" {
				fmt.Fprintf(&b, "- %s: %s\n", res.Project, res.Error)
			}
		}
	} else if r.Error != "" {
		fmt.Fprintf(&b, "\n**Stopped:** %s\n\nRe-run the release: finished packages are kept and the rest resume.\n", r.Error)
	}
	return b.String()
}

// ImageLine describes a pushed image on one line: the reference by digest
// when the registry's digest is known (the immutable one a deploy refers to),
// else without a tag, then the tags pushed.
func ImageLine(image protocol.PublishedImage) string {
	ref := image.Reference
	if image.Digest != "" {
		ref += "@" + image.Digest
	}
	return ref + " [tags: " + strings.Join(image.Tags, ", ") + "]"
}
