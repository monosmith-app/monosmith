package scaffold

import (
	"bufio"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/choiceselect"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func triggerNeed() *choiceselect.NeedsAxisError {
	offer := func(name string, suggested bool) choiceselect.Offer {
		return choiceselect.Offer{Choice: protocol.Choice{Axis: "trigger", Name: name, Description: name + " text", Suggested: suggested}}
	}
	return &choiceselect.NeedsAxisError{Choice: "azure-functions", Axis: "trigger", Options: []choiceselect.Offer{offer("http-trigger", true), offer("timer-trigger", false)}}
}

func TestAskAxisAnswers(t *testing.T) {
	for name, tc := range map[string]struct{ input, want string }{
		"empty line takes the suggested one": {"\n", "http-trigger"},
		"a number":                           {"2\n", "timer-trigger"},
		"a name":                             {"timer-trigger\n", "timer-trigger"},
		"a bad answer is asked again":        {"nope\n2\n", "timer-trigger"},
	} {
		t.Run(name, func(t *testing.T) {
			var out strings.Builder
			got, err := askAxis(bufio.NewReader(strings.NewReader(tc.input)), &out, triggerNeed())
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v", got, err)
			}
			if !strings.Contains(out.String(), "http-trigger (suggested)") || !strings.Contains(out.String(), "does not choose it for you") {
				t.Errorf("the question must mark the suggested option and say nothing is chosen silently:\n%s", out.String())
			}
		})
	}
}

func TestAskAxisRefusesAnEndedInput(t *testing.T) {
	var out strings.Builder
	_, err := askAxis(bufio.NewReader(strings.NewReader("")), &out, triggerNeed())
	if err == nil || !strings.Contains(err.Error(), "add one of http-trigger, timer-trigger to --with") {
		t.Fatalf("got %v", err)
	}
}
