package cachecommand

import (
	"fmt"
	"sort"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/cachemaintenance"
)

// shownPerList is how many tasks or problems a text report lists before it
// says how many more there are.
const shownPerList = 10

// statusText is what `smith cache status` prints.
func statusText(r StatusReport) string {
	var b strings.Builder
	fmt.Fprintf(&b, "local cache  %s\n", r.Dir)
	row := func(label, value string) { fmt.Fprintf(&b, "  %-11s %s\n", label, value) }
	if !r.Exists {
		row("entries", "none yet: nothing has been cached")
	} else {
		row("entries", fmt.Sprintf("%d", r.Entries))
		row("blobs", fmt.Sprintf("%d", r.Blobs))
	}
	switch {
	case r.LimitBytes <= 0:
		row("size", formatSize(r.Bytes)+" (no limit: [cache] max_size_mb is 0)")
	default:
		percent := 0.0
		if r.UsedPercent != nil {
			percent = *r.UsedPercent
		}
		row("size", fmt.Sprintf("%s of %s (%.1f%% full)", formatSize(r.Bytes), formatSize(r.LimitBytes), percent))
	}
	if r.LastPrune == nil {
		row("last prune", "never")
	} else {
		row("last prune", r.LastPrune.Format("2006-01-02 15:04"))
	}
	row("remote", remoteText(r.Remote))
	if r.OverLimit() {
		fmt.Fprintf(&b, "  ✖ the cache is over its limit\n    → smith cache prune brings it under %s\n", formatSize(r.LimitBytes))
	}
	if r.BrokenEntries > 0 {
		fmt.Fprintf(&b, "  ✖ %d %s cannot be used: unreadable, or naming a blob that is gone\n    → smith cache verify --fix removes them\n", r.BrokenEntries, plural(r.BrokenEntries, "entry", "entries"))
	}
	if len(r.Largest) > 0 {
		b.WriteString("largest entries\n")
		for _, u := range r.Largest {
			fmt.Fprintf(&b, "  %-24s %10s  %d %s\n", u.Task, formatSize(u.LargestBytes), u.Entries, plural(u.Entries, "entry", "entries"))
		}
	}
	return b.String()
}

func remoteText(r RemoteSummary) string {
	if !r.Configured {
		return "none"
	}
	address := r.Scheme + "://" + r.Host
	switch {
	case !r.Active:
		return address + " (off: set SMITH_CACHE_KEY to use it)"
	case r.ReadOnly:
		return address + " (active, read-only)"
	}
	return address + " (active)"
}

// pruneText is what `smith cache prune` prints.
func pruneText(o cachemaintenance.PruneOutcome) string {
	var b strings.Builder
	limit := "no limit"
	if o.LimitBytes > 0 {
		limit = "its limit of " + formatSize(o.LimitBytes)
	}
	if !o.Exists {
		return fmt.Sprintf("The cache at %s is empty: nothing to prune.\n", o.Dir)
	}
	if o.Entries == 0 && o.Blobs == 0 {
		fmt.Fprintf(&b, "Nothing to prune: the cache is %s in %d %s, within %s.\n", formatSize(o.SizeAfter), o.Kept, plural(o.Kept, "entry", "entries"), limit)
	} else {
		verb, left := "Pruned the cache: removed", "left"
		if o.DryRun {
			verb, left = "Would remove", "would be left"
		}
		fmt.Fprintf(&b, "%s %d %s and %d %s (%s of %s); %s %s in %d %s.\n", verb, o.Entries, plural(o.Entries, "entry", "entries"), o.Blobs, plural(o.Blobs, "blob", "blobs"), formatSize(o.Freed), formatSize(o.SizeBefore), formatSize(o.SizeAfter), left, o.Kept, plural(o.Kept, "entry", "entries"))
		for i, line := range taskLines(o.Tasks) {
			if i == shownPerList {
				fmt.Fprintf(&b, "  and %d more\n", len(o.Tasks)-shownPerList)
				break
			}
			b.WriteString(line)
		}
	}
	if o.Skipped > 0 {
		fmt.Fprintf(&b, "! %d %s could not be removed (another process has them open)\n  → run `smith cache prune` again when no build is running\n", o.Skipped, plural(o.Skipped, "file", "files"))
	}
	return b.String()
}

// taskLines list the tasks whose entries go, most first.
func taskLines(tasks map[string]int) []string {
	names := make([]string, 0, len(tasks))
	for name := range tasks {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		if tasks[names[i]] != tasks[names[j]] {
			return tasks[names[i]] > tasks[names[j]]
		}
		return names[i] < names[j]
	})
	lines := make([]string, len(names))
	for i, name := range names {
		label := name
		if label == "" {
			label = "(unreadable entries)"
		}
		lines[i] = fmt.Sprintf("  %-24s %d %s\n", label, tasks[name], plural(tasks[name], "entry", "entries"))
	}
	return lines
}

// verifyText is what `smith cache verify` prints.
func verifyText(r cachemaintenance.VerifyReport) string {
	var b strings.Builder
	if len(r.Problems) == 0 {
		fmt.Fprintf(&b, "Checked %d %s and %d %s in %s: no problems.\n", r.Entries, plural(r.Entries, "entry", "entries"), r.Blobs, plural(r.Blobs, "blob", "blobs"), r.Dir)
		return b.String()
	}
	fmt.Fprintf(&b, "Checked %d %s and %d %s in %s: %d %s.\n", r.Entries, plural(r.Entries, "entry", "entries"), r.Blobs, plural(r.Blobs, "blob", "blobs"), r.Dir, len(r.Problems), plural(len(r.Problems), "problem", "problems"))
	for i, p := range r.Problems {
		if i == shownPerList*3 {
			fmt.Fprintf(&b, "  and %d more\n", len(r.Problems)-i)
			break
		}
		state := ""
		if p.Fixed {
			state = " (removed)"
		}
		fmt.Fprintf(&b, "  %-16s %-14s %s%s\n", p.Kind, shortKey(p.Key), p.Detail, state)
	}
	return b.String()
}

// shortKey shortens a digest for a table.
func shortKey(key string) string {
	if len(key) > 12 {
		return key[:12] + "…"
	}
	return key
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
