package languageserver

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/pelletier/go-toml/v2"

	"github.com/monosmith-app/monosmith/apps/smith/workspace"
)

// diagnosticSource names smith in the editor's problem list.
const diagnosticSource = "smith"

// Diagnose checks a smith.toml text the way `smith` loads it: TOML syntax, a
// key smith does not know, a value of the wrong type, and the values smith
// rejects. A problem with a position is underlined where it is; one without
// (a missing [workspace] name) is attached to the first line.
func Diagnose(text string) []Diagnostic {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	var cfg workspace.Config
	err := toml.NewDecoder(strings.NewReader(text)).DisallowUnknownFields().Decode(&cfg)
	var strict *toml.StrictMissingError
	var decode *toml.DecodeError
	switch {
	case err == nil:
		if verr := workspace.Validate(&cfg); verr != nil {
			return []Diagnostic{wholeLine(lines, 0, trimFilePrefix(verr.Error()))}
		}
		return nil
	case errors.As(err, &strict):
		out := make([]Diagnostic, 0, len(strict.Errors))
		for i := range strict.Errors {
			e := &strict.Errors[i]
			row, col := e.Position()
			out = append(out, Diagnostic{
				Range:    keyRange(lines, row, col),
				Severity: SeverityError,
				Source:   diagnosticSource,
				Message:  "unknown key " + strings.Join(e.Key(), "."),
			})
		}
		return out
	case errors.As(err, &decode):
		row, col := decode.Position()
		return []Diagnostic{{
			Range:    restOfLine(lines, row, col),
			Severity: SeverityError,
			Source:   diagnosticSource,
			Message:  decode.Error(),
		}}
	}
	return []Diagnostic{wholeLine(lines, 0, err.Error())}
}

// trimFilePrefix drops the "smith.toml: " that smith's own messages start with.
func trimFilePrefix(msg string) string {
	return strings.TrimPrefix(msg, workspace.FileName+": ")
}

// runeColumn converts a 1-based byte column of a line to a 0-based rune column.
func runeColumn(line string, byteCol int) int {
	idx := max(0, min(byteCol-1, len(line)))
	return utf8.RuneCountInString(line[:idx])
}

// lineAt returns line row (1-based), or "" outside the text.
func lineAt(lines []string, row int) string {
	if row < 1 || row > len(lines) {
		return ""
	}
	return lines[row-1]
}

// keyRange spans the bare key that starts at the 1-based row and byte column.
func keyRange(lines []string, row, col int) Range {
	line := lineAt(lines, row)
	start := runeColumn(line, col)
	end := start
	runes := []rune(line)
	for end < len(runes) && (isKeyChar(runes[end]) || runes[end] == '.') {
		end++
	}
	if end == start {
		end = start + 1
	}
	return Range{Start: Position{Line: max(row-1, 0), Character: start}, End: Position{Line: max(row-1, 0), Character: end}}
}

// restOfLine spans from the 1-based row and byte column to the end of the line.
func restOfLine(lines []string, row, col int) Range {
	line := lineAt(lines, row)
	start := runeColumn(line, col)
	end := max(utf8.RuneCountInString(strings.TrimRight(line, " \t")), start+1)
	return Range{Start: Position{Line: max(row-1, 0), Character: start}, End: Position{Line: max(row-1, 0), Character: end}}
}

// wholeLine builds a diagnostic that spans the 0-based line.
func wholeLine(lines []string, line int, msg string) Diagnostic {
	end := 0
	if line < len(lines) {
		end = utf8.RuneCountInString(strings.TrimRight(lines[line], "\r"))
	}
	return Diagnostic{
		Range:    Range{Start: Position{Line: line}, End: Position{Line: line, Character: end}},
		Severity: SeverityError,
		Source:   diagnosticSource,
		Message:  msg,
	}
}
