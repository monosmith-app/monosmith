package adoptdecision

import (
	"reflect"
	"testing"
)

func keys(qs []Question) []string {
	var out []string
	for _, q := range qs {
		out = append(out, q.Key)
	}
	return out
}

func TestQuestionsAskOnlyWhatTheRepositoryLeavesOpen(t *testing.T) {
	for _, tc := range []struct {
		name  string
		facts Facts
		want  []string
	}{
		{"a bare repository", Facts{}, []string{KeyGitattributes, KeyCI}},
		{"it has a .gitattributes", Facts{GitattributesExists: true}, []string{KeyCI}},
		{"it has a pipeline of its own", Facts{OwnPipelines: map[string]string{"github": ".github/workflows/ci.yml"}}, []string{KeyGitattributes, KeyCI, KeyExisting}},
	} {
		if got := keys(Questions(tc.facts)); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: got %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestEveryQuestionHasASafeDefaultFirstAndSaysWhatEachAnswerDoes(t *testing.T) {
	for _, q := range Questions(Facts{OwnPipelines: map[string]string{"github": "x"}, Detected: []string{"gitlab"}}) {
		if len(q.Answers) < 2 {
			t.Errorf("%s: a choice needs two answers", q.Key)
		}
		for _, a := range q.Answers {
			if a.Hint == "" || a.Label == "" {
				t.Errorf("%s: answer %q says nothing about what it does", q.Key, a.Value)
			}
		}
	}
	gitlab := Questions(Facts{Detected: []string{"gitlab"}})[1]
	if gitlab.Answers[0].Value != DefaultCI || gitlab.Answers[0].Label != "gitlab" {
		t.Errorf("the default is the provider the repository uses: %+v", gitlab.Answers[0])
	}
	for _, a := range gitlab.Answers[1:] {
		if a.Value == "gitlab" {
			t.Error("the detected provider is not offered twice")
		}
	}
}

func TestFromAnswers(t *testing.T) {
	f := Facts{Detected: []string{"gitlab"}, OwnPipelines: map[string]string{"gitlab": ".gitlab-ci.yml"}}
	if got := FromAnswers(f, nil); got.Gitattributes != GitattributesSkip || got.Agent != AgentSkip || got.Existing != ExistingReplace || !reflect.DeepEqual(got.CI, []string{"gitlab"}) {
		t.Errorf("nothing answered takes the safe defaults: %+v", got)
	}
	got := FromAnswers(f, map[string]string{KeyGitattributes: GitattributesAdd, KeyCI: "azure", KeyExisting: ExistingKeep})
	if got.Gitattributes != GitattributesAdd || !reflect.DeepEqual(got.CI, []string{"azure"}) || got.Existing != ExistingKeep {
		t.Errorf("%+v", got)
	}
	if got := FromAnswers(f, map[string]string{KeyCI: NoCI}); len(got.CI) != 0 {
		t.Errorf("none writes no pipeline: %+v", got)
	}
}

func TestArgvCarriesEveryChoiceSoYesTakesNoDefaultOfItsOwn(t *testing.T) {
	c := Choices{Gitattributes: GitattributesAdd, Agent: AgentSkip, CI: []string{"gitlab", "github"}, Existing: ExistingKeep}
	want := []string{"--gitattributes", "add", "--agent", "skip", "--ci", "gitlab,github", "--ci-existing", "keep"}
	if got := c.Argv(); !reflect.DeepEqual(got, want) {
		t.Errorf("got %v", got)
	}
	if got := (Choices{}).Argv(); !reflect.DeepEqual(got, []string{"--gitattributes", "skip", "--agent", "skip", "--ci", "none", "--ci-existing", "replace"}) {
		t.Errorf("an empty choice is the safe one: %v", got)
	}
}
