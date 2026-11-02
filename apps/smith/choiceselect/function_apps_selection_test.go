package choiceselect

import (
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// functionOffers shapes the function-app choices of ADR 0072 for node and python.
func functionOffers() []Offer {
	scaffold := []protocol.Role{protocol.RoleScaffold}
	var out []Offer
	for _, lang := range []string{"node", "python"} {
		plugin := map[string]string{"node": "functions", "python": "python"}[lang]
		add := func(axis, name string, mutate func(*Offer)) {
			out = append(out, offer(plugin, axis, name, func(o *Offer) {
				o.Language = lang
				if mutate != nil {
					mutate(o)
				}
			}))
		}
		for _, host := range []string{"azure-functions", "aws-lambda", "gcp-functions"} {
			add("framework", host, func(o *Offer) { o.Roles, o.Needs = scaffold, []string{"trigger", "runtime"} })
		}
		add("trigger", "http-trigger", func(o *Offer) { o.Suggested = true })
		add("trigger", "timer-trigger", func(o *Offer) { o.Conflicts = []string{"gcp-functions"} })
		add("runtime", map[string]string{"node": "node24", "python": "python313"}[lang], func(o *Offer) { o.Suggested = true })
		add("runtime", map[string]string{"node": "node22", "python": "python312"}[lang], nil)
	}
	return out
}

func TestTheSameNameIsOfferedPerLanguage(t *testing.T) {
	offers := functionOffers()
	sel, err := SolveIn(offers, "python", []string{"azure-functions", "http-trigger", "python313"})
	if err != nil {
		t.Fatal(err)
	}
	if st := sel.Stack("x", ""); st.Language != "python" || st.Scaffold != "python" {
		t.Errorf("stack %+v", st)
	}
	_, err = Solve(offers, []string{"azure-functions", "http-trigger", "python313"})
	if err == nil || !strings.Contains(err.Error(), "azure-functions is offered for node, python: add --lang") {
		t.Errorf("without a language the name is ambiguous: %v", err)
	}
}

func TestAHostWithoutAnAnswerAsksInsteadOfChoosing(t *testing.T) {
	_, err := SolveIn(functionOffers(), "node", []string{"azure-functions"})
	var need *NeedsAxisError
	if !errors.As(err, &need) || need.Axis != "trigger" || need.Choice != "azure-functions" {
		t.Fatalf("got %v", err)
	}
	if got, ok := need.Suggested(); !ok || got.Name != "http-trigger" {
		t.Errorf("suggested %v", got)
	}
	if !strings.Contains(err.Error(), "add one of http-trigger (suggested), timer-trigger to --with") {
		t.Errorf("the refusal must name the options: %v", err)
	}
	_, err = SolveIn(functionOffers(), "node", []string{"azure-functions", "http-trigger"})
	if !errors.As(err, &need) || need.Axis != "runtime" {
		t.Fatalf("the runtime is asked next: %v", err)
	}
}

func TestAnOptionThatConflictsWithTheHostIsNotOffered(t *testing.T) {
	_, err := SolveIn(functionOffers(), "node", []string{"gcp-functions"})
	var need *NeedsAxisError
	if !errors.As(err, &need) {
		t.Fatal(err)
	}
	if len(need.Options) != 1 || need.Options[0].Name != "http-trigger" {
		t.Errorf("timer-trigger conflicts with gcp-functions: %+v", need.Options)
	}
	_, err = SolveIn(functionOffers(), "node", []string{"gcp-functions", "timer-trigger", "node24"})
	if err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Errorf("got %v", err)
	}
}

func TestTwoHostsOrAHostAndAFrameworkAreRefused(t *testing.T) {
	offers := append(functionOffers(), webOffers()...)
	for _, names := range [][]string{
		{"azure-functions", "aws-lambda", "http-trigger", "node24"},
		{"azure-functions", "react", "http-trigger", "node24"},
	} {
		_, err := SolveIn(offers, "node", names)
		if err == nil || !strings.Contains(err.Error(), "both on the framework axis: keep one") {
			t.Errorf("%v: got %v", names, err)
		}
	}
}

func TestProblemsAreKeyedByLanguageAndName(t *testing.T) {
	got := Problems(functionOffers(), func(l string) bool { return l == "node" })
	if got["python/azure-functions"] == "" || got["node/azure-functions"] != "" {
		t.Errorf("%v", got)
	}
}
