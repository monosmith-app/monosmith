package scaffold

import (
	"encoding/json"
	"testing"
)

func TestNewQuestionsJSONListsEveryQuestionWithItsAnswers(t *testing.T) {
	raw, err := NewQuestionsJSON()
	if err != nil {
		t.Fatal(err)
	}
	var got []newQuestionJSON
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(NewQuestions()) {
		t.Fatalf("got %d questions, want %d", len(got), len(NewQuestions()))
	}
	byKey := map[string]newQuestionJSON{}
	for _, q := range got {
		byKey[q.Key] = q
	}
	ci := byKey[AnswerCI]
	if !ci.Many || ci.Default != "github" || ci.Flag != "ci" || len(ci.Options) != 5 {
		t.Errorf("ci question = %+v", ci)
	}
	if byKey[AnswerNodeStack].When[AnswerLanguage] != "node" {
		t.Errorf("node stack should apply only to node: %+v", byKey[AnswerNodeStack])
	}
	if byKey[AnswerReleaseBranch].When[AnswerMore] != MoreYes || len(byKey[AnswerReleaseBranch].Options) != 0 {
		t.Errorf("release branch is free text behind more options: %+v", byKey[AnswerReleaseBranch])
	}
	layout := byKey[AnswerArchitecture]
	if layout.Default != DefaultArchitecture || layout.Options[0].Value != DefaultArchitecture {
		t.Errorf("layout default is not first: %+v", layout)
	}
}
