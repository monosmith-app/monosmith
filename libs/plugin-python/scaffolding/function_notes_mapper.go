package scaffolding

import "github.com/monosmith-app/monosmith/packages/plugin-sdk/functionhost"

// functionNotes are the next steps printed after a function app is made: how to
// run it, what the local tool is, and for Azure the deploy pipeline smith has.
func functionNotes(host string) []string {
	tool, _ := functionhost.ToolOf(host)
	notes := []string{"Run it locally with `smith run start`. `smith test` needs no cloud tool."}
	if tool.Binary != "" {
		notes = append(notes, "The local run needs "+tool.Name+" ("+tool.Binary+"); `smith doctor` checks it. To get it: "+tool.Install+".")
	}
	if host == functionhost.Azure {
		notes = append(notes, "Copy local.settings.sample.json to local.settings.json (git ignores it) for your settings.")
	}
	return notes
}
