package agentfiles

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
)

const goldenAgent = "testdata/smith-agent.golden.md"

// The agent file as generated is pinned. A tool added, renamed or reworded in
// the MCP catalog changes it; run with UPDATE_GOLDEN=1 to accept the new text
// after reading the diff.
func TestAgentFileMatchesTheGolden(t *testing.T) {
	got := RenderAgent(CatalogTools())
	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(goldenAgent, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(goldenAgent)
	if err != nil {
		t.Fatal(err)
	}
	if strings.ReplaceAll(string(want), "\r\n", "\n") != got {
		t.Fatalf("%s is out of date with the MCP tool catalog (apps/smith/mcpserver); run UPDATE_GOLDEN=1 go test ./agentfiles and read the diff", goldenAgent)
	}
}

func TestEveryToolIsListedExactlyOnceAndNoOtherNameIsInvented(t *testing.T) {
	text := RenderAgent(CatalogTools())
	real := map[string]bool{}
	for _, tool := range mcpserver.Tools {
		real[tool.Name] = true
		if n := strings.Count(text, "- "+tool.Name+": "); n != 1 {
			t.Errorf("%s is listed %d times", tool.Name, n)
		}
	}
	for _, name := range regexp.MustCompile(`smith_[a-z_]+`).FindAllString(text, -1) {
		if !real[name] {
			t.Errorf("the agent text names %s, which is not an MCP tool", name)
		}
	}
}

func TestToolsAreListedUnderTheirSafety(t *testing.T) {
	text := RenderAgent(CatalogTools())
	split := strings.Index(text, "Tools that change things")
	for _, tool := range mcpserver.Tools {
		at := strings.Index(text, "- "+tool.Name+": ")
		if tool.ReadOnly != (at < split) {
			t.Errorf("%s (read-only %v) is under the wrong heading", tool.Name, tool.ReadOnly)
		}
	}
}

func TestTheMarkerHashCoversTheText(t *testing.T) {
	text := RenderAgent(CatalogTools())
	plain, recorded, ok := unmarked(text)
	if !ok || digest(plain) != recorded {
		t.Fatalf("marker %q does not match the text (%v)", recorded, ok)
	}
	if !strings.HasPrefix(text, "---\nname: smith\n") {
		t.Error("Claude Code reads the frontmatter from the first line")
	}
	if _, _, ok := unmarked(strings.ReplaceAll(text, "\n", "\r\n")); !ok {
		t.Error("a checkout with CRLF line endings loses the marker")
	}
}
