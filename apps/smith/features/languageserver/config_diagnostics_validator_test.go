package languageserver

import (
	"strings"
	"testing"
)

func TestDiagnoseValidConfigIsClean(t *testing.T) {
	if d := Diagnose("[workspace]\nname = \"acme\"\n"); len(d) != 0 {
		t.Fatalf("want none, got %+v", d)
	}
}

func TestDiagnoseUnknownKeyIsUnderlinedWhereItIs(t *testing.T) {
	d := Diagnose("[workspace]\nname = \"acme\"\nscpoe = \"@acme\"\n")
	if len(d) != 1 {
		t.Fatalf("want 1, got %+v", d)
	}
	if !strings.Contains(d[0].Message, "workspace.scpoe") {
		t.Errorf("message %q", d[0].Message)
	}
	r := d[0].Range
	if r.Start.Line != 2 || r.Start.Character != 0 || r.End.Character != 5 {
		t.Errorf("range %+v, want line 2, columns 0-5", r)
	}
	if d[0].Severity != SeverityError || d[0].Source != "smith" {
		t.Errorf("severity or source: %+v", d[0])
	}
}

func TestDiagnoseUnknownTable(t *testing.T) {
	d := Diagnose("[workspace]\nname = \"a\"\n\n[nonsense]\nx = 1\n")
	if len(d) == 0 || !strings.Contains(d[0].Message, "nonsense") {
		t.Fatalf("got %+v", d)
	}
	if d[0].Range.Start.Line != 3 {
		t.Errorf("line %d, want 3", d[0].Range.Start.Line)
	}
}

func TestDiagnoseSyntaxError(t *testing.T) {
	d := Diagnose("[workspace\nname = \"a\"\n")
	if len(d) != 1 || d[0].Message == "" {
		t.Fatalf("got %+v", d)
	}
}

func TestDiagnoseWrongType(t *testing.T) {
	d := Diagnose("[workspace]\nname = 5\n")
	if len(d) != 1 || d[0].Range.Start.Line != 1 {
		t.Fatalf("got %+v", d)
	}
}

func TestDiagnoseRejectedValueMovesToFirstLine(t *testing.T) {
	d := Diagnose("[workspace]\nname = \"a\"\n[ci]\nproviders = [\"nope\"]\n")
	if len(d) != 1 || d[0].Range.Start.Line != 0 || !strings.Contains(d[0].Message, "nope") {
		t.Fatalf("got %+v", d)
	}
	if strings.HasPrefix(d[0].Message, "smith.toml:") {
		t.Errorf("file prefix kept: %q", d[0].Message)
	}
}
