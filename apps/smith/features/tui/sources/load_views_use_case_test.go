package sources

import (
	"reflect"
	"testing"
)

func TestATaskRunAsksForTheEventStreamAndOtherCommandsDoNot(t *testing.T) {
	for argv, want := range map[string][]string{
		"run build -p web":       {"run", "build", "-p", "web", "--json-events"},
		"test":                   {"test", "--json-events"},
		"run lint --json":        {"run", "lint", "--json"},
		"run lint --json-events": {"run", "lint", "--json-events"},
		"plugin list":            {"plugin", "list"},
		"doctor":                 {"doctor"},
	} {
		in := splitWords(argv)
		got := withEventStream(in)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %v, want %v", argv, got, want)
		}
	}
	in := []string{"run", "build"}
	_ = withEventStream(in)
	if len(in) != 2 {
		t.Fatalf("the caller's command line was changed: %v", in)
	}
}

func splitWords(s string) []string {
	var out []string
	word := ""
	for _, r := range s {
		if r == ' ' {
			out, word = append(out, word), ""
			continue
		}
		word += string(r)
	}
	return append(out, word)
}
