package scaffold

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/choiceprompt"
	"github.com/monosmith-app/monosmith/apps/smith/tui/theme"
)

// maxAnswerTries is how many times one question is asked before its default
// is taken.
const maxAnswerTries = 3

// askNewQuestions asks, on out, each question of smith new that has no answer
// yet and that applies, and adds the answers to answers. Enter takes the
// default; so does the end of the input, for that question and the rest, so a
// closed terminal never blocks.
func askNewQuestions(answers map[string]string, in *bufio.Reader, out io.Writer) {
	closed := false
	for _, q := range NewQuestions() {
		if _, given := answers[q.Key]; given {
			continue
		}
		if q.AtEnd || (q.Applies != nil && !q.Applies(answers)) {
			continue
		}
		if closed {
			answers[q.Key] = q.Default
			continue
		}
		answers[q.Key], closed = askNewQuestion(q, in, out)
	}
}

// askNewQuestion asks one question and says whether the input ended.
func askNewQuestion(q NewQuestion, in *bufio.Reader, out io.Writer) (answer string, ended bool) {
	fmt.Fprintf(out, "\n%s\n", q.Prompt)
	def := q.Default
	if q.Options != nil {
		labelWidth := 0
		for _, o := range q.Options {
			labelWidth = max(labelWidth, len(o.Label))
		}
		for i, o := range q.Options {
			line := fmt.Sprintf("  %d) %-*s", i+1, labelWidth, o.Label)
			if o.Hint != "" {
				line += "  " + o.Hint
			}
			fmt.Fprintln(out, strings.TrimRight(line, " "))
		}
		for i, o := range q.Options {
			if o.Value == q.Default {
				def = strconv.Itoa(i + 1)
			}
		}
	}
	for attempt := 0; attempt < maxAnswerTries; attempt++ {
		fmt.Fprintf(out, "%s [%s]: ", answerLabel(q), def)
		line, err := in.ReadString('\n')
		value, problem := resolveAnswer(q, strings.TrimSpace(line))
		if problem == nil {
			return value, false
		}
		if err != nil {
			return q.Default, true
		}
		fmt.Fprintf(out, "%v\n", problem)
	}
	fmt.Fprintf(out, "Using the default, %s.\n", q.Default)
	return q.Default, false
}

func answerLabel(q NewQuestion) string {
	if q.Options != nil {
		return "Choice"
	}
	return "Answer"
}

// resolveAnswer turns what a person typed into the answer's value: empty is
// the default, a number picks an option, and a value or a label (any case) is
// taken as it is.
func resolveAnswer(q NewQuestion, typed string) (string, error) {
	if typed == "" {
		return q.Default, nil
	}
	if q.Options == nil {
		if q.Check != nil {
			if err := q.Check(typed); err != nil {
				return "", err
			}
		}
		return typed, nil
	}
	if n, err := strconv.Atoi(typed); err == nil && n >= 1 && n <= len(q.Options) {
		return q.Options[n-1].Value, nil
	}
	for _, o := range q.Options {
		if strings.EqualFold(typed, o.Value) || strings.EqualFold(typed, o.Label) {
			return o.Value, nil
		}
	}
	return "", fmt.Errorf("type a number from 1 to %d, or Enter for the default", len(q.Options))
}

// describeChoices prints what was chosen, as a short table, before anything is
// written.
func describeChoices(opts NewOptions, out io.Writer) {
	stack := opts.DefaultStacks["node"]
	if stack == "" {
		stack = DefaultNodeStack
	}
	language := "decide later"
	if opts.Language != "" {
		language = opts.Language
	}
	architecture := opts.Architecture
	if architecture == "" {
		architecture = DefaultArchitecture
	}
	rows := [][2]string{
		{"CI", strings.Join(opts.Providers, ", ")},
		{"Main language", language},
	}
	if opts.Language == "node" {
		rows = append(rows, [2]string{"Node tests and lint", stack})
	}
	rows = append(rows,
		[2]string{"Code layout", architecture},
		[2]string{"Releases", describeReleases(opts)},
		[2]string{"Cache", describeCache(opts)},
	)
	fmt.Fprintln(out, "\nYour choices")
	for _, r := range rows {
		fmt.Fprintf(out, "  %-20s %s\n", r[0], r[1])
	}
}

func describeReleases(opts NewOptions) string {
	text := fmt.Sprintf("from %s, tags pushed to %s", orDefault(opts.ReleaseBranch, DefaultReleaseBranch), orDefault(opts.ReleaseRemote, DefaultReleaseRemote))
	if opts.GitHubReleases {
		text += ", with a GitHub Release"
	}
	return text
}

func describeCache(opts NewOptions) string {
	size := fmt.Sprintf("up to %d MB", DefaultCacheMaxMB)
	switch {
	case opts.CacheMaxMB < 0:
		size = "no size limit"
	case opts.CacheMaxMB > 0:
		size = fmt.Sprintf("up to %d MB", opts.CacheMaxMB)
	}
	return orDefault(opts.CacheDir, DefaultCacheDir) + ", " + size
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// newBanner opens the questions in the style of smith's other steps: the
// same rule, and the ASCII one where the terminal cannot draw it.
func newBanner(name string, icons theme.Icons) string {
	rule := "== New workspace: " + name + " =="
	if icons.Passed != theme.ASCII.Passed {
		rule = "══ New workspace: " + name + " "
		rule += strings.Repeat("═", max(60-len([]rune(rule)), 3))
	}
	return rule + "\nA few questions. Enter takes the default shown; smith new --yes skips them."
}

// AskChoice asks one question with numbered answers on out and reads the
// choice from in. Enter takes def. A closed input takes def as well, so a
// pipe never blocks. Other packages use it so every smith question looks and
// answers the same way.
//
// Parameters: prompt is the question; options are the answers, each with a
// Value, a Label and a Hint; def is the Value Enter takes.
// Returns: the Value of the answer.
func AskChoice(prompt string, options []NewOption, def string, in *bufio.Reader, out io.Writer) string {
	return choiceprompt.Ask(prompt, options, def, in, out)
}
