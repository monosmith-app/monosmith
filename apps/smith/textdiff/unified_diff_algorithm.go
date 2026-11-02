package textdiff

import (
	"fmt"
	"strings"
)

const contextLines = 3

// UnifiedDiff renders the change from before to after as a unified diff of
// lines, with three lines of context; "" when they are equal.
func UnifiedDiff(path, before, after string) string {
	if before == after {
		return ""
	}
	a, b := lines(before), lines(after)
	// lcs[i][j] is the length of the longest common subsequence of a[i:] and
	// b[j:].
	lcs := make([][]int, len(a)+1)
	for i := range lcs {
		lcs[i] = make([]int, len(b)+1)
	}
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	type op struct {
		kind byte // ' ', '-', '+'
		text string
		ai   int // line index in a (for ' ' and '-')
		bi   int // line index in b (for ' ' and '+')
	}
	var ops []op
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		switch {
		case i < len(a) && j < len(b) && a[i] == b[j]:
			ops = append(ops, op{' ', a[i], i, j})
			i, j = i+1, j+1
		case i < len(a) && (j == len(b) || lcs[i+1][j] >= lcs[i][j+1]):
			ops = append(ops, op{'-', a[i], i, j})
			i++
		default:
			ops = append(ops, op{'+', b[j], i, j})
			j++
		}
	}
	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n+++ b/%s\n", path, path)
	for k := 0; k < len(ops); {
		if ops[k].kind == ' ' {
			k++
			continue
		}
		start := max(k-contextLines, 0)
		end := k
		for end < len(ops) {
			if ops[end].kind != ' ' {
				end++
				continue
			}
			run := end
			for run < len(ops) && ops[run].kind == ' ' {
				run++
			}
			if run-end > 2*contextLines || run == len(ops) {
				end = min(end+contextLines, len(ops))
				break
			}
			end = run
		}
		aStart, bStart, aLen, bLen := ops[start].ai, ops[start].bi, 0, 0
		var body strings.Builder
		for _, o := range ops[start:end] {
			body.WriteString(string(o.kind) + o.text + "\n")
			if o.kind != '+' {
				aLen++
			}
			if o.kind != '-' {
				bLen++
			}
		}
		fmt.Fprintf(&out, "@@ -%d,%d +%d,%d @@\n%s", aStart+1, aLen, bStart+1, bLen, body.String())
		k = end
	}
	return out.String()
}

func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}
