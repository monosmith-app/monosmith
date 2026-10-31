package lineendings

import "testing"

func TestConvert(t *testing.T) {
	cases := []struct {
		name, in string
		crlf     bool
		want     string
	}{
		{"lf to crlf", "a\nb\n", true, "a\r\nb\r\n"},
		{"crlf to lf", "a\r\nb\r\n", false, "a\nb\n"},
		{"mixed to crlf", "a\r\nb\nc\n", true, "a\r\nb\r\nc\r\n"},
		{"already crlf", "a\r\nb\r\n", true, "a\r\nb\r\n"},
		{"no newline", "abc", true, "abc"},
	}
	for _, c := range cases {
		if got := Convert(c.in, c.crlf); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}

func TestIsCRLF(t *testing.T) {
	cases := []struct {
		name, in    string
		crlf, known bool
	}{
		{"crlf", "a\r\nb\r\n", true, true},
		{"lf", "a\nb\n", false, true},
		{"mostly crlf", "a\r\nb\r\nc\n", true, true},
		{"mostly lf", "a\nb\nc\r\n", false, true},
		{"no line ending", "abc", false, false},
		{"empty", "", false, false},
	}
	for _, c := range cases {
		crlf, known := IsCRLF(c.in)
		if crlf != c.crlf || known != c.known {
			t.Errorf("%s: got (%v, %v), want (%v, %v)", c.name, crlf, known, c.crlf, c.known)
		}
	}
}

func TestMatchFollowsTheExistingText(t *testing.T) {
	if got := Match("x\r\ny\r\n", "a\nb\n"); got != "a\r\nb\r\n" {
		t.Errorf("CRLF existing: got %q", got)
	}
	if got := Match("x\ny\n", "a\r\nb\r\n"); got != "a\nb\n" {
		t.Errorf("LF existing: got %q", got)
	}
	if got := Match("one line", "a\nb\n"); got != "a\nb\n" {
		t.Errorf("no ending to follow: got %q", got)
	}
}
