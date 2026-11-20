package release

import (
	"bytes"
	"testing"
)

func TestProgressGoesToStderrWithJSON(t *testing.T) {
	var out, errOut bytes.Buffer
	if progressWriter(true, &out, &errOut) != &errOut {
		t.Error("with --json, progress goes to stderr so stdout holds only the report")
	}
	if progressWriter(false, &out, &errOut) != &out {
		t.Error("without --json, progress goes to stdout")
	}
}
