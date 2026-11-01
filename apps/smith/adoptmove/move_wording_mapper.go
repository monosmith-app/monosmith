package adoptmove

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/problem"
)

// Explanation says why the project has to move. The prompt, the
// non-interactive refusal, the dry run and the docs all use these words.
func Explanation(c Candidate) string {
	return fmt.Sprintf("The project (%s, %q) is at the repository root. smith manages one folder per project, so it has to move into a folder of its own, and smith does not choose the folder for you.", c.Manifest, c.Name)
}

// Question is the question the prompts ask.
const Question = "Move it into a folder of its own so smith can manage it?"

// QuestionText is the question with its numbered choices, the suggested one marked.
func QuestionText(c Candidate) string {
	choices := Choices(c)
	var b strings.Builder
	b.WriteString(Question + "\n")
	for i, ch := range choices {
		mark := ""
		if ch.Default {
			mark = " (suggested)"
		}
		fmt.Fprintf(&b, "  %d) %s%s\n", i+1, ch.Label, mark)
	}
	return b.String()
}

// Command is the command that moves the project into dir and adopts the repository.
func Command(dir string) string { return "smith adopt --init --move " + dir }

// NeedsAnswer is the refusal of a run that cannot ask (no terminal, or --yes)
// and was not given --move. It changes nothing; the problem names the exact
// commands.
func NeedsAnswer(c Candidate) error {
	choices := Choices(c)
	suggested := choices[DefaultChoice(choices)-1]
	return problem.New(Explanation(c),
		Command(suggested.Dir)+"   (the suggested folder; any other folder works the same way)",
		Command("<folder>")+"   (for example "+Command("packages/"+c.Name)+")",
		"smith adopt --init --move "+Leave+"   (keep it where it is: an empty workspace, add projects with smith add)",
	)
}
