package guide

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/scaffold"
	"github.com/monosmith-app/monosmith/apps/smith/tui/words"
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
			Prompt: words.Say("guide.new_workspace_decision_steps.this_folder_already_holds_files_and"),
			Options: []Option{
				{Label: words.Say("guide.new_workspace_decision_steps.only_the_files_smith_writes"), Value: scaffold.CommitSmith, Hint: words.Say("guide.new_workspace_decision_steps.your_own_files_stay_out_of")},
				{Label: words.Say("guide.new_workspace_decision_steps.nothing"), Value: scaffold.CommitNone, Hint: words.Say("guide.new_workspace_decision_steps.smith_makes_no_commit_you_commit")},
				{Label: words.Say("guide.new_workspace_decision_steps.everything_in_the_folder"), Value: scaffold.CommitAll, Hint: words.Say("guide.new_workspace_decision_steps.every_file_goes_into_the_first")},
			},
			Skip: func(a map[string][]string) bool {
				c, ok := holds(a)
				return !ok || len(c.Secrets) > 0
			},
		},
		{
			Key:    answerFirstCommitSecrets,
			Prompt: words.Say("guide.new_workspace_decision_steps.this_folder_holds_files_whose_names"),
			Options: []Option{
				{Label: words.Say("guide.new_workspace_decision_steps.only_the_files_smith_writes"), Value: scaffold.CommitSmith, Hint: words.Say("guide.new_workspace_decision_steps.your_files_secrets_included_stay_out")},
				{Label: words.Say("guide.new_workspace_decision_steps.nothing"), Value: scaffold.CommitNone, Hint: words.Say("guide.new_workspace_decision_steps.smith_makes_no_commit")},
				{Label: words.Say("guide.new_workspace_decision_steps.everything_secrets_included"), Value: scaffold.CommitAll + "+secrets", Hint: words.Say("guide.new_workspace_decision_steps.a_secret_in_history_stays_there")},
			},
			Skip: func(a map[string][]string) bool {
				c, ok := holds(a)
				return !ok || len(c.Secrets) == 0
			},
		},
		{
			Key:         answerModulePrefix,
			Prompt:      words.Say("guide.new_workspace_decision_steps.this_workspace_has_no_git_remote"),
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
