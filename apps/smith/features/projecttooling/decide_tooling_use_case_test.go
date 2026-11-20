package projecttooling

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

var legacyWeb = Assessment{Name: "web", Node: true, Legacy: []string{"an ESLint 8 configuration (.eslintrc.js)"}, Recommended: Smith}
var currentAPI = Assessment{Name: "api", Node: true, Recommended: Keep}

func TestDecideTakesTheFlagsAnswer(t *testing.T) {
	got, err := Decide(DecideRequest{Projects: []Assessment{legacyWeb, currentAPI, {Name: "py"}}, Given: Answers{All: Smith}, Out: &bytes.Buffer{}})
	if err != nil {
		t.Fatal(err)
	}
	if got["web"] != Smith || got["api"] != Smith || len(got) != 2 {
		t.Errorf("answers %v (a project that is not Node gets none)", got)
	}
}

func TestDecideWithYesKeepsTheToolsAndSaysHowToChooseSmiths(t *testing.T) {
	var out bytes.Buffer
	got, err := Decide(DecideRequest{Projects: []Assessment{legacyWeb}, Yes: true, Out: &out})
	if err != nil {
		t.Fatal(err)
	}
	if got["web"] != Keep {
		t.Errorf("--yes never picks smith's tooling: %v", got)
	}
	if !strings.Contains(out.String(), "--tooling web=smith") || !strings.Contains(out.String(), ".eslintrc.js") {
		t.Errorf("output %q", out.String())
	}
}

func TestDecideWithoutTerminalOrYesRefusesAndNamesTheFlag(t *testing.T) {
	_, err := Decide(DecideRequest{Projects: []Assessment{legacyWeb}, Out: &bytes.Buffer{}})
	if err == nil || !strings.Contains(err.Error(), "--tooling smith|keep") {
		t.Fatalf("err = %v", err)
	}
}

func TestDecideAsksAndEnterTakesTheRecommendation(t *testing.T) {
	var out bytes.Buffer
	got, err := Decide(DecideRequest{Projects: []Assessment{legacyWeb, currentAPI}, Interactive: true, In: bufio.NewReader(strings.NewReader("\n\n")), Out: &out})
	if err != nil {
		t.Fatal(err)
	}
	if got["web"] != Smith || got["api"] != Keep {
		t.Errorf("answers %v", got)
	}
	if !strings.Contains(out.String(), "(recommended)") {
		t.Errorf("the question recommends: %q", out.String())
	}
}
