package convertplan

import (
	"fmt"
	"regexp"
	"strings"
)

// CommandLine is a shell command line split into what a smith task runs: an
// argv (never passed through a shell) and leading VAR=value assignments.
type CommandLine struct {
	Argv []string
	Env  map[string]string
}

var assignment = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

// SplitCommandLine splits a POSIX-style command line. It refuses anything
// only a shell can run - operators, redirects, substitutions, variable
// expansion and globs - and says which construct it found, so the caller can
// report it instead of producing a task that behaves differently.
func SplitCommandLine(line string) (CommandLine, error) {
	var words []string
	var cur strings.Builder
	inWord := false
	flush := func() {
		if inWord {
			words = append(words, cur.String())
			cur.Reset()
			inWord = false
		}
	}
	rs := []rune(strings.TrimSpace(line))
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			flush()
		case c == '\'':
			inWord = true
			end := indexRune(rs, i+1, '\'')
			if end < 0 {
				return CommandLine{}, fmt.Errorf("unterminated single quote")
			}
			cur.WriteString(string(rs[i+1 : end]))
			i = end
		case c == '"':
			inWord = true
			j := i + 1
			for ; j < len(rs) && rs[j] != '"'; j++ {
				switch rs[j] {
				case '\\':
					if j+1 < len(rs) && strings.ContainsRune("\"\\$`", rs[j+1]) {
						j++
					}
				case '$', '`':
					return CommandLine{}, fmt.Errorf("it expands %q inside double quotes", string(rs[j]))
				}
				cur.WriteRune(rs[j])
			}
			if j >= len(rs) {
				return CommandLine{}, fmt.Errorf("unterminated double quote")
			}
			i = j
		case c == '\\':
			if i+1 < len(rs) {
				i++
				cur.WriteRune(rs[i])
				inWord = true
			}
		case strings.ContainsRune("|&;<>()", c):
			return CommandLine{}, fmt.Errorf("it uses the shell operator %q", operatorAt(rs, i))
		case c == '$' || c == '`':
			return CommandLine{}, fmt.Errorf("it expands %q", string(c))
		case c == '*' || c == '?':
			return CommandLine{}, fmt.Errorf("it relies on the shell to expand the glob %q", string(c))
		default:
			cur.WriteRune(c)
			inWord = true
		}
	}
	flush()
	out := CommandLine{}
	for len(words) > 0 && assignment.MatchString(words[0]) {
		k, v, _ := strings.Cut(words[0], "=")
		if out.Env == nil {
			out.Env = map[string]string{}
		}
		out.Env[k] = v
		words = words[1:]
	}
	if len(words) == 0 {
		return CommandLine{}, fmt.Errorf("it has no command")
	}
	out.Argv = words
	return out, nil
}

func indexRune(rs []rune, from int, r rune) int {
	for i := from; i < len(rs); i++ {
		if rs[i] == r {
			return i
		}
	}
	return -1
}

func operatorAt(rs []rune, i int) string {
	if i+1 < len(rs) && (rs[i+1] == rs[i] || (rs[i] == '>' && rs[i+1] == '&')) {
		return string(rs[i : i+2])
	}
	return string(rs[i])
}
