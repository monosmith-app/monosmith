package guide

// ForcedReleaseTitle is the title of the guided form that releases a package
// no commit releases: the same code under a new version.
const ForcedReleaseTitle = "Release without a releasing commit"

// ForcedReleaseAction is the form: which package, then which level. The
// packages are those the plan lists as unchanged. The answers run
// `smith release --bump <package>=<level> --force`, after the confirmation
// that says what a release does and that it cannot be undone.
func ForcedReleaseAction(unchanged []string) Action {
	var packages []Option
	if len(unchanged) == 0 {
		packages = append(packages, Option{Label: "(read the plan in the Release view first)", Value: "", Hint: "the Release view lists the packages no commit releases"})
	}
	for _, name := range unchanged {
		packages = append(packages, Option{Label: name, Value: name, Hint: "no commit since its last release releases it"})
	}
	return Action{
		Title: ForcedReleaseTitle, Hint: "smith release --bump <package>=<level> --force: the same code under a new version", NeedsWorkspace: true,
		Steps: []Step{
			{Key: "package", Prompt: "Which package should release without a releasing commit?", Options: packages, Filter: true},
			{Key: "level", Prompt: "Which level?", Options: []Option{
				{Label: "patch", Value: "patch", Hint: "1.2.3 to 1.2.4 (default)"},
				{Label: "minor", Value: "minor", Hint: "1.2.3 to 1.3.0"},
				{Label: "major", Value: "major", Hint: "1.2.3 to 2.0.0"},
			}},
		},
		Argv: func(a map[string][]string) []string {
			pick := func(key string) string {
				if v := a[key]; len(v) > 0 {
					return v[0]
				}
				return ""
			}
			return []string{"release", "--bump", pick("package") + "=" + pick("level"), "--force"}
		},
	}
}
