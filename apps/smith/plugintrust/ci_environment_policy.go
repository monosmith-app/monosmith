package plugintrust

// InCI reports whether the environment says smith runs in a CI service: the
// CI variable every service sets, or the GitHub Actions and Azure Pipelines
// markers. In CI nobody is there to be asked, and a person's trust decisions
// are not on the machine.
func InCI(getenv func(string) string) bool {
	return getenv("CI") != "" || getenv("GITHUB_ACTIONS") == "true" || getenv("TF_BUILD") != ""
}
