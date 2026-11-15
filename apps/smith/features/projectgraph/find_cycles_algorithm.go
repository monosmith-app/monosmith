package projectgraph

import (
	"sort"
	"strings"
)

// Cycles lists every import cycle, each as a path that ends where it starts
// (web → ui → web). Projects that import each other through several paths
// form one strongly connected group; its cycle starts at the group's first
// name and follows the first import that stays in the group.
func (g *Graph) Cycles() [][]string {
	index, low := map[string]int{}, map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	var groups [][]string
	next := 0
	var connect func(string)
	connect = func(v string) {
		index[v], low[v] = next, next
		next++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range g.byName[v].Deps {
			if g.byName[w] == nil {
				continue
			}
			if _, seen := index[w]; !seen {
				connect(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] != index[v] {
			return
		}
		var group []string
		for {
			w := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[w] = false
			group = append(group, w)
			if w == v {
				break
			}
		}
		groups = append(groups, group)
	}
	names := g.Names()
	for _, n := range names {
		if _, seen := index[n]; !seen {
			connect(n)
		}
	}
	var out [][]string
	for _, group := range groups {
		in := map[string]bool{}
		for _, n := range group {
			in[n] = true
		}
		sort.Strings(group)
		start := group[0]
		selfLoop := false
		for _, d := range g.byName[start].Deps {
			selfLoop = selfLoop || d == start
		}
		if len(group) == 1 && !selfLoop {
			continue
		}
		out = append(out, g.cycleThrough(start, in))
	}
	sort.Slice(out, func(i, j int) bool { return strings.Join(out[i], ",") < strings.Join(out[j], ",") })
	return out
}

// cycleThrough finds a path from start back to start inside the group.
func (g *Graph) cycleThrough(start string, in map[string]bool) []string {
	visited := map[string]bool{}
	var path []string
	var walk func(string) bool
	walk = func(v string) bool {
		path = append(path, v)
		deps := append([]string(nil), g.byName[v].Deps...)
		sort.Strings(deps)
		for _, d := range deps {
			if d == start {
				path = append(path, start)
				return true
			}
			if in[d] && !visited[d] {
				visited[d] = true
				if walk(d) {
					return true
				}
			}
		}
		path = path[:len(path)-1]
		return false
	}
	visited[start] = true
	walk(start)
	return path
}

// CycleRemedy names the import to remove to break a cycle (a path that ends
// where it starts): the last edge, the one that closes the loop.
func CycleRemedy(cycle []string) string {
	if len(cycle) < 2 {
		return ""
	}
	from, to := cycle[len(cycle)-2], cycle[len(cycle)-1]
	return "remove " + from + "'s import of " + to + " (" + from + " → " + to + ")"
}
