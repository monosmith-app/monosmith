package cachecommand

import (
	"encoding/json"
	"io"

	"github.com/spf13/cobra"
)

// Command is `smith cache`: status, prune, clear and verify of the local task
// cache.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Look after the local task cache: status, prune, clear, verify",
		Long: `Look after the local task cache (.smith/cache, or [cache] dir, or SMITH_CACHE_DIR).

smith keeps the cache under [cache] max_size_mb by itself, after runs that
stored results. These commands show what is in it, trim it now, delete it, or
check it for damage. They act on this machine's cache and never on the remote
one.`,
		Example: `  smith cache status
  smith cache prune --dry-run
  smith cache verify --fix`,
	}
	cmd.AddCommand(statusCommand(), pruneCommand(), clearCommand(), verifyCommand())
	return cmd
}

func encode(out io.Writer, v any) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
