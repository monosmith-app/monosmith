package projecttooling

import "testing"

func TestParseAnswersReadsOneForAllAndOnePerProject(t *testing.T) {
	a, err := ParseAnswers([]string{"keep", "web=smith"})
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := a.For("web"); !ok || got != Smith {
		t.Errorf("web: %q %v", got, ok)
	}
	if got, ok := a.For("api"); !ok || got != Keep {
		t.Errorf("api takes the answer for all: %q %v", got, ok)
	}
}

func TestParseAnswersWithoutAnAnswerForAProjectSaysSo(t *testing.T) {
	a, err := ParseAnswers([]string{"web=keep"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.For("api"); ok {
		t.Error("api has no answer")
	}
}

func TestParseAnswersRefusesWhatIsNotSmithOrKeep(t *testing.T) {
	for _, v := range []string{"full", "web=yes", "=smith"} {
		if _, err := ParseAnswers([]string{v}); err == nil {
			t.Errorf("%q was accepted", v)
		}
	}
}
