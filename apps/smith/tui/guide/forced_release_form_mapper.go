package guide

import (
	"slices"
	"strings"

	"github.com/monosmith-app/monosmith/libs/words/words"
)

// ForcedReleaseTitle is the title of the guided form that releases a package
// no commit releases: the same code under a new version.
var ForcedReleaseTitle = words.Say("guide.forced_release.title")

// NeverReleasedNote is the one line that says why a package that was never
// released is not offered, and the commit that gives it its first version.
// names are the packages; it returns "" for none.
func NeverReleasedNote(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return words.Say("guide.forced_release.not_offered", "names", strings.Join(names, ", "), "first", names[0])
}

// ForcedReleaseAction is the form: which package, then which level. The
// packages are those the plan lists as unchanged, except the ones never
// released: smith refuses a bump for them, so a note names them and the commit
// that sets their first version. The answers run
// `smith release --bump <package>=<level> --force`, after the confirmation
// that says what a release does and that it cannot be undone.
func ForcedReleaseAction(unchanged, neverReleased []string) Action {
	unchanged = slices.DeleteFunc(slices.Clone(unchanged), func(n string) bool { return slices.Contains(neverReleased, n) })
	var packages []Option
	if len(unchanged) == 0 {
		packages = append(packages, Option{Label: words.Say("guide.forced_release_form.read_the_plan_in_the_release"), Value: "", Hint: words.Say("guide.forced_release_form.the_release_view_lists_the_packages")})
	}
	for _, name := range unchanged {
		packages = append(packages, Option{Label: name, Value: name, Hint: words.Say("guide.forced_release_form.no_commit_since_its_last_release")})
	}
	return Action{
		Title: ForcedReleaseTitle, Hint: words.Say("guide.forced_release_form.smith_release_bump_package_level_force"), NeedsWorkspace: true,
		Steps: []Step{
			{Key: "package", Prompt: words.Say("guide.forced_release_form.which_package_should_release_without_a"), Options: packages, Filter: true, Note: NeverReleasedNote(neverReleased)},
			{Key: "level", Prompt: words.Say("guide.forced_release_form.which_level"), Options: []Option{
				{Label: "patch", Value: "patch", Hint: words.Say("guide.forced_release_form.1_2_3_to_1_2")},
				{Label: "minor", Value: "minor", Hint: words.Say("guide.forced_release_form.1_2_3_to_1_3")},
				{Label: "major", Value: "major", Hint: words.Say("guide.forced_release_form.1_2_3_to_2_0")},
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
