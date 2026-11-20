package manifestoffer

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

var finds = []Find{
	{Language: "node", Manifest: "package.json", Files: []string{"a.ts", "b.ts"}},
	{Language: "python", Manifest: "pyproject.toml", Files: []string{"x.py"}},
}

func TestSettleTakesTheFlag(t *testing.T) {
	f, ok, err := Settle(Question{Folder: "../x", Finds: finds, Answer: "python"})
	if err != nil || !ok || f.Language != "python" {
		t.Fatalf("got %+v %v %v, want python", f, ok, err)
	}
	if _, ok, err := Settle(Question{Folder: "../x", Finds: finds, Answer: "none"}); ok || err != nil {
		t.Fatalf("--manifest none gave %v %v, want no manifest and no error", ok, err)
	}
}

func TestSettleRefusesAFlagThatIsNotAChoice(t *testing.T) {
	_, _, err := Settle(Question{Folder: "../x", Finds: finds, Answer: "go"})
	if err == nil || !strings.Contains(err.Error(), "--manifest go is not a choice") {
		t.Fatalf("got %v", err)
	}
}

func TestSettleNeverPicksWithoutATerminal(t *testing.T) {
	_, ok, err := Settle(Question{Folder: "../x", Finds: finds, Command: "smith import ../x"})
	if ok || err == nil {
		t.Fatalf("got %v %v, want a refusal", ok, err)
	}
	if !strings.Contains(err.Error(), "smith import ../x --manifest node") {
		t.Errorf("the refusal does not name the flag: %v", err)
	}
}

func TestSettleAsksInATerminal(t *testing.T) {
	var out bytes.Buffer
	f, ok, err := Settle(Question{Folder: "../x", Finds: finds, Ask: true, In: bufio.NewReader(strings.NewReader("2\n")), Out: &out})
	if err != nil || !ok || f.Language != "python" {
		t.Fatalf("got %+v %v %v, want python", f, ok, err)
	}
	if !strings.Contains(out.String(), "write package.json (node)") || !strings.Contains(out.String(), "none") {
		t.Errorf("the question does not list the choices:\n%s", out.String())
	}
	if _, ok, _ := Settle(Question{Folder: "../x", Finds: finds, Ask: true, In: bufio.NewReader(strings.NewReader("3\n")), Out: &out}); ok {
		t.Error("the answer none wrote a manifest")
	}
}
