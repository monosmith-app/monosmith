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

// moduleName is a VBA identifier made from a project name: letters, digits
// and underscores, starting with a letter.
func moduleName(name string) string {
	var b strings.Builder
	for _, c := range name {
		if c == '_' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			b.WriteRune(c)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 || !strings.ContainsAny(b.String()[:1], "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		return "M" + b.String()
	}
	return b.String()
}
