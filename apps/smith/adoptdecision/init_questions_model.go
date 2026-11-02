package adoptdecision

import (
	"slices"
	"strings"
)

// The keys of the questions Questions returns.
const (
	KeyGitattributes = "gitattributes"
	KeyCI            = "ci"
	KeyExisting      = "existing"
)

// DefaultCI is the value of the CI answer that takes the providers smith
// proposes: the ones the repository uses, else GitHub.
const DefaultCI = "default"

// Answer is one answer to a question, with what it does in plain words.
type Answer struct{ Label, Value, Hint string }

// Question is one decision of adopt --init, for an interface that asks it
// (the terminal of smith adopt asks the same through Resolve).
type Question struct {
	Key, Prompt string
	// Answers are the options, the default first.
	Answers []Answer
}

// Questions are the decisions an interface has to ask before adopt --init
// runs, for a repository with these facts: the line endings (unless a
// .gitattributes exists), the CI provider, and a pipeline of the person's own
// (only when the repository has one). The agent files are not here: the
// wizard offers them in a step of their own after the adoption.
//
// Parameters: f are the facts of the repository.
// Returns: the questions in the order to ask them.
func Questions(f Facts) []Question {
	var qs []Question
	if !f.GitattributesExists {
		qs = append(qs, Question{Key: KeyGitattributes,
			Prompt: "Line endings: should smith write a .gitattributes with `* text=auto eol=lf`? Git then converts every text file in the repository to LF.",
			Answers: []Answer{
				{Label: "No, keep my line endings", Value: GitattributesSkip, Hint: "nothing is renormalised (default)"},
				{Label: "Yes, write it", Value: GitattributesAdd, Hint: "text files become LF; on a CRLF repository this changes many files"},
			}})
	}
	label, hint := "GitHub Actions", "no pipeline was found; GitHub is the usual choice"
	if len(f.Detected) > 0 {
		label, hint = strings.Join(f.Detected, ", "), "the repository already uses it"
	}
	answers := []Answer{{Label: label, Value: DefaultCI, Hint: hint + " (default)"}}
	fallback := proposedCI(f)
	for _, p := range Providers {
		if !slices.Contains(fallback, p) {
			answers = append(answers, Answer{Label: p, Value: p, Hint: "smith writes its pipeline for " + p})
		}
	}
	answers = append(answers, Answer{Label: "None", Value: NoCI, Hint: "smith writes no pipeline"})
	qs = append(qs, Question{Key: KeyCI,
		Prompt:  "Which CI should smith write the pipelines for? It also writes .devcontainer and tools/smith/install.sh.",
		Answers: answers})
	if len(f.OwnPipelines) > 0 {
		var paths []string
		for _, p := range Providers {
			if path := f.OwnPipelines[p]; path != "" {
				paths = append(paths, path)
			}
		}
		qs = append(qs, Question{Key: KeyExisting,
			Prompt: "You already have a pipeline: " + strings.Join(paths, ", ") + ". What should happen to it?",
			Answers: []Answer{
				{Label: "Replace it, keep a copy", Value: ExistingReplace, Hint: "smith writes its pipeline; yours moves to <file>.pre-smith (default)"},
				{Label: "Keep mine", Value: ExistingKeep, Hint: "smith writes no pipeline for that provider"},
			}})
	}
	return qs
}

// proposedCI are the providers smith writes for when the person takes the default.
func proposedCI(f Facts) []string {
	if len(f.Detected) > 0 {
		return f.Detected
	}
	return []string{"github"}
}

// FromAnswers turns the answers of Questions into settled choices. A question
// that was not asked (or not answered) takes its safe default. The agent files
// are skipped: the wizard offers them later.
//
// Parameters: f are the facts; answers map a question key to the Value chosen.
// Returns: the choices.
func FromAnswers(f Facts, answers map[string]string) Choices {
	c := Choices{Gitattributes: GitattributesSkip, Agent: AgentSkip, Existing: ExistingReplace}
	if v := answers[KeyGitattributes]; v == GitattributesAdd {
		c.Gitattributes = GitattributesAdd
	}
	switch v := answers[KeyCI]; v {
	case NoCI:
	case "", DefaultCI:
		c.CI = proposedCI(f)
	default:
		c.CI = []string{v}
	}
	if answers[KeyExisting] == ExistingKeep {
		c.Existing = ExistingKeep
	}
	return c
}

// Argv is the flags of adopt --init that carry the choices, so a command that
// runs with --yes takes exactly these and not its own defaults.
func (c Choices) Argv() []string {
	argv := []string{"--gitattributes", cmpOr(c.Gitattributes, GitattributesSkip), "--agent", cmpOr(c.Agent, AgentSkip)}
	if len(c.CI) == 0 {
		argv = append(argv, "--ci", NoCI)
	} else {
		argv = append(argv, "--ci", strings.Join(c.CI, ","))
	}
	return append(argv, "--ci-existing", cmpOr(c.Existing, ExistingReplace))
}

func cmpOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
