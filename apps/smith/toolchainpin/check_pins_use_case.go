package toolchainpin

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// ToolVersionsFile is the file mise and asdf read the pinned versions from.
const ToolVersionsFile = ".tool-versions"

// CheckPins compares each pin ([toolchains], by toolchain label: node, go,
// python, rustc, java, ...) with what the projects' plugins probed on this
// machine. A pin no project reports on (the tool is missing, or no project
// uses it) gives no finding. A mismatch fails, and its fix writes the pins to
// .tool-versions (existing is that file's current text) for mise or asdf to
// install; the fix is safe when the file is new.
func CheckPins(pins map[string]string, projects []protocol.Project, existing string) []protocol.Finding {
	installed := map[string]map[string]bool{}
	for _, p := range projects {
		if label, v := Installed(p.Toolchain); label != "" {
			if installed[label] == nil {
				installed[label] = map[string]bool{}
			}
			installed[label][v] = true
		}
	}
	labels := make([]string, 0, len(pins))
	for l := range pins {
		labels = append(labels, l)
	}
	sort.Strings(labels)
	var out []protocol.Finding
	for _, label := range labels {
		versions := installed[label]
		if len(versions) == 0 {
			continue
		}
		pin := pins[label]
		var wrong []string
		for v := range versions {
			if !Satisfies(pin, v) {
				wrong = append(wrong, v)
			}
		}
		sort.Strings(wrong)
		f := protocol.Finding{Check: fmt.Sprintf("%s matches the pinned version %s", label, pin), OK: len(wrong) == 0}
		if !f.OK {
			f.Detail = fmt.Sprintf("[toolchains] pins %s %s, but this machine runs %s: tasks cached on another version are misses, and results can differ from CI", label, pin, strings.Join(wrong, ", "))
			f.Remedy = fmt.Sprintf("install %s %s (`mise use %s@%s`, asdf, or your installer; `smith doctor --fix` writes %s for mise and asdf), or change [toolchains] %s", label, pin, ToolName(label), pin, ToolVersionsFile, label)
			f.Fix = &protocol.Fix{
				Description: "pin the runtimes in " + ToolVersionsFile,
				Safe:        existing == "",
				Files:       []protocol.FileEdit{{Path: ToolVersionsFile, Content: MergeToolVersions(existing, pins)}},
			}
		}
		out = append(out, f)
	}
	return out
}
