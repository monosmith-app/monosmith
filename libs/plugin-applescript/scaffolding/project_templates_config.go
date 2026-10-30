package scaffolding

// appTemplate is a script that runs: osascript <name>.scpt calls its run
// handler.
const appTemplate = `-- smithKind: [[.Kind]]
[[- if .Stack]]
-- smithStack: [[.Stack]]
[[- end]]
-- smithVersion: 0.0.0

on run
	return "Hello from [[.Name]]"
end run
`

// libTemplate is a script others load: it offers a handler.
const libTemplate = `-- smithKind: [[.Kind]]
[[- if .Stack]]
-- smithStack: [[.Stack]]
[[- end]]
-- smithVersion: 0.0.0

on greet(who)
	return "Hello, " & who & "!"
end greet
`
