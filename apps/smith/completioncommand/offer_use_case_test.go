package completioncommand

import (
	"bufio"
	"github.com/monosmith-app/monosmith/apps/smith/shellcompletion"
	"github.com/monosmith-app/monosmith/apps/smith/shellenv"
	"strings"
	"testing"
	"time"
)

// answerStore is a Memory kept in a variable.
type answerStore struct{ a Answer }

func (s *answerStore) Load() Answer        { return s.a }
func (s *answerStore) Save(a Answer) error { s.a = a; return nil }

// offerWith runs the offer with typed answers and returns what it printed.
func offerWith(t *testing.T, env shellenv.Environment, mem Memory, typed string, now time.Time) string {
	t.Helper()
	var out strings.Builder
	err := Offer(OfferOptions{
		Env: env, Memory: mem, Now: now,
		Terminal: Terminal{In: bufio.NewReader(strings.NewReader(typed)), Out: &out},
		Script:   func(shellenv.Shell) (string, error) { return testScript, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	return out.String()
}

func bashEnv(t *testing.T) *tempEnv {
	env := newTempEnv(t, "linux")
	env.parent = "bash"
	return env
}

func TestTheOfferShowsTheFilesAndYesWritesThemAndIsRemembered(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	out := offerWith(t, env, mem, "1\n", time.Now())
	if !strings.Contains(out, "Set up tab completion for Bash?") || !strings.Contains(out, "smith.bash") {
		t.Fatalf("the offer does not name the shell and the file:\n%s", out)
	}
	if mem.a.Choice != ChoiceInstalled || !shellcompletion.HasBlock(env.read(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path)) {
		t.Fatalf("not installed or not remembered: %+v", mem.a)
	}
}

func TestNoIsRememberedAndNeverAskedAgain(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	offerWith(t, env, mem, "2\n", time.Now())
	if mem.a.Choice != ChoiceNo {
		t.Fatalf("%+v", mem.a)
	}
	if env.Exists(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path) {
		t.Fatal("a file was written after No")
	}
	if out := offerWith(t, env, mem, "1\n", time.Now().Add(365*24*time.Hour)); out != "" {
		t.Fatalf("asked again after No:\n%s", out)
	}
}

func TestAskLaterWaitsAndThenAsksAgain(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	now := time.Now()
	offerWith(t, env, mem, "3\n", now)
	if mem.a.Choice != ChoiceLater || !mem.a.AskAfter.After(now) {
		t.Fatalf("%+v", mem.a)
	}
	if out := offerWith(t, env, mem, "3\n", now.Add(24*time.Hour)); out != "" {
		t.Fatalf("asked before the time:\n%s", out)
	}
	if out := offerWith(t, env, mem, "3\n", now.Add(15*24*time.Hour)); !strings.Contains(out, "Set up tab completion") {
		t.Fatalf("did not ask after the time:\n%s", out)
	}
}

func TestEnterAloneTakesAskMeLaterAndWritesNothing(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	offerWith(t, env, mem, "\n", time.Now())
	if mem.a.Choice != ChoiceLater || env.Exists(shellcompletion.Targets(env, shellenv.Bash, "")[0].Path) {
		t.Fatalf("%+v", mem.a)
	}
}

func TestShowMeTheScriptPrintsItAndAsksAgain(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	out := offerWith(t, env, mem, "4\n2\n", time.Now())
	if !strings.Contains(out, "line four") || mem.a.Choice != ChoiceNo {
		t.Fatalf("%+v\n%s", mem.a, out)
	}
}

func TestNoOfferWhenCompletionIsThereOrTheShellIsUnknown(t *testing.T) {
	env, mem := bashEnv(t), &answerStore{}
	if _, err := shellcompletion.Install(env, shellcompletion.PlanFor(env, shellenv.Bash, testScript)); err != nil {
		t.Fatal(err)
	}
	if out := offerWith(t, env, mem, "1\n", time.Now()); out != "" {
		t.Fatalf("offered when installed:\n%s", out)
	}
	unknown := newTempEnv(t, "linux")
	if out := offerWith(t, unknown, &answerStore{}, "1\n", time.Now()); out != "" {
		t.Fatalf("offered for an unknown shell:\n%s", out)
	}
}

func TestHomeOfferStaysAfterAskLaterButNotAfterNoOrInstall(t *testing.T) {
	env := bashEnv(t)
	for answer, want := range map[string]bool{"": true, ChoiceLater: true, ChoiceNo: false, ChoiceInstalled: false} {
		if got := HomeOfferWanted(env, Answer{Choice: answer}); got != want {
			t.Errorf("%q: got %v", answer, got)
		}
	}
}
