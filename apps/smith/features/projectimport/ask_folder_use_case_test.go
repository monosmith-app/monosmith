package projectimport

import (
	"io"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

func TestAskFolderTakesTheSuggestedKindOnAnEmptyAnswer(t *testing.T) {
	dir, kind, err := AskFolder("api", protocol.ProjectApp, strings.NewReader("\n"), io.Discard)
	if err != nil || dir != "apps/api" || kind != protocol.ProjectApp {
		t.Fatalf("%q %q %v", dir, kind, err)
	}
}

func TestAskFolderTakesAnotherFolderAndItsKind(t *testing.T) {
	dir, kind, err := AskFolder("api", protocol.ProjectApp, strings.NewReader("4\nlibs/shared-api\n"), io.Discard)
	if err != nil || dir != "libs/shared-api" || kind != protocol.ProjectLib {
		t.Fatalf("%q %q %v", dir, kind, err)
	}
	dir, kind, err = AskFolder("api", protocol.ProjectApp, strings.NewReader("4\ntools/api\n"), io.Discard)
	if err != nil || dir != "tools/api" || kind != protocol.ProjectApp {
		t.Fatalf("outside the kind folders the suggested kind holds: %q %q %v", dir, kind, err)
	}
}

func TestAskFolderGivesUpWithoutAnAnswer(t *testing.T) {
	if _, _, err := AskFolder("api", protocol.ProjectLib, strings.NewReader("9\n7\nx\n"), io.Discard); err == nil {
		t.Error("three bad answers must stop the question")
	}
	if _, _, err := AskFolder("api", protocol.ProjectLib, strings.NewReader(""), io.Discard); err == nil {
		t.Error("no answer must stop the question")
	}
}
