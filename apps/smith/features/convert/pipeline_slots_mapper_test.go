package convert

import (
	"strings"
	"testing"
)

func TestRenameSlots(t *testing.T) {
	old := `steps:
      # mnci:slot after-install  (your own steps go between these two lines; mnci upgrade keeps them)
      - run: echo team step
      # mnci:slot-end after-install
      # mnci:slot before-release  (your own steps go between these two lines)
      # mnci:slot-end before-release
      # mnci:slot after-release
      - run: notify
      # mnci:slot-end after-release
`
	got, orphans := RenameSlots(old)
	if !strings.Contains(got, "      # smith:slot before-setup\n      - run: echo team step\n      # smith:slot-end before-setup\n") {
		t.Fatalf("after-install must become before-setup:\n%s", got)
	}
	if !strings.Contains(got, "# smith:slot after-verify\n      # smith:slot-end after-verify") {
		t.Fatalf("before-release must become after-verify:\n%s", got)
	}
	if len(orphans) != 1 || orphans[0] != "after-release" {
		t.Fatalf("orphans %v", orphans)
	}
}
