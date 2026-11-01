package scaffold

import (
	"fmt"
	"io"
)

// maxMentionsShown bounds the list printed under "Still mentions".
const maxMentionsShown = 20

// WriteMentions prints the "Still mentions <name>" section: at most 20 files,
// grouped by kind, then "and N more". It prints nothing when there are none.
//
// Parameters: out is the writer; r is the Removal that carries the mentions.
func WriteMentions(out io.Writer, r Removal) {
	if len(r.Mentions) == 0 {
		return
	}
	fmt.Fprintf(out, "\nStill mentions %s\n", r.Project)
	if r.MentionsRootOnly {
		fmt.Fprintf(out, "  (the name is short, so only the path %s was searched)\n", r.Root)
	}
	fmt.Fprintln(out, "  smith does not change these files. Review them.")
	titles := map[string]string{MentionWorkflow: "workflow files", MentionEditor: "editor files", MentionOther: "other files"}
	last := ""
	for i, m := range r.Mentions {
		if i == maxMentionsShown {
			fmt.Fprintf(out, "  and %d more\n", len(r.Mentions)-maxMentionsShown)
			return
		}
		if m.Kind != last {
			fmt.Fprintf(out, "  %s\n", titles[m.Kind])
			last = m.Kind
		}
		fmt.Fprintf(out, "    %s:%d\n", m.File, m.Line)
	}
}
