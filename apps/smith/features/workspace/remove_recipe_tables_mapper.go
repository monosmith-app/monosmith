package workspace

import "strings"

// RemoveRecipeTables deletes from smith.toml text the tables a recipe wrote
// for the given projects ([projects.<name>.recipe], its answers and
// [projects.<name>.settings] with every plugin's table) and the given stacks
// ([stacks.<name>]), keeping every other line exactly, so the caller can
// write them again. The comment `smith add` put above a removed table goes
// with it. It reports whether anything was removed.
func RemoveRecipeTables(src string, projects, stacks []string) (string, bool) {
	return removeTables(src, func(key string) bool {
		for _, name := range projects {
			if recipeTable(key, name) {
				return true
			}
		}
		for _, name := range stacks {
			if stackTable(key, name) {
				return true
			}
		}
		return false
	})
}

// removeTables deletes every table whose key match accepts, with the tables
// nested under it only if match accepts them too.
func removeTables(src string, match func(key string) bool) (string, bool) {
	lines := strings.SplitAfter(src, "\n")
	var out []string
	skipping, removed := false, false
	for _, line := range lines {
		if m := tableHeader.FindStringSubmatch(strings.TrimRight(line, "\r\n")); m != nil {
			starting := match(m[1]) && !skipping
			skipping = match(m[1])
			removed = removed || skipping
			// The comment smith wrote above a table goes with it.
			if starting && len(out) > 0 && recipeCommentOf(out[len(out)-1]) != "" {
				out = out[:len(out)-1]
			}
		} else if skipping && recipeCommentOf(line) != "" {
			// The comment above the next table sits after the end of this one.
			out = append(out, line)
			continue
		}
		if !skipping {
			out = append(out, line)
		}
	}
	text := strings.Join(out, "")
	for strings.Contains(text, "\n\n\n") {
		text = strings.ReplaceAll(text, "\n\n\n", "\n\n")
	}
	if removed {
		// What was removed leaves the blank line that separated it behind.
		if text = strings.TrimRight(text, "\r\n"); text != "" {
			text += "\n"
		}
	}
	return text, removed
}

// recipeTable reports whether a table key is a project's recipe record or its
// settings, or a table nested in either.
func recipeTable(key, project string) bool {
	rest, ok := strings.CutPrefix(key, "projects.")
	if !ok {
		return false
	}
	for _, form := range []string{project, `"` + project + `"`} {
		below, ok := strings.CutPrefix(rest, form+".")
		if !ok {
			continue
		}
		for _, table := range []string{"recipe", "settings"} {
			if below == table || strings.HasPrefix(below, table+".") {
				return true
			}
		}
	}
	return false
}

// stackTable reports whether a table key is [stacks.<name>].
func stackTable(key, name string) bool {
	rest, ok := strings.CutPrefix(key, "stacks.")
	return ok && (rest == name || rest == `"`+name+`"`)
}
