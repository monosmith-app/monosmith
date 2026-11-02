package scaffold

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/operationplan"
)

func TestChooseUnsaved(t *testing.T) {
	preview := Removal{Project: "web", Root: "apps/web", Unsaved: []string{"apps/web/notes.txt"}, Ignored: 2}
	cases := []struct {
		name    string
		canAsk  bool
		input   string
		want    string
		wantErr string
	}{
		{"no terminal: refused, flags named", false, "", "", "--keep-copy"},
		{"Enter keeps a copy", true, "\n", removeKeepCopy, ""},
		{"delete for good", true, "2\n", removeDelete, ""},
		{"cancel", true, "3\n", "", "cancelled"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := chooseUnsaved(preview, c.canAsk, bufio.NewReader(strings.NewReader(c.input)), &out)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("got %v, want an error with %q", err, c.wantErr)
				}
				if c.canAsk && !errors.Is(err, operationplan.ErrDeclined) {
					t.Fatalf("a cancel must be ErrDeclined: %v", err)
				}
				return
			}
			if err != nil || got != c.want {
				t.Fatalf("got %q, %v; want %q\n%s", got, err, c.want, out.String())
			}
			if !strings.Contains(out.String(), "apps/web/notes.txt") {
				t.Fatalf("the files are not listed:\n%s", out.String())
			}
		})
	}
}

func TestPlanRemovalSaysWhatHappensToTheFolder(t *testing.T) {
	cases := []struct {
		name string
		r    Removal
		want string
	}{
		{"clean", Removal{Project: "web", Root: "apps/web"}, "the whole directory"},
		{"unsaved files", Removal{Project: "web", Root: "apps/web", Unsaved: []string{"a"}}, "git cannot bring back"},
		{"kept", Removal{Project: "web", Root: "apps/web", Unsaved: []string{"a"}, KeptAt: ".smith/removed/web-1"}, "nothing is lost"},
	}
	for _, c := range cases {
		note := PlanRemoval(t.TempDir(), c.r).Groups[0].Lines[0].Note
		if !strings.Contains(note, c.want) {
			t.Errorf("%s: note %q lacks %q", c.name, note, c.want)
		}
	}
}
