package adoption

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/monosmith-app/monosmith/apps/smith/features/release"
	"github.com/monosmith-app/monosmith/apps/smith/features/releaselineage"
	"github.com/monosmith-app/monosmith/apps/smith/features/session"
	"github.com/monosmith-app/monosmith/apps/smith/features/workspace"
)

// IsWorkspace reports whether dir holds a smith.toml that reads: an adopted
// workspace, where smith adopt --tags still makes the baseline tags that
// smith release and smith doctor point at.
func IsWorkspace(dir string) bool {
	raw, err := os.ReadFile(filepath.Join(dir, workspace.FileName))
	if err != nil {
		return false
	}
	_, err = workspace.Parse(raw)
	return err == nil
}

// TagWorkspaceBaselines is smith adopt --tags in a workspace that exists: it
// tags, locally, each commit whose release tag is under an old name (v1.2.3
// for a package now tagged name@1.2.3) with the current name. Nothing is
// pushed.
//
// Parameters: dir is the workspace; version is this smith's; out gets one
// line per tag, or a line that says there is nothing to tag.
// Returns: an error when the workspace cannot be opened or git cannot tag.
func TagWorkspaceBaselines(dir, version string, out io.Writer) error {
	s, err := session.Open(session.Options{Dir: dir, Version: version, Log: io.Discard})
	if err != nil {
		return err
	}
	defer s.Close()
	var baselines []*releaselineage.Baseline
	for _, f := range release.CheckLineage(s, false) {
		baselines = append(baselines, f.Baseline)
	}
	made, err := createTags(s.Root, baselines, out)
	if err != nil {
		return err
	}
	if made == 0 {
		fmt.Fprintln(out, "No release tag to create: each project's tags are under its tag prefix.")
	}
	return nil
}
