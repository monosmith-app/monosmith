package adoption

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

const engineWarning = "npm warn EBADENGINE Unsupported engine {\nnpm warn EBADENGINE   package: 'a@1.0.0',\nnpm warn EBADENGINE   required: { node: '>=22' }\nnpm warn EBADENGINE }\n"

func TestAnInstallThatWorkedFoldsNpmWarningsIntoOneLine(t *testing.T) {
	var out bytes.Buffer
	reportInstall(&out, engineWarning, nil)
	text := out.String()
	if strings.Contains(text, "EBADENGINE") || !strings.Contains(text, "1 package wants a newer Node") || !strings.HasPrefix(text, "  npm install done\n") {
		t.Fatalf("%s", text)
	}
}

func TestAnInstallThatFailedShowsNpmsWholeOutput(t *testing.T) {
	var out bytes.Buffer
	reportInstall(&out, engineWarning, errors.New("exit 1"))
	if text := out.String(); !strings.Contains(text, "EBADENGINE") || !strings.Contains(text, "run it yourself") {
		t.Fatalf("%s", text)
	}
}
