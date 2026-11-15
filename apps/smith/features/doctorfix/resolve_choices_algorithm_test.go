package doctorfix

import (
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestResolveRefusesAnOptionThatIsDisabledAndSaysWhy(t *testing.T) {
	fix := protocol.Fix{Description: "get the tool", Choices: []protocol.FixChoice{{Question: "How?", Options: []protocol.FixChoiceOption{
		{Label: "install", Fix: protocol.Fix{Steps: []protocol.FixStep{{Argv: []string{"go", "install"}}}}},
		{Label: "update", Disabled: "smith does not know how this copy was installed"},
	}}}}
	if got, _, err := Resolve(fix, []int{1}); err != nil || len(got.Steps) != 1 {
		t.Fatalf("an enabled option still resolves: %+v %v", got, err)
	}
	got, _, err := Resolve(fix, []int{2})
	if err == nil || !strings.Contains(err.Error(), "disabled: smith does not know how this copy was installed") || len(got.Steps) != 0 {
		t.Fatalf("a disabled option adds nothing and names the reason: %+v %v", got, err)
	}
}
