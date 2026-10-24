package commandargs

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestExactlyNamesWhatIsMissing(t *testing.T) {
	root := &cobra.Command{Use: "smith"}
	add := &cobra.Command{Use: "add"}
	root.AddCommand(add)
	check := Exactly(Spec{Names: []string{"a kind (app, lib or package)", "a name"}, Example: "smith add app web --lang node", More: "smith kinds lists the kinds"})

	if err := check(add, []string{"app", "web"}); err != nil {
		t.Fatalf("right count rejected: %v", err)
	}
	err := check(add, []string{"app"})
	if err == nil || !strings.Contains(err.Error(), "smith add needs a name") || !strings.Contains(err.Error(), "smith add app web --lang node") || !strings.Contains(err.Error(), "smith kinds") {
		t.Fatalf("one missing: %v", err)
	}
	err = check(add, nil)
	if err == nil || !strings.Contains(err.Error(), "needs a kind (app, lib or package) and a name") {
		t.Fatalf("none given: %v", err)
	}
	err = check(add, []string{"app", "web", "extra"})
	if err == nil || !strings.Contains(err.Error(), `"extra"`) {
		t.Fatalf("too many: %v", err)
	}
}
