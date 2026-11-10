package adoptmove

import (
	"bytes"
	"strings"
	"testing"
)

func TestAskWhere(t *testing.T) {
	app := Candidate{Manifest: "package.json", Name: "web", App: true}
	lib := Candidate{Manifest: "go.mod", Name: "kit"}
	cases := []struct {
		name  string
		c     Candidate
		input string
		want  string
		fails bool
	}{
		{"enter takes the suggested app folder", app, "\n\n", "apps/web", false},
		{"enter takes the suggested package folder", lib, "\n\n", "packages/kit", false},
		{"libs with an edited name", app, "3\nweb-ui\n", "libs/web-ui", false},
		{"another folder", app, "4\ntools/web\n", "tools/web", false},
		{"leave it", app, "5\n", Leave, false},
		{"a bad number asks again", app, "9\nx\n2\n\n", "packages/web", false},
		{"a bad name asks the question again", app, "1\nBad Name\n1\nok\n", "apps/ok", false},
		{"three bad answers give up", app, "x\ny\nz\n", "", true},
		{"end of input is no answer", app, "", "", true},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		got, err := AskWhere(tc.c, strings.NewReader(tc.input), &out)
		if (err != nil) != tc.fails || got != tc.want {
			t.Errorf("%s: got %q, %v; want %q (fails=%v)\n%s", tc.name, got, err, tc.want, tc.fails, out.String())
		}
	}
}

func TestAskWhereMarksTheSuggestedChoice(t *testing.T) {
	var out bytes.Buffer
	_, _ = AskWhere(Candidate{Manifest: "package.json", Name: "web", App: true}, strings.NewReader("5\n"), &out)
	for _, want := range []string{"1) apps/web (suggested)", "2) packages/web\n", "3) libs/web\n", "4) another folder", "5) no, leave it", Question} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output lacks %q:\n%s", want, out.String())
		}
	}
}
