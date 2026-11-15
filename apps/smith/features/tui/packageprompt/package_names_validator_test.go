package packageprompt

import (
	"strings"
	"testing"
)

func TestPackagesAreSplitAtSpacesAndNeverReachTheManagerAsFlags(t *testing.T) {
	got, err := ParsePackages("  zod  @types/node   react@19 ")
	if err != nil || strings.Join(got, ",") != "zod,@types/node,react@19" {
		t.Fatalf("got %v, %v", got, err)
	}
	for _, bad := range []string{"", "   ", "zod -D", "--registry=http://evil", "ok\x1bbad"} {
		if _, err := ParsePackages(bad); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
	if _, err := ParsePackages("zod -D"); err == nil || !strings.Contains(err.Error(), "development dependency") {
		t.Fatalf("a flag is told where the option lives: %v", err)
	}
}
