package convert

import (
	"regexp"
	"strings"
)

// slotRenames maps the old mnci's pipeline slots to the closest slot of the
// pipelines smith generates. A slot with no equivalent is reported.
var slotRenames = map[string]string{
	"after-install":  "before-setup",
	"before-release": "after-verify",
}

var oldSlot = regexp.MustCompile(`(?ms)^([ \t]*)# mnci:slot ([\w-]+)[^\n]*\n(.*?)^[ \t]*# mnci:slot-end ([\w-]+)[^\n]*$`)

// RenameSlots rewrites the old pipeline's slot markers to the new names, so
// regenerating the pipeline keeps the team's steps. It returns the slots
// that hold steps but have no equivalent.
func RenameSlots(old string) (string, []string) {
	var orphans []string
	out := oldSlot.ReplaceAllStringFunc(old, func(block string) string {
		m := oldSlot.FindStringSubmatch(block)
		indent, name, body := m[1], m[2], m[3]
		to, ok := slotRenames[name]
		if !ok {
			if strings.TrimSpace(body) != "" {
				orphans = append(orphans, name)
			}
			to = name
		}
		return indent + "# smith:slot " + to + "\n" + body + indent + "# smith:slot-end " + to
	})
	return out, orphans
}
