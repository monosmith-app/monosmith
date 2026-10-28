package agentfiles

import (
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/mcpserver"
)

// CatalogTools are the tools smith mcp serves, for the agent text.
func CatalogTools() []Tool {
	tools := make([]Tool, 0, len(mcpserver.Tools))
	for _, t := range mcpserver.Tools {
		tools = append(tools, Tool{Name: t.Name, Summary: firstSentence(t.Description), ReadOnly: t.ReadOnly})
	}
	return tools
}

func firstSentence(s string) string {
	if i := strings.Index(s, ". "); i >= 0 {
		return s[:i]
	}
	return strings.TrimSuffix(s, ".")
}
