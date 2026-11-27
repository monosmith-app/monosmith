package palette

// row is one entry of the list: a heading (index -1) with the count of the
// entries under it, or the entry at index of the ordered matches.
type row struct {
	heading string
	count   int
	index   int
}

// headings names the groups the list makes besides the sections of the
// commands: the first match of a search, and the commands that cannot be used now.
type headings struct {
	best, unavailable string
}

// groupRows puts the matches in the order they are drawn and gives the rows
// that draw them. A search (typed) puts its first usable match under the Best
// heading. The other usable entries follow under their sections, in the order
// a section first shows, and the entries that cannot be used now come last
// under the Unavailable heading. Every heading counts its entries.
//
// Parameters: found are the matches, best first; typed is true while a search
// is typed; names are the headings of the two extra groups.
// Returns: the entries in the drawn order (the cursor counts in it) and the rows.
func groupRows(found []Entry, typed bool, names headings) ([]Entry, []row) {
	var out []Entry
	var rows []row
	add := func(heading string, entries []Entry) {
		if len(entries) == 0 {
			return
		}
		if heading != "" {
			rows = append(rows, row{heading: heading, count: len(entries), index: -1})
		}
		for _, e := range entries {
			rows = append(rows, row{index: len(out)})
			out = append(out, e)
		}
	}
	rest := found
	if typed && len(found) > 0 && !found[0].unavailable() {
		add(names.best, found[:1])
		rest = found[1:]
	}
	var order []string
	bySection := map[string][]Entry{}
	var blocked []Entry
	for _, e := range rest {
		if e.unavailable() {
			blocked = append(blocked, e)
			continue
		}
		if _, seen := bySection[e.Section]; !seen {
			order = append(order, e.Section)
		}
		bySection[e.Section] = append(bySection[e.Section], e)
	}
	for _, s := range order {
		add(s, bySection[s])
	}
	add(names.unavailable, blocked)
	return out, rows
}

// flatRows are the rows of matches drawn without headings, in their order.
func flatRows(found []Entry) []row {
	rows := make([]row, len(found))
	for i := range found {
		rows[i] = row{index: i}
	}
	return rows
}
