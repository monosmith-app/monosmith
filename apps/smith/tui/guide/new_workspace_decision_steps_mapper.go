package guide

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
)

// The keys of the decisions the form asks beyond smith new's own questions.
const (
	answerFirstCommit        = "first-commit"
	answerFirstCommitSecrets = "first-commit-secrets"
	answerModulePrefix       = "module-prefix"
)

// decisionSteps are the questions smith new asks only when the situation calls
// for them: what the first commit holds in a folder that has files, and the
// Go module prefix when the workspace has no git remote.
func decisionSteps() []Step {
	holds := func(a map[string][]string) (scaffold.FolderContent, bool) {
		dir := strings.TrimSpace(first(a, "name"))
		if dir == "" {
			return scaffold.FolderContent{}, false
		}
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return scaffold.FolderContent{}, false
		}
		c, err := scaffold.InspectFolder(dir)
		return c, err == nil && len(c.Files) > 0
	}
	return []Step{
		{
			Key:    answerFirstCommit,
			Prompt: "This folder already holds files and is not a git repository. What should the first commit hold?",
			Options: []Option{
				{Label: "Only the files smith writes", Value: scaffold.CommitSmith, Hint: "your own files stay out of the commit (default)"},
				{Label: "Nothing", Value: scaffold.CommitNone, Hint: "smith makes no commit; you commit when you are ready"},
				{Label: "Everything in the folder", Value: scaffold.CommitAll, Hint: "every file goes into the first commit and its history"},
			},
			Skip: func(a map[string][]string) bool {
				c, ok := holds(a)
				return !ok || len(c.Secrets) > 0
			},
		},
		{
			Key:    answerFirstCommitSecrets,
			Prompt: "This folder holds files whose names look like secrets (.env, keys). What should the first commit hold?",
			Options: []Option{
				{Label: "Only the files smith writes", Value: scaffold.CommitSmith, Hint: "your files, secrets included, stay out of the commit (default)"},
				{Label: "Nothing", Value: scaffold.CommitNone, Hint: "smith makes no commit"},
				{Label: "Everything, secrets included", Value: scaffold.CommitAll + "+secrets", Hint: "a secret in history stays there after you delete the file"},
			},
			Skip: func(a map[string][]string) bool {
				c, ok := holds(a)
				return !ok || len(c.Secrets) == 0
			},
		},
		{
			Key:         answerModulePrefix,
			Prompt:      "This workspace has no git remote. Which Go module prefix should its projects use? Every Go import path starts with it and it is hard to change later. Enter takes the one shown.",
			Placeholder: "github.com/acme/my-studio",
			Optional:    true,
			PlaceholderFor: func(a map[string][]string) string {
				p, _ := scaffold.SuggestModulePrefix(strings.TrimSpace(first(a, "name")))
				return p
			},
			Skip: func(a map[string][]string) bool {
				return first(a, scaffold.AnswerLanguage) != "go"
			},
		},
	}
}

// decisionFlags are the flags the answers of decisionSteps build.
func decisionFlags(a map[string][]string) []string {
	var argv []string
	mode := first(a, answerFirstCommit)
	if mode == "" {
		mode = first(a, answerFirstCommitSecrets)
	}
	switch mode {
	case "":
	case scaffold.CommitAll + "+secrets":
		argv = append(argv, "--first-commit", scaffold.CommitAll, "--include-secrets")
	default:
		argv = append(argv, "--first-commit", mode)
	}
	if first(a, scaffold.AnswerLanguage) == "go" {
		prefix := strings.TrimSpace(first(a, answerModulePrefix))
		if prefix == "" {
			prefix, _ = scaffold.SuggestModulePrefix(strings.TrimSpace(first(a, "name")))
		}
		argv = append(argv, "--module-prefix", prefix)
	}
	return argv
}
