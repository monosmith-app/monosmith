package splitrepo

import (
	"fmt"
	"regexp"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// TokenEnv is the secret that can push to the mirror. GITHUB_TOKEN cannot:
// it is scoped to the repository the workflow runs in.
const TokenEnv = "SPLIT_REPO_TOKEN"

// Branch is the mirror's branch, which a registry reads alongside the tags.
const Branch = "main"

var repoPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$`)

// Valid says why a split_repo value is not a GitHub owner/repo, or nothing.
func Valid(repo string) error {
	if !repoPattern.MatchString(repo) {
		return fmt.Errorf("split_repo %q is not a GitHub repository (owner/repo)", repo)
	}
	return nil
}

// Credential is the release credential a mirrored project needs.
func Credential(repo string) protocol.Credential {
	return protocol.Credential{Env: TokenEnv, Description: "token that can push to " + repo + ", the mirror a split package is released to", Required: true}
}

// Tag is the mirror's tag for a version: vX.Y.Z, the spelling Composer and
// most registries that read tags expect. The monorepo keeps its own tag.
func Tag(version string) string { return "v" + version }
