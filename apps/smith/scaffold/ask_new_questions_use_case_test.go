package scaffold

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func ask(t *testing.T, input string, preset map[string]string) (map[string]string, string) {
	t.Helper()
	answers := map[string]string{}
	for k, v := range preset {
		answers[k] = v
	}
	var out bytes.Buffer
	askNewQuestions(answers, bufio.NewReader(strings.NewReader(input)), &out)
	return answers, out.String()
}

func TestEnterTakesEveryDefaultAndSkipsTheAdvancedQuestions(t *testing.T) {
	answers, out := ask(t, "\n\n\n\n\n", nil)
	for key, want := range map[string]string{AnswerCI: "github", AnswerLanguage: LanguageLater, AnswerArchitecture: DefaultArchitecture, AnswerMore: "no"} {
		if answers[key] != want {
			t.Errorf("%s = %q, want %q", key, answers[key], want)
		}
	}
	if _, asked := answers[AnswerReleaseBranch]; asked || strings.Contains(out, "release tags") {
		t.Errorf("the advanced questions need 'more options':\n%s", out)
	}
	if _, asked := answers[AnswerNodeStack]; asked {
		t.Error("the Node question is for Node")
	}
}

func TestNodeAsksItsStackAndMoreOptionsAsksTheRest(t *testing.T) {
	// ci azure, language node (3rd... by number), stack vitest, layout 2, more options, then the advanced ones.
	answers, out := ask(t, "azure\nnode\n2\nhexagonal\n\n2\ntrunk\n\nyes\n.cache\n0\n", nil)
	want := map[string]string{
		AnswerCI: "azure", AnswerLanguage: "node", AnswerNodeStack: "node-vitest", AnswerArchitecture: "hexagonal",
		AnswerMore: MoreYes, AnswerReleaseBranch: "trunk", AnswerReleaseRemote: "origin",
		AnswerGitHubReleases: "yes", AnswerCacheDir: ".cache", AnswerCacheMaxMB: "0",
	}
	for key, v := range want {
		if answers[key] != v {
			t.Errorf("%s = %q, want %q\n%s", key, answers[key], v, out)
		}
	}
}

func TestAQuestionWithAFlagIsNotAsked(t *testing.T) {
	answers, out := ask(t, "\n\n\n", map[string]string{AnswerCI: "gitlab", AnswerMore: "no"})
	if answers[AnswerCI] != "gitlab" || strings.Contains(out, "Which CI") {
		t.Errorf("asked what the flag answered:\n%s", out)
	}
}

func TestAWrongAnswerIsExplainedAndTheDefaultTakenAfterThreeTries(t *testing.T) {
	answers, out := ask(t, "9\nx\ny\n", map[string]string{AnswerLanguage: "go", AnswerArchitecture: "none", AnswerMore: "no"})
	if answers[AnswerCI] != "github" {
		t.Errorf("ci = %q", answers[AnswerCI])
	}
	if !strings.Contains(out, "type a number from 1 to 5") || !strings.Contains(out, "Using the default, github.") {
		t.Errorf("\n%s", out)
	}
}

func TestAClosedInputKeepsTheDefaults(t *testing.T) {
	answers, _ := ask(t, "", nil)
	if answers[AnswerCI] != "github" || answers[AnswerMore] != "no" {
		t.Errorf("%v", answers)
	}
}

func TestEveryQuestionHasAFlagAndADefaultAmongItsOptions(t *testing.T) {
	for _, q := range NewQuestions() {
		if q.Flag == "" && q.Key != AnswerMore && q.Key != AnswerSaveRecipe { // the yes/no is answered by --save-recipe <name>
			t.Errorf("%s has no flag", q.Key)
		}
		if q.Options != nil && !hasOption(q.Options, q.Default) {
			t.Errorf("%s: default %q is not an option", q.Key, q.Default)
		}
		if q.Options != nil && q.Options[0].Value != q.Default {
			t.Errorf("%s: the default comes first", q.Key)
		}
	}
}
