package scaffold

import (
	"reflect"
	"testing"
)

func TestParseUnsavedSeparatesIgnoredFromWork(t *testing.T) {
	got := parseUnsaved("?? apps/web/notes.txt\n M apps/web/src/a.ts\nA  apps/web/new.ts\n!! apps/web/node_modules/\n")
	want := UnsavedFiles{Files: []string{"apps/web/notes.txt", "apps/web/src/a.ts", "apps/web/new.ts"}, Ignored: 1}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	if got := parseUnsaved(""); len(got.Files) != 0 || got.Ignored != 0 {
		t.Fatalf("%+v", got)
	}
}
