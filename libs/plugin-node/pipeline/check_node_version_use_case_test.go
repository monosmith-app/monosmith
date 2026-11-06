package pipeline

import (
	"errors"
	"strings"
	"testing"
)

func TestCheckNodeVersion(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	for _, c := range []struct {
		out       string
		err       error
		ok, warns bool
	}{
		{"v24.11.0\n", nil, true, false},
		{"v25.1.0\n", nil, false, true}, // newer than CI warns
		{"v22.3.0\n", nil, false, true}, // older than CI warns too
		{"", errors.New("not found"), false, false},
	} {
		probe = func(string, []string) (string, error) { return c.out, c.err }
		f := CheckNodeVersion(".", "24")
		if f.OK != c.ok || f.Warning != c.warns {
			t.Errorf("%q: %+v", c.out, f)
		}
	}
}

func TestAnOtherNodeMajorOffersKeepingItOrGettingTheCIOneAndNeverRunsAnythingItself(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "v22.3.0\n", nil }
	f := CheckNodeVersion(".", "24")
	if f.OK || !f.Warning || f.Fix == nil || f.Fix.Safe || len(f.Fix.Steps) != 0 || len(f.Fix.Files) != 0 {
		t.Fatalf("a difference is a warning with a question and no step of its own: %+v", f)
	}
	options := f.Fix.Choices[0].Options
	if len(options) != 2 || len(options[0].Fix.Steps) != 0 || !strings.HasPrefix(options[0].Label, "Keep Node.js 22.3.0") {
		t.Fatalf("keeping it is the first answer and runs nothing: %+v", options)
	}
	if got := strings.Join(options[1].Fix.Steps[0].Argv, " "); got != "smith tools install node" {
		t.Fatalf("the install goes through the tools flow: %q", got)
	}
	if !strings.Contains(f.Detail, "older than the 24") || !strings.Contains(f.Detail, "warning") {
		t.Fatalf("the finding says what differs and why it may matter: %q", f.Detail)
	}
}

func TestANodeThatIsMissingStaysAFailureWithoutAFix(t *testing.T) {
	real := probe
	defer func() { probe = real }()
	probe = func(string, []string) (string, error) { return "", errors.New("not found") }
	if f := CheckNodeVersion(".", "24"); f.OK || f.Warning || f.Fix != nil {
		t.Fatalf("%+v", f)
	}
}
