package projectstandard

import (
	"fmt"
	"io"

	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/packages/plugin-sdk/protocol"
)

// Differences asks every plugin of the open workspace for the standard of its
// projects' files and returns those that differ, plugin by plugin.
//
// Parameters: s is the open session.
// Returns: the differences; a plugin that fails to answer is named in the
// error and the others still count.
func Differences(s *session.Session) ([]Difference, error) {
	var out []Difference
	var failed error
	for _, c := range s.Plugins.Clients {
		var own []protocol.Project
		for _, p := range s.Graph.Projects {
			if p.Plugin == c.Info.Plugin.Name {
				own = append(own, p.Project)
			}
		}
		a, err := Ask(s.Root, c, own)
		if err != nil {
			failed = fmt.Errorf("plugin %s did not give its standard files: %w", c.Info.Plugin.Name, err)
			continue
		}
		out = append(out, a.Differences...)
	}
	return out, failed
}

// Standardize is the step of smith adopt --init and smith convert that takes
// the answer to "your project configuration differs from smith's standard":
// with Replace it writes the standard files, keeping each whole file it
// replaces as <file>.pre-smith; with Keep it writes nothing and says that
// smith doctor lists the differences.
//
// Parameters: s is the session of the converted workspace; answer is Replace
// or Keep; out gets the lines.
// Returns: the workspace-relative paths written (copies included) and the
// steps left to do, in words.
// Errors: a file cannot be written.
func Standardize(s *session.Session, answer string, out io.Writer) (written, left []string, err error) {
	diffs, askErr := Differences(s)
	if askErr != nil {
		left = append(left, askErr.Error()+": run smith doctor once it answers")
	}
	if len(diffs) == 0 {
		return nil, left, nil
	}
	if answer == Keep {
		fmt.Fprintf(out, "  standard: %d file(s) differ from smith's standard and stay as they are (--project-config keep); smith doctor lists them\n", len(diffs))
		return nil, left, nil
	}
	written, err = Apply(s.Root, diffs, out)
	return written, left, err
}
