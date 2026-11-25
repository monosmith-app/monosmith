package projectstandard

import (
	"fmt"
	"time"

	"github.com/monosmith-app/monosmith/packages/plugin-sdk/host"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Answer is what one plugin said about its projects' files.
type Answer struct {
	// Projects are the projects the plugin was asked about.
	Projects []protocol.Project
	// Standard lists, by project name ("" for the workspace), the files the
	// plugin has a standard for, whether they differ or not.
	Standard map[string][]string
	// Differences are the files that are not as the standard has them, in the
	// plugin's order.
	Differences []Difference
}

// Ask asks one plugin for the standard of its projects' files and compares
// them with the files on disk. A plugin without the standard capability, or
// with no projects, gives nothing.
//
// Parameters: root is the workspace; c is the plugin; projects are the
// projects it serves.
// Returns: the answer.
// Errors: the call fails, or a standard file does not fit its mode or the
// file on disk does not parse.
func Ask(root string, c *host.Client, projects []protocol.Project) (Answer, error) {
	a := Answer{Projects: projects, Standard: map[string][]string{}}
	if len(projects) == 0 || !c.Supports(protocol.MethodStandard) {
		return a, nil
	}
	var res protocol.StandardResult
	if err := c.Call(protocol.MethodStandard, protocol.StandardParams{Projects: projects}, &res, time.Minute); err != nil {
		return a, err
	}
	return Collect(root, projects, res)
}

// Collect compares a plugin's standard files with the files on disk.
//
// Parameters: root is the workspace; projects are the projects asked about;
// res is the plugin's answer.
// Returns: the answer; a file of a project not asked about is left out.
// Errors: a standard file does not fit its mode, or the file on disk does not
// parse in its format.
func Collect(root string, projects []protocol.Project, res protocol.StandardResult) (Answer, error) {
	a := Answer{Projects: projects, Standard: map[string][]string{}}
	asked := map[string]bool{"": true}
	for _, p := range projects {
		asked[p.Name] = true
	}
	for _, f := range res.Files {
		if !asked[f.Project] {
			continue
		}
		a.Standard[f.Project] = append(a.Standard[f.Project], f.Path)
		d, differs, err := Compare(root, f)
		if err != nil {
			return a, fmt.Errorf("the standard of %s: %w", f.Path, err)
		}
		if differs {
			a.Differences = append(a.Differences, d)
		}
	}
	return a, nil
}
