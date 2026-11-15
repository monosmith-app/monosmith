package releaseassets

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// uploadFiles attaches files to the release of tag in repo (owner/repo) with
// gh release upload --clobber, so a re-run replaces a file instead of failing.
// It is a variable so tests do not need gh.
var uploadFiles = func(ctx context.Context, dir, repo, tag string, files []string, out io.Writer) error {
	args := append([]string{"release", "upload", tag}, files...)
	args = append(args, "--clobber", "--repo", repo)
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir, cmd.Stdout, cmd.Stderr = dir, out, out
	cmd.Env = os.Environ()
	if os.Getenv("GH_TOKEN") == "" && os.Getenv("GITHUB_TOKEN") != "" {
		cmd.Env = append(cmd.Env, "GH_TOKEN="+os.Getenv("GITHUB_TOKEN"))
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh release upload %s: %w (is gh installed and signed in, and does the release exist?)", tag, err)
	}
	return nil
}
