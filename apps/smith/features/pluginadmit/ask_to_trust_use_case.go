package pluginadmit

import (
	"fmt"
	"strings"
)

// Review prints what a plugin is and what it asks for, so the person can
// decide: its name, where it comes from, the command, the files it runs with
// their digests, and the environment it may see.
func Review(p *Prompter, st Status) {
	fmt.Fprintf(p.Out, "\nThe external plugin %q wants to run on this machine.\n", st.Name)
	fmt.Fprintf(p.Out, "  from      %s\n", sourceLine(st.Plugin))
	fmt.Fprintf(p.Out, "  command   %s\n", strings.Join(st.Command, " "))
	for i, e := range st.Fingerprint.Executables {
		label := "          "
		if i == 0 {
			label = "  runs    "
		}
		fmt.Fprintf(p.Out, "%s  %s\n            sha256 %s\n", label, e.Path, e.SHA256)
	}
	fmt.Fprintf(p.Out, "  sees      %s\n", environmentLine(st))
	fmt.Fprintln(p.Out, "A trusted plugin runs with your privileges and is not sandboxed. Read its code before you trust it.")
}

// Ask shows the plugin and asks whether to trust it; only "y" or "yes" is a
// yes.
func Ask(p *Prompter, st Status) bool {
	Review(p, st)
	fmt.Fprintf(p.Out, "Trust %q? [y/N] ", st.Name)
	answer, _ := p.In.ReadString('\n')
	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes"
}

func sourceLine(p Plugin) string {
	if p.Where == "" || p.Where == p.Source {
		return p.Source
	}
	return p.Source + " (" + p.Where + ")"
}

func environmentLine(st Status) string {
	if len(st.Env) == 0 {
		return "the baseline environment only (PATH, HOME, TEMP, locale and proxy variables)"
	}
	return "the baseline environment plus " + strings.Join(st.Env, ", ")
}
