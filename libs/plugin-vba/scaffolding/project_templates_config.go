package scaffolding

import "strings"

// source is the text of a new project's main module, with the CRLF line
// endings the VBA editor exports. The header comments sit under the Attribute
// line, which must come first. An app has a Main procedure, a lib a Greet
// function.
func source(module, name, kind, stack string) string {
	lines := []string{`Attribute VB_Name = "` + module + `"`, "' smithKind: " + kind}
	if stack != "" {
		lines = append(lines, "' smithStack: "+stack)
	}
	lines = append(lines, "' smithVersion: 0.0.0", "Option Explicit", "")
	if kind == "app" {
		lines = append(lines, "Public Sub Main()", `    Debug.Print "Hello from `+name+`"`, "End Sub")
	} else {
		lines = append(lines, "Public Function Greet(ByVal who As String) As String", `    Greet = "Hello, " & who & "!"`, "End Function")
	}
	return strings.Join(lines, "\r\n") + "\r\n"
}
