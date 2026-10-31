package tauriadd

import (
	"fmt"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/configedit"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// renderStack is the smith.toml text that declares a stack.
func renderStack(st protocol.Stack, project string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Written by smith tauri add for the project %s.\n", project)
	fmt.Fprintf(&b, "[stacks.%s]\n", configedit.QuoteKey(st.Name))
	fmt.Fprintf(&b, "description = %s\n", configedit.Quote(st.Description))
	fmt.Fprintf(&b, "language = %s\n", configedit.Quote(st.Language))
	list(&b, "toolchain", st.Toolchain)
	list(&b, "structure", st.Structure)
	if st.Scaffold != "" {
		fmt.Fprintf(&b, "scaffold = %s\n", configedit.Quote(st.Scaffold))
	}
	list(&b, "deploy", st.Deploy)
	return b.String()
}

func list(b *strings.Builder, name string, items []string) {
	if len(items) == 0 {
		return
	}
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = configedit.Quote(item)
	}
	fmt.Fprintf(b, "%s = [%s]\n", name, strings.Join(quoted, ", "))
}
