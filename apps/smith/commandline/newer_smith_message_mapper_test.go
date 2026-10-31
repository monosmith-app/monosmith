package commandline

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

const newerConfig = "[workspace]\nname = \"a\"\n[ci]\nazure_pool = \"x\"\n[[future_groups]]\nname = \"g\"\n[future_table]\nk = 1\n"

func TestNewerSmithMessageIsOneSentenceAndOneCommand(t *testing.T) {
	_, err := workspace.Parse([]byte(newerConfig))
	got, ok := newerSmithMessage(fmt.Errorf("opening: %w", err), "1.2.3", false)
	if !ok {
		t.Fatalf("a smith.toml with unknown tables needs a newer smith, got %v", err)
	}
	want := "This smith (1.2.3) is older than smith.toml needs. Run `npm install --global monosmith`."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestNewerSmithMessageKeepsTheDetailsBehindVerbose(t *testing.T) {
	_, err := workspace.Parse([]byte(newerConfig))
	got, _ := newerSmithMessage(err, "1.2.3", true)
	if !strings.Contains(got, "future_groups") || !strings.Contains(got, "missing table") {
		t.Errorf("--verbose lists the parse errors:\n%s", got)
	}
}

func TestNewerSmithMessageIgnoresATypoAndOtherErrors(t *testing.T) {
	_, typo := workspace.Parse([]byte("[workspace]\nname = \"a\"\nnmae = \"typo\"\n"))
	for _, err := range []error{typo, errors.New("boom")} {
		if got, ok := newerSmithMessage(err, "1.2.3", false); ok {
			t.Errorf("%v is not a newer-smith error, got %q", err, got)
		}
	}
}

func TestNewerSmithMessageForANewerSchema(t *testing.T) {
	_, err := workspace.Parse([]byte("schema = 99\n[workspace]\nname = \"a\"\n"))
	if _, ok := newerSmithMessage(err, "1.2.3", false); !ok {
		t.Errorf("a newer schema needs a newer smith, got %v", err)
	}
}

func TestWantsVerbose(t *testing.T) {
	if !wantsVerbose([]string{"doctor", "--verbose"}) || wantsVerbose([]string{"doctor"}) {
		t.Error("--verbose is read from the command line")
	}
}
