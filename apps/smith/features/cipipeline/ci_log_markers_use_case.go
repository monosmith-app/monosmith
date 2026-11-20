package cipipeline

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
)

// inGroupedLog reports whether the CI log folds groups (GitHub or Azure).
func inGroupedLog() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("TF_BUILD") != ""
}

func group(out io.Writer, name string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::group::%s\n", name)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##[group]%s\n", name)
	}
}

func endGroup(out io.Writer) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintln(out, "::endgroup::")
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintln(out, "##[endgroup]")
	}
}

func annotate(out io.Writer, phase string, err error) {
	msg := strings.ReplaceAll(err.Error(), "\n", " ")
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::error title=smith ci %s::%s\n", phase, msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=error]smith ci %s: %s\n", phase, msg)
	}
}

// notice puts a result in the run summary of the CI that has notices.
func notice(out io.Writer, msg string) {
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		fmt.Fprintf(out, "::notice title=smith ci release::%s\n", msg)
	}
}

func warn(out io.Writer, msg string) {
	switch {
	case os.Getenv("GITHUB_ACTIONS") == "true":
		fmt.Fprintf(out, "::warning::%s\n", msg)
	case os.Getenv("TF_BUILD") != "":
		fmt.Fprintf(out, "##vso[task.logissue type=warning]%s\n", msg)
	default:
		fmt.Fprintln(out, "warning: "+msg)
	}
}

// tellAzure writes the [ci.azure] logging commands of the release plan (build
// tags, DEPLOY_<PROJECT> variables, the build number) when the run is on Azure
// Pipelines (TF_BUILD set); elsewhere it writes nothing. Only smith ci pack
// calls it, so the release phase does not write them a second time.
func tellAzure(s *session.Session, plan release.Plan, out io.Writer) {
	if os.Getenv("TF_BUILD") == "" {
		return
	}
	for _, line := range azureReleaseMarkers(s.Config.CI.Azure, plan, deployableNames(s.Graph.Projects)) {
		fmt.Fprintln(out, line)
	}
}
