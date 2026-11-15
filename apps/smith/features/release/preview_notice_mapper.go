package release

import "github.com/monosmith-app/monosmith/libs/words/words"

// PreviewNotice is the one line that says a plan is a preview of a merge.
//
// Parameters: current is the branch checked out; plan is the plan, whose
// PreviewOf names the release branch.
// Returns: the line, or "" when the plan is not a preview.
func PreviewNotice(current string, plan Plan) string {
	if plan.PreviewOf == "" {
		return ""
	}
	return words.Say("releaseview.preview", "branch", current, "release", plan.PreviewOf)
}
