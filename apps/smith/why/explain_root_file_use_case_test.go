package why

import (
	"strings"
	"testing"
)

func TestARootFileSmithAddWritesIsRecognised(t *testing.T) {
	e := Explain(testInputs(t), "tsconfig.base.json")
	if !e.Known || e.Owner != OwnerYou || !strings.Contains(e.Summary, "Written once by smith add") {
		t.Fatalf("got %+v", e)
	}
}
