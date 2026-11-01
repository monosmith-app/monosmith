package pluginusage

import (
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/monosmith-app/monosmith/apps/smith/session"
)

// Projects names, for each plugin, the projects it discovers.
//
// Parameters: dir is a folder inside the workspace; version is the running
// smith's version; names are the plugins to count.
// Returns: the projects per plugin name (a plugin that finds none is absent),
// or the error that stopped the workspace from loading.
func Projects(dir, version string, names []string) (map[string][]string, error) {
	s, err := session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
	if err != nil {
		return nil, err
	}
	defer s.Close()
	out := map[string][]string{}
	for _, project := range s.Graph.Names() {
		if plugin := s.Graph.Get(project).Plugin; slices.Contains(names, plugin) {
			out[plugin] = append(out[plugin], project)
		}
	}
	return out, nil
}

// Warning says how many projects switching off the plugins hides, naming the
// projects (up to five each). It returns "" when no project uses them.
//
// Parameters: used is what Projects returned; names are the plugins in the
// order the person gave them.
func Warning(used map[string][]string, names []string) string {
	var parts []string
	for _, name := range names {
		projects := used[name]
		if len(projects) == 0 {
			continue
		}
		list := strings.Join(projects[:min(len(projects), 5)], ", ")
		if len(projects) > 5 {
			list += ", and " + strconv.Itoa(len(projects)-5) + " more"
		}
		noun := "projects"
		if len(projects) == 1 {
			noun = "project"
		}
		parts = append(parts, name+" finds "+strconv.Itoa(len(projects))+" "+noun+" ("+list+")")
	}
	return strings.Join(parts, "; ")
}
