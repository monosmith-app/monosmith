package toolrun

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestToolchainReadsTheFirstVersionNumber(t *testing.T) {
	// The test binary itself is a tool that answers a version on request.
	t.Setenv("TOOLCHAIN_TEST_ANSWER", "tool version go1.23.4 linux/amd64")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := exec.LookPath(exe); err != nil {
		t.Skip("the test binary is not runnable by path")
	}
	got := Toolchain("demo", exe, "-test.run=TestHelperPrintsAVersion", "--")
	if got != "demo 1.23.4" {
		t.Errorf("%q", got)
	}
}

func TestHelperPrintsAVersion(t *testing.T) {
	if answer := os.Getenv("TOOLCHAIN_TEST_ANSWER"); answer != "" && len(os.Args) > 1 && strings.Contains(strings.Join(os.Args, " "), "TestHelperPrintsAVersion") {
		_, _ = os.Stdout.WriteString(answer + "\n")
	}
}

func TestAMissingToolHasNoToolchain(t *testing.T) {
	if got := Toolchain("nope", "smith-no-such-tool-anywhere", "--version"); got != "" {
		t.Errorf("%q", got)
	}
}

func TestFirstLineSkipsBlankLines(t *testing.T) {
	if got := firstLine("\n \nopenjdk version \"21.0.2\"\nsecond"); got != "openjdk version \"21.0.2\"" {
		t.Errorf("%q", got)
	}
}
