package images

import (
	"strings"
)

// CommitTagLength is how many characters of the commit hash the sha- tag has.
const CommitTagLength = 7

// maxTagLength is the longest tag a registry takes.
const maxTagLength = 128

// Tags returns the tags a release pushes, in push order: sha-<short commit>
// and the branch when the target asks for them, then latest, then the
// version last, so the version tag shows the registry is complete. A tag the
// target asks for but the release cannot name (no commit, no branch) is left
// out and described in missing.
func Tags(t Target, version, commit, branch string) (tags, missing []string) {
	if t.TagCommit {
		if short := shortCommit(commit); short != "" {
			tags = append(tags, "sha-"+short)
		} else {
			missing = append(missing, "tag_commit: the commit is unknown")
		}
	}
	if t.TagBranch {
		if name := BranchTag(branch); name != "" {
			tags = append(tags, name)
		} else {
			missing = append(missing, "tag_branch: the branch is unknown")
		}
	}
	return append(tags, "latest", version), missing
}

func shortCommit(commit string) string {
	commit = strings.ToLower(strings.TrimSpace(commit))
	if len(commit) < CommitTagLength {
		return ""
	}
	return commit[:CommitTagLength]
}

// BranchTag turns a branch name into a valid image tag: letters, digits,
// underscores, dots and dashes, not starting with a dot or a dash, at most 128
// characters ("feature/x" becomes "feature-x"). It is "" for a name with
// nothing usable in it.
func BranchTag(branch string) string {
	var b strings.Builder
	for _, r := range strings.TrimPrefix(strings.TrimSpace(branch), "refs/heads/") {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_', r == '.', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	tag := strings.TrimLeft(b.String(), ".-")
	if len(tag) > maxTagLength {
		tag = tag[:maxTagLength]
	}
	return tag
}
