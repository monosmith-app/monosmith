package termcolour

import "regexp"

// paintRules are the result lines smith paint recognises, first match wins.
// They cover go test, which has no colour of its own, and the
// "error CODE:" and "warning CODE:" lines of tsc and MSBuild (dotnet build).
var paintRules = []struct {
	line *regexp.Regexp
	code string
}{
	{regexp.MustCompile(`^\s*--- FAIL`), Red},
	{regexp.MustCompile(`^FAIL(\s|$)`), Red},
	{regexp.MustCompile(`^(panic|fatal error): `), Red},
	{regexp.MustCompile(`^\s*(error|Error|ERROR)(\[[\w-]+\])?: `), Red},
	{regexp.MustCompile(`(^|\s)error [A-Z]+\d+: `), Red},
	{regexp.MustCompile(`^\s*(warning|Warning|WARNING|WARN)(\[[\w-]+\])?: `), Yellow},
	{regexp.MustCompile(`(^|\s)warning [A-Z]+\d+: `), Yellow},
	{regexp.MustCompile(`^\s*--- SKIP`), Dim},
	{regexp.MustCompile(`^\?\s+\S+\s+\[no test files\]`), Dim},
	{regexp.MustCompile(`^\s*--- PASS`), Green},
	{regexp.MustCompile(`^ok\s+\S`), Green},
	{regexp.MustCompile(`^PASS$`), Green},
}

// Paint is smith paint: it colours a result line it recognises (pass green,
// fail and error red, warning yellow, skipped dim) and leaves every other
// line, and every line that is coloured already, as it is. It only wraps the
// line in SGR codes; the characters stay the same.
//
// Parameters: line is one line of tool output, without its newline.
// Returns: the line, painted or unchanged.
func Paint(line string) string {
	if line == "" || HasSGR(line) {
		return line
	}
	for _, r := range paintRules {
		if r.line.MatchString(line) {
			return Wrap(line, r.code)
		}
	}
	return line
}

// Line prepares one line of tool output for its destination: with colour on,
// only SGR codes stay and smith paint applies; with colour off, every escape
// sequence goes.
//
// Parameters: line is one line of tool output; colour is the destination's
// decision.
// Returns: the line to write.
func Line(line string, colour bool) string {
	if colour {
		return Paint(KeepSGR(line))
	}
	return Strip(line)
}
