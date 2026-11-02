package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// What the first commit of smith new holds.
const (
	// CommitNone makes no commit.
	CommitNone = "none"
	// CommitSmith commits only the files smith wrote.
	CommitSmith = "smith"
	// CommitAll commits every file in the folder (secret names need a second yes).
	CommitAll = "all"
)

// maxListed is how many files the question lists before it says "and N more".
const maxListed = 15

// CommitRequest is what ChooseFirstCommit needs to decide.
type CommitRequest struct {
	// Dir is the folder smith new will fill.
	Dir string
	// Mode is --first-commit: none, smith, all, or "".
	Mode string
	// NoCommit is --no-commit.
	NoCommit bool
	// IncludeSecrets is --include-secrets: files with secret names may be committed.
	IncludeSecrets bool
	// Yes is --yes: take the safe default instead of asking.
	Yes bool
	// Interactive says a person can answer.
	Interactive bool
	// In and Out carry the questions.
	In  *bufio.Reader
	Out io.Writer
}

// CommitDecision is the answer: what the first commit holds.
type CommitDecision struct {
	// Mode is CommitNone, CommitSmith, CommitAll, or "" when the folder is
	// fresh (empty or already a repository) and the usual rule applies.
	Mode string
	// IncludeSecrets lets CommitAll take files with secret names.
	IncludeSecrets bool
}

// ChooseFirstCommit decides what the first commit of smith new holds. A fresh
// folder (empty, missing, or already a git repository) keeps the usual rule
// and only --no-commit changes it. A folder with files and no repository is a
// real choice, because the first commit could carry those files into history:
// smith lists them and offers three answers. In a terminal it asks. With --yes
// it takes the safe one (CommitSmith) and says so. Without a terminal and
// without --yes it refuses and names the flags.
//
// Parameters: r is the request.
// Returns: the decision; an error when the flags contradict, a secret would
// be committed without --include-secrets, or nobody can answer.
func ChooseFirstCommit(r CommitRequest) (CommitDecision, error) {
	if r.Mode != "" && r.Mode != CommitNone && r.Mode != CommitSmith && r.Mode != CommitAll {
		return CommitDecision{}, fmt.Errorf("--first-commit %s: use none, smith or all", r.Mode)
	}
	if r.NoCommit && r.Mode != "" && r.Mode != CommitNone {
		return CommitDecision{}, problem.New("--no-commit and --first-commit "+r.Mode+" say opposite things", "use one of them")
	}
	if r.NoCommit {
		return CommitDecision{Mode: CommitNone}, nil
	}
	if _, err := os.Stat(filepath.Join(r.Dir, ".git")); err == nil {
		// Someone else's repository: smith commits only what it wrote, unless told not to.
		if r.Mode == CommitAll {
			return CommitDecision{}, problem.New("--first-commit all needs a folder that is not a git repository yet", "use --first-commit smith or none")
		}
		return CommitDecision{Mode: r.Mode}, nil
	}
	content, err := InspectFolder(r.Dir)
	if err != nil {
		return CommitDecision{}, err
	}
	if len(content.Files) == 0 {
		if r.Mode == CommitAll {
			return CommitDecision{Mode: ""}, nil
		}
		return CommitDecision{Mode: r.Mode}, nil
	}
	mode := r.Mode
	if mode == "" {
		switch {
		case r.Interactive:
			mode = askFirstCommit(r, content)
		case r.Yes:
			mode = CommitSmith
			fmt.Fprintf(r.Out, "The folder holds %d file(s) and is not a git repository. smith commits only its own files (--first-commit all commits everything, --no-commit commits nothing).\n", len(content.Files))
		default:
			return CommitDecision{}, problem.New(
				fmt.Sprintf("%s holds %d file(s) and is not a git repository, and nobody can say what the first commit should hold", r.Dir, len(content.Files)),
				"--first-commit smith commits only the files smith writes", "--first-commit all commits everything in the folder", "--no-commit commits nothing")
		}
	}
	d := CommitDecision{Mode: mode, IncludeSecrets: r.IncludeSecrets}
	if mode == CommitAll && len(content.Secrets) > 0 && !r.IncludeSecrets {
		if !r.Interactive || !confirmSecrets(r, content.Secrets) {
			if r.Interactive {
				fmt.Fprintln(r.Out, "These files stay out of the commit.")
			} else {
				return CommitDecision{}, problem.New(
					"these files look like secrets and would be committed: "+strings.Join(content.Secrets, ", "),
					"--include-secrets commits them anyway", "--first-commit smith commits only the files smith writes")
			}
		} else {
			d.IncludeSecrets = true
		}
	}
	return d, nil
}

// askFirstCommit lists what the folder holds and asks what the first commit takes.
func askFirstCommit(r CommitRequest, c FolderContent) string {
	fmt.Fprintf(r.Out, "\n%s holds %d file(s) and is not a git repository:\n", r.Dir, len(c.Files))
	for i, f := range c.Files {
		if i == maxListed {
			fmt.Fprintf(r.Out, "  and %d more\n", len(c.Files)-i)
			break
		}
		mark := ""
		if IsSecretName(f) {
			mark = "   <- looks like a secret"
		}
		fmt.Fprintf(r.Out, "  %s%s\n", f, mark)
	}
	q := NewQuestion{
		Prompt:  "What should the first commit hold? (smith runs git init either way)",
		Default: CommitSmith,
		Flag:    "first-commit",
		Options: []NewOption{
			{Label: "Only smith's files", Value: CommitSmith, Hint: "your files stay uncommitted, for you to add"},
			{Label: "Nothing", Value: CommitNone, Hint: "smith writes its files and commits nothing"},
			{Label: "Everything shown above", Value: CommitAll, Hint: "all of it goes into git history; files that look like secrets need a second yes"},
		},
	}
	answer, _ := askNewQuestion(q, r.In, r.Out)
	fmt.Fprintf(r.Out, "First commit: %s (change it with --first-commit smith|all|none, or --no-commit)\n", map[string]string{
		CommitSmith: "only smith's files", CommitNone: "nothing", CommitAll: "everything shown",
	}[answer])
	return answer
}

// confirmSecrets asks, once more, whether files with secret names may be committed.
func confirmSecrets(r CommitRequest, secrets []string) bool {
	fmt.Fprintf(r.Out, "\nThese files look like secrets: %s\nOnce committed they stay in git history even if you delete them later.\n", strings.Join(secrets, ", "))
	q := NewQuestion{
		Prompt: "Commit them anyway?", Default: "no", Flag: "include-secrets",
		Options: []NewOption{
			{Label: "No", Value: "no", Hint: "commit everything else"},
			{Label: "Yes", Value: "yes", Hint: "commit them too"},
		},
	}
	answer, _ := askNewQuestion(q, r.In, r.Out)
	return answer == "yes"
}
