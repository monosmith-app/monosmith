package depsview

import (
	"strings"
	"testing"
)

func TestTheMajorWarningAgreesWithTheCount(t *testing.T) {
	if got := majorWarning(0); got != "" {
		t.Errorf("no major release, no warning: %q", got)
	}
	if got := majorWarning(1); !strings.Contains(got, "The 1 major one can break") || strings.Contains(got, "ones") {
		t.Errorf("one: %q", got)
	}
	if got := majorWarning(3); !strings.Contains(got, "The 3 major ones can break") {
		t.Errorf("many: %q", got)
	}
}
