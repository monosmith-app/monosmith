package releaseassets

import (
	"bytes"
	"strings"
	"testing"
)

func TestChooseSource(t *testing.T) {
	for name, tc := range map[string]struct {
		source, answer string
		terminal       bool
		want           string
		wantErr        []string
	}{
		"flag current":              {source: "current", want: SourceCurrent},
		"flag tag":                  {source: "tag", want: SourceTag},
		"unknown flag":              {source: "main", wantErr: []string{`"main"`, "--source current", "--source tag"}},
		"no terminal, no flag":      {wantErr: []string{"--source current", "--source tag", "1.2.0", "exact code"}},
		"terminal answers 1":        {terminal: true, answer: "1\n", want: SourceCurrent},
		"terminal answers tag":      {terminal: true, answer: "tag\n", want: SourceTag},
		"terminal answers 2":        {terminal: true, answer: "2\n", want: SourceTag},
		"terminal answers nonsense": {terminal: true, answer: "maybe\n", wantErr: []string{"maybe", "--source"}},
		"terminal gives no answer":  {terminal: true, answer: "", wantErr: []string{"--source current"}},
	} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := chooseSource(tc.source, "app@1.2.0", "1.2.0", tc.terminal, strings.NewReader(tc.answer), &out)
			if len(tc.wantErr) > 0 {
				if err == nil {
					t.Fatalf("want an error, got %q", got)
				}
				for _, w := range tc.wantErr {
					if !strings.Contains(err.Error(), w) {
						t.Errorf("the error misses %q: %v", w, err)
					}
				}
				return
			}
			if err != nil || got != tc.want {
				t.Fatalf("got %q, %v; want %q", got, err, tc.want)
			}
			if tc.terminal && !strings.Contains(out.String(), "newer code under an old number") {
				t.Errorf("the question must say what each option means:\n%s", out.String())
			}
		})
	}
}

func TestCheckSource(t *testing.T) {
	for _, ok := range []string{"", "current", "tag"} {
		if err := checkSource(ok); err != nil {
			t.Errorf("%q: %v", ok, err)
		}
	}
	if err := checkSource("choose"); err == nil {
		t.Error("choose is turned into no source by Run, not accepted by checkSource")
	}
}
