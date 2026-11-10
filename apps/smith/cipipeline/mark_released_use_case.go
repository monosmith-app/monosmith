package cipipeline

import (
	"os"

	"github.com/monosmith-app/monosmith/apps/smith/release"
)

// markReleased tells the workflow, as soon as the run knows it, that at least
// one project was released: it appends released=true to GitHub Actions' step
// output file. A release that stops half way (six plugins released, the seventh
// failed) still leaves releases that need their files, so the jobs that attach
// them run on this output, not on the release step's success.
func markReleased(report release.Report) {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return
	}
	for _, r := range report.Results {
		if r.Status != "released" {
			continue
		}
		if f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644); err == nil {
			_, _ = f.WriteString("released=true\n")
			_ = f.Close()
		}
		return
	}
}
