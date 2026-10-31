package lineendings

import "testing"

func TestEOLFromAttributes(t *testing.T) {
	attrs := "# comment\n* text=auto eol=lf\n*.bat eol=crlf\n*.png binary\napps/win/** eol=crlf\n**/legacy/*.txt eol=crlf\n"
	cases := []struct{ rel, want string }{
		{"package.json", "lf"},
		{"scripts/run.bat", "crlf"},
		{"logo.png", ""},
		{"apps/win/readme.md", "crlf"},
		{"x/legacy/a.txt", "crlf"},
		{"x/other/a.txt", "lf"},
	}
	for _, c := range cases {
		if got := EOLFromAttributes(attrs, c.rel); got != c.want {
			t.Errorf("%s: got %q, want %q", c.rel, got, c.want)
		}
	}
	if got := EOLFromAttributes("* text=auto\n", "a.txt"); got != "" {
		t.Errorf("no eol rule: got %q", got)
	}
}
