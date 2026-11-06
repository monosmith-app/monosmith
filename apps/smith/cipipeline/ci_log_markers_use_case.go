package cipipeline

import (
	"fmt"
	"io"
	"os"
	"strings"
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
