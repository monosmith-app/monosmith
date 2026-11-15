package workspace

import (
	"strings"
	"testing"
)

func TestPropagationWaitMustBeADuration(t *testing.T) {
	ok := &Config{Release: ReleaseSection{PropagationWait: "20m"}, Projects: map[string]ProjectOverride{"cli": {Release: &ProjectRelease{PropagationWait: "0"}}}}
	if err := validatePropagationWait(ok); err != nil {
		t.Fatal(err)
	}
	bad := &Config{Projects: map[string]ProjectOverride{"cli": {Release: &ProjectRelease{PropagationWait: "soon"}}}}
	if err := validatePropagationWait(bad); err == nil || !strings.Contains(err.Error(), "[projects.cli.release]") {
		t.Errorf("the project's table must be named: %v", err)
	}
	if err := validatePropagationWait(&Config{Release: ReleaseSection{PropagationWait: "-1m"}}); err == nil || !strings.Contains(err.Error(), "[release]") {
		t.Errorf("a negative wait must be refused: %v", err)
	}
}
