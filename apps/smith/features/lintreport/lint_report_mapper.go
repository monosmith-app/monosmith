package lintreport

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Item is one thing the lint of a project left to do, with its first step.
type Item struct {
	// Problem says what the lint found, in one line.
	Problem string
	// Remedy is the step that clears it.
	Remedy string
}

var (
	eslintTotal   = regexp.MustCompile(`✖ (\d+) problems? \(`)
	eslintFixable = regexp.MustCompile(`(\d+) errors? and (\d+) warnings? potentially fixable`)
	eslintParse   = regexp.MustCompile(`^\s*\d+:\d+\s+error\s+Parsing error: `)
	eslintRule    = regexp.MustCompile(`^\s*\d+:\d+\s+(?:error|warning)\s+.*\S\s{2,}(\S+)\s*$`)
	ruffTotal     = regexp.MustCompile(`^Found (\d+) errors?\.`)
	ruffFixable   = regexp.MustCompile(`^\[\*\] (\d+) fixable`)
	ruffRule      = regexp.MustCompile(`(?:^|:\d+:\d+: )([A-Z]{1,4}[0-9]{3,4}) `)
	sliceTotal    = regexp.MustCompile(`(\d+) file\(s\) or folder\(s\) break the vertical-slice layout`)
)

// shownRules is how many rules a lint item names.
const shownRules = 3

// Summarize reads the output of `smith run lint -p <name>` that failed and
// says, for each task that failed, what it found and the step that clears
// it: the count of ESLint or ruff problems with the ones --fix repairs and
// the rules found most, the files that break the vertical-slice layout with
// the three ways on, or the task that stopped the lint before it ran.
//
// Parameters: name is the project; output is what smith printed.
// Returns: the items, one per failed task; one item that names the command
// when the output names no failed task.
func Summarize(name, output string) []Item {
	logs, failed := split(name, output)
	var items []Item
	for _, task := range failed {
		lines := logs[task]
		switch task {
		case "slice-check":
			items = append(items, sliceItem(name, lines))
		case "lint":
			items = append(items, lintItem(name, lines))
		default:
			items = append(items, Item{
				Problem: fmt.Sprintf("the lint of %s did not run: its task %s fails (%s)", name, task, cause(lines)),
				Remedy:  fmt.Sprintf("smith run lint -p %s shows the whole output; fix the cause it names", name),
			})
		}
	}
	if len(items) == 0 {
		items = append(items, Item{
			Problem: "the lint of " + name + " fails",
			Remedy:  "smith run lint -p " + name + " shows why",
		})
	}
	return items
}

// split sorts the output lines by task ("<name>:<task> │ text") and lists
// the tasks that failed, in the order smith reported them.
func split(name, output string) (map[string][]string, []string) {
	logs := map[string][]string{}
	var failed []string
	prefix := name + ":"
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		rest := line[len(prefix):]
		task, text, ok := strings.Cut(rest, " │")
		if ok {
			logs[task] = append(logs[task], strings.TrimPrefix(text, " "))
			continue
		}
		if task, _, ok := strings.Cut(rest, " × "); ok && !contains(failed, task) {
			failed = append(failed, task)
		}
	}
	return logs, failed
}

// lintItem is the item of a linter that ran and found problems.
func lintItem(name string, lines []string) Item {
	total, fixable := 0, 0
	rules := map[string]int{}
	for _, line := range lines {
		if m := eslintTotal.FindStringSubmatch(line); m != nil {
			total = atoi(m[1])
		}
		if m := eslintFixable.FindStringSubmatch(line); m != nil {
			fixable = atoi(m[1]) + atoi(m[2])
		}
		if m := ruffTotal.FindStringSubmatch(line); m != nil {
			total = atoi(m[1])
		}
		if m := ruffFixable.FindStringSubmatch(line); m != nil {
			fixable = atoi(m[1])
		}
		if eslintParse.MatchString(line) {
			// A file ESLint cannot parse has no rule: the cause is in the
			// file or its tsconfig.
			rules["parsing error"]++
		} else if m := eslintRule.FindStringSubmatch(line); m != nil {
			rules[m[1]]++
		} else if m := ruffRule.FindStringSubmatch(line); m != nil {
			rules[m[1]]++
		}
	}
	if total == 0 {
		return Item{
			Problem: fmt.Sprintf("the lint of %s fails (%s)", name, cause(lines)),
			Remedy:  fmt.Sprintf("smith run lint -p %s shows the whole output", name),
		}
	}
	problem := fmt.Sprintf("the lint of %s finds %d problem(s) in its code", name, total)
	if top := topRules(rules); top != "" {
		problem += ", most from " + top
	}
	remedy := fmt.Sprintf("smith lint -p %s lists them; fix them, or turn a rule off in the lint config", name)
	if fixable > 0 {
		remedy = fmt.Sprintf("smith lint --fix -p %s repairs %d of them (it changes your files, so run it when you want that); smith lint -p %s lists the rest", name, fixable, name)
	}
	return Item{Problem: problem, Remedy: remedy}
}

// sliceItem is the item of the vertical-slice check.
func sliceItem(name string, lines []string) Item {
	count, first := 0, ""
	for _, line := range lines {
		if m := sliceTotal.FindStringSubmatch(line); m != nil {
			count = atoi(m[1])
			continue
		}
		if first == "" && line != "" && !strings.HasPrefix(line, "smith:") && !strings.HasPrefix(line, "review:") {
			first = line
		}
	}
	problem := fmt.Sprintf("the lint of %s stops at the vertical-slice check", name)
	if count > 0 {
		problem = fmt.Sprintf("the lint of %s stops at the vertical-slice check: %d file(s) or folder(s) break the layout", name, count)
	}
	if first != "" {
		problem += " (first: " + first + ")"
	}
	return Item{
		Problem: problem,
		Remedy: "move the files into slices (smith run slice-check -p " + name + " lists each one with its rule)\n" +
			"or write the exception in the project README\n" +
			"or turn the check off for the project: add \"" + name + "\" to exclude under [conventions.vertical-slices] in smith.toml",
	}
}

// cause is the line of a task's output that says why it failed: smith's own
// last line, else the first line that names an error, else the first line
// with text.
func cause(lines []string) string {
	for i := len(lines) - 1; i >= 0; i-- {
		if text, ok := strings.CutPrefix(lines[i], "smith: "); ok {
			return text
		}
	}
	first := ""
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if strings.Contains(strings.ToLower(text), "error") && !strings.HasPrefix(text, "npm error") {
			return text
		}
		if first == "" {
			first = text
		}
	}
	if first != "" {
		return first
	}
	return "it names no cause"
}

// topRules names the rules found most, with their counts.
func topRules(rules map[string]int) string {
	names := make([]string, 0, len(rules))
	for r := range rules {
		names = append(names, r)
	}
	sort.Slice(names, func(i, j int) bool {
		if rules[names[i]] != rules[names[j]] {
			return rules[names[i]] > rules[names[j]]
		}
		return names[i] < names[j]
	})
	if len(names) > shownRules {
		names = names[:shownRules]
	}
	parts := make([]string, len(names))
	for i, r := range names {
		parts[i] = fmt.Sprintf("%s (%d)", r, rules[r])
	}
	return strings.Join(parts, ", ")
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
