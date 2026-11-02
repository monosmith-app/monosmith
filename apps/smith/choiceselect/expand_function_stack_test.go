package choiceselect

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAStackThatNamesItsLanguageExpandsAnAmbiguousName(t *testing.T) {
	st := protocol.Stack{Name: "api", Language: "python", Choices: []string{"azure-functions", "http-trigger", "python313"}}
	got, err := ExpandStack(st, functionOffers())
	if err != nil || got.Language != "python" || got.Scaffold != "python" {
		t.Fatalf("got %+v, %v", got, err)
	}
	st.Language = ""
	if _, err := ExpandStack(st, functionOffers()); err == nil || !strings.Contains(err.Error(), "add --lang") {
		t.Errorf("without a language the stack is ambiguous: %v", err)
	}
}
